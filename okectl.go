package main

// import libraries..
import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/containerengine"
	"github.com/oracle/oci-go-sdk/example/helpers"
)

// variables..
var (
	// general..
	cleanUp                 = false        // func configureFileSystem
	nodeLifeCycleState      = ("nil")      // func getNodeLifeCycleState
	// app..
	app                     = kingpin.New("okectl", "A command-line application for configuring Oracle OKE (Container Engine for Kubernetes.)")
	configDir               = app.Flag("configDir", "Path where output files are created or referenced - e.g. kubeconfig file. Specify as absolute path.").Default(".okectl").String()
	// (c1) :: create cluster..
	c1                      = app.Command("createOkeCluster", "Create new OKE Kubernetes cluster.")
	c1VcnId                 = c1.Flag("vcnId", "OCI VCN Id where cluster will be created.").Required().String()
	c1CompartmentId         = c1.Flag("compartmentId", "OCI Compartment-Id where cluster will be created.").Required().String()
	c1Subnet1Id             = c1.Flag("subnet1Id", "Cluster Control Plane LB Subnet 1.").Required().String()
	c1Subnet2Id             = c1.Flag("subnet2Id", "Cluster Control Plane LB Subnet 2.").Required().String()
	c1Subnet3Id             = c1.Flag("subnet3Id", "Worker Node Subnet 1.").Required().String()
	c1Subnet4Id             = c1.Flag("subnet4Id", "Worker Node Subnet 2.").String()
	c1Subnet5Id             = c1.Flag("subnet5Id", "Worker Node Subnet 3.").String()
	c1ClusterName           = c1.Flag("clusterName", "Kubernetes cluster name.").Default("dev-oke-001").String()
	c1KubeVersion           = c1.Flag("kubeVersion", "Kubernetes cluster version.").Default("v1.10.3").String()
	c1NodeImageName         = c1.Flag("nodeImageName", "OS image used for Worker Node(s).").Default("Oracle-Linux-7.4").String()
	c1NodeShape             = c1.Flag("nodeShape", "CPU/RAM allocated to Worker Node(s).").Default("VM.Standard1.1").String()
	c1NodeSshKey            = c1.Flag("nodeSshKey", "SSH key to provision to Worker Node(s) for remote access.").String()
	c1QuantityWkrSubnets    = c1.Flag("quantityWkrSubnets", "Number of subnets used to host Worker Node(s).").Default("1").Int()
	c1QuantityPerSubnet     = c1.Flag("quantityPerSubnet", "Number of Worker Nodes per subnet.").Default("1").Int()
	c1WaitNodesActive       = c1.Flag("waitNodesActive", "If waitNodesActive=all, wait & return when all nodes in the pool are active. " +
	                                  "If waitNodesActive=any, wait & return when any of the nodes in the pool are active. " +
	                                  "If waitNodesActive=false, no wait & return when the node pool is active.").Default("false").String()
	// (d1) :: delete cluster..
	d1                      = app.Command("deleteOkeCluster", "Delete OKE Kubernetes cluster.")
	d1ClusterId             = d1.Flag("clusterId", "OKE Kubernetes cluster Id. If not specified, clusterId contained in nodepool.json will be used.").String()
	// (c2) :: create kubeconfig.. //update to read clusterId from file..
	c2                      = app.Command("createOkeKubeconfig", "Create kubeconfig autentication artefact for kubectl.")
	c2ClusterId             = c2.Flag("clusterId", "OKE Kubernetes cluster ID. If not specified, clusterId contained in nodepool.json will be used.").String()
	// (c3) :: create nodepool..
	// (g3) :: get nodepool..
	g3                      = app.Command("getOkeNodePool", "Get cluster, node pool, and node details for a specified node pool.")
	g3NodePoolId            = g3.Flag("nodePoolId", "OKE Node Pool Id. If not specified, Id contained in nodepool.json will be used.").String()
	g3TfExternalDs          = g3.Flag("tfExternalDs", "Run as a Terraform external data source, & provide json only response data for Terraform.").Default("false").String()
	g3WaitNodesActive       = g3.Flag("waitNodesActive", "If waitNodesActive=all, wait & return when all nodes in the pool are active. " +
	                                  "If waitNodesActive=any, wait & return when any of the nodes in the pool are active. " +
	                                  "If waitNodesActive=false, no wait & return when the node pool is active.").Default("false").String()
	// (d3) :: delete nodepool..
)

// oke crud..
func main() {
	ctx := context.Background()
	c, clerr := containerengine.NewContainerEngineClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(clerr)

	// command-line args & flags..
	app.Version("0.0.3")
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	// create cluster..
	case c1.FullCommand():

		fmt.Println("")
		fmt.Println("OKECTL :: Create Cluster :: Request Parameters ...")
		fmt.Println("-------------------------------------------------------")
		fmt.Println("configDir:", *configDir)
		fmt.Println("clusterName:", *c1ClusterName)
		fmt.Println("kubeVersion:", *c1KubeVersion)
		fmt.Println("vcnId:", *c1VcnId)
		fmt.Println("compartmentId:", *c1CompartmentId)
		fmt.Println("subnet1Id:", *c1Subnet1Id)
		fmt.Println("subnet2Id:", *c1Subnet2Id)
		fmt.Println("subnet3Id:", *c1Subnet3Id)
		fmt.Println("subnet4Id:", *c1Subnet4Id)
		fmt.Println("subnet5Id:", *c1Subnet5Id)
		fmt.Println("nodeImageName:", *c1NodeImageName)
		fmt.Println("nodeShape:", *c1NodeShape)
		fmt.Println("nodeSshKey:", *c1NodeSshKey)
		fmt.Println("quantityWkrSubnets:", *c1QuantityWkrSubnets)
		fmt.Println("quantityPerSubnet:", *c1QuantityPerSubnet)
		fmt.Println("waitNodesActive:", *c1WaitNodesActive)
		fmt.Println("")

		// brief pause..
		time.Sleep(5 * time.Second)

		// configure file system..
		cleanUp = true
		configDirPath := configureFileSystem(*configDir, cleanUp)

		// create cluster..
		createClusterResp := createCluster(ctx, c, *c1ClusterName, *c1VcnId, *c1CompartmentId, *c1KubeVersion, *c1Subnet1Id, *c1Subnet2Id)

		// wait for create cluster completion..
		workReqRespCls := waitUntilWorkRequestComplete(c, createClusterResp.OpcWorkRequestId)
		fmt.Println("OKECTL :: Create Cluster :: Complete ...")
		clusterId := getResourceID(workReqRespCls.Resources, containerengine.WorkRequestResourceActionTypeCreated, "CLUSTER")

		// create nodepool..
		createNodePoolResp := createNodePool(ctx, c, *c1CompartmentId, *c1ClusterName, *clusterId, *c1KubeVersion, *c1NodeImageName, *c1NodeShape, *c1NodeSshKey, *c1Subnet3Id, *c1Subnet4Id, *c1Subnet5Id, *c1QuantityWkrSubnets, *c1QuantityPerSubnet)

		// wait for create nodepool completion..
		workReqRespNpl := waitUntilWorkRequestComplete(c, createNodePoolResp.OpcWorkRequestId)
		fmt.Println("OKECTL :: Create NodePool :: Complete ...")
		nodePoolId := getResourceID(workReqRespNpl.Resources, containerengine.WorkRequestResourceActionTypeCreated, "NODEPOOL")

		// wait for create node completion..
		if *c1WaitNodesActive != "false" {
			nodeProvisioning := true
			// This loop continues while "nodeProvisioning" is true..
			for nodeProvisioning {
				getNodeLifeCycleState(ctx, c, *nodePoolId)
				// if we are waiting for all nodes to be active..
				if *c1WaitNodesActive == "all" {
					boolLcs := (strings.Contains(nodeLifeCycleState, "ATING"))
						if nodeLifeCycleState != "{}" {
							if boolLcs != true {
								nodeProvisioning = false
							}
						}
				}
				// if we are waiting for any nodes to be active..
				if *c1WaitNodesActive == "any" {
					boolLcs := (strings.Contains(nodeLifeCycleState, "ACTIVE"))
						if nodeLifeCycleState != "{}" {
							if boolLcs == true {
								nodeProvisioning = false
							}
						}
				}
					time.Sleep(15 * time.Second)
			}
		}
		fmt.Println("OKECTL :: Create Node(s) :: Complete ...")

		// get nodepool details & create nodepool.json..
		getNodePool(ctx, c, *nodePoolId, configDirPath)

		// create kubeconfig file..
		getKubeConfig(ctx, c, *clusterId, configDirPath)

		// done, output config data..
		fmt.Println("")
		fmt.Println("OKECTL :: Create Cluster :: Complete ...")
		fmt.Println("-------------------------------------------------------")
		configFilePath := configDirPath + string(os.PathSeparator) + "nodepool.json"
			content, err := ioutil.ReadFile(configFilePath)
			if err != nil {
				fmt.Println(err)
			}
			strNodePool := string(content)
			fmt.Println(strNodePool)

	// delete cluster..
	case d1.FullCommand():
		var clusterId(string)

		// no --clusterId flag provided, reading nodepool.json..
		if *d1ClusterId == "" {
			// configure file system..
			cleanUp = false
			configDirPath := configureFileSystem(*configDir, cleanUp)

			// read nodepool.json..
			configFilePath := configDirPath + string(os.PathSeparator) + "nodepool.json"
			content, err := ioutil.ReadFile(configFilePath)
			if err != nil {
				fmt.Println("OKECTL :: No --clusterId flag provided, error reading nodepool.json at specified path :: Exiting..")
				fmt.Println(err)
				os.Exit(3)
			}

			// get clusterId from nodepool.json..
			jsonParsed, err := gabs.ParseJSON(content)
			clusterId = (jsonParsed.Path("clusterId").String())
			*d1ClusterId = clusterId[1 : len(clusterId)-1]
		}

		fmt.Println("")
		fmt.Println("OKECTL :: Delete Cluster :: Request Parameters ...")
		fmt.Println("-------------------------------------------------------")
		fmt.Println("clusterId:", *d1ClusterId)
		fmt.Println("")

		// brief pause..
		time.Sleep(5 * time.Second)

		// delete cluster..
		deleteClusterResp := deleteCluster(ctx, c, *d1ClusterId)

		// wait for delete cluster completion..
		waitUntilWorkRequestComplete(c, deleteClusterResp.OpcWorkRequestId)

		// done..
		fmt.Println("")
		fmt.Println("OKECTL :: Delete Cluster :: Complete ...")

	// create kubeconfig..
	case c2.FullCommand():
		var clusterId(string)

		// no --clusterId flag provided, reading nodepool.json..
		if *c2ClusterId == "" {
			// configure file system..
			cleanUp = false
			configDirPath := configureFileSystem(*configDir, cleanUp)

			// read nodepool.json..
			configFilePath := configDirPath + string(os.PathSeparator) + "nodepool.json"
			content, err := ioutil.ReadFile(configFilePath)
			if err != nil {
				fmt.Println("OKECTL :: No --clusterId flag provided, error reading nodepool.json at specified path :: Exiting..")
				fmt.Println(err)
				os.Exit(3)
			}

			// get clusterId from nodepool.json..
			jsonParsed, err := gabs.ParseJSON(content)
			clusterId = (jsonParsed.Path("clusterId").String())
			*c2ClusterId = clusterId[1 : len(clusterId)-1]
		}

		fmt.Println("")
		fmt.Println("OKECTL :: Create kubeconfig :: Request Parameters ...")
		fmt.Println("-------------------------------------------------------")
		fmt.Println("configDir:", *configDir)
		fmt.Println("clusterId:", *c2ClusterId)
		fmt.Println("")

		// brief pause..
		time.Sleep(5 * time.Second)

		// configure file system..
		cleanUp = false
		configDirPath := configureFileSystem(*configDir, cleanUp)

		// create kubeconfig file..
		getKubeConfig(ctx, c, *c2ClusterId, configDirPath)

		// done..
		fmt.Println("")
		fmt.Println("OKECTL :: Create kubeconfig :: Complete ...")

	// get node pool..
	case g3.FullCommand():
		var nodePoolId(string)

		// configure file system..
		cleanUp = false
		configDirPath := configureFileSystem(*configDir, cleanUp)

		// no --nodePoolId flag provided, reading nodepool.json..
		if *g3NodePoolId == "" {

			// read nodepool.json..
			configFilePath := configDirPath + string(os.PathSeparator) + "nodepool.json"
			content, err := ioutil.ReadFile(configFilePath)
			if err != nil {
				fmt.Println("OKECTL :: No --nodePoolId flag provided, error reading nodepool.json at specified path :: Exiting..")
				fmt.Println(err)
				os.Exit(3)
			}

			// get node pool id from nodepool.json..
			jsonParsed, err := gabs.ParseJSON(content)
			nodePoolId = (jsonParsed.Path("id").String())
			*g3NodePoolId = nodePoolId[1 : len(nodePoolId)-1]
		}

		if *g3TfExternalDs == "false" {
			fmt.Println("")
			fmt.Println("OKECTL :: Get NodePool :: Request Parameters ...")
			fmt.Println("-------------------------------------------------------")
			fmt.Println("nodePoolId:", *g3NodePoolId)
			fmt.Println("waitNodesActive:", *g3WaitNodesActive)
			fmt.Println("tfExternalDs:", *g3TfExternalDs)
			fmt.Println("")
		}

		// brief pause..
		time.Sleep(5 * time.Second)

		// wait for create node completion - all nodes..
		if *g3WaitNodesActive != "false" {
			nodeProvisioning := true
			// This loop continues while "nodeProvisioning" is true..
			for nodeProvisioning {
				getNodeLifeCycleState(ctx, c, *g3NodePoolId)
				// if we are waiting for all nodes to be active..
				if *g3WaitNodesActive == "all" {
					boolLcs := (strings.Contains(nodeLifeCycleState, "ATING"))
						if nodeLifeCycleState != "{}" {
							if boolLcs != true {
								nodeProvisioning = false
							}
						}
				}
				// if we are waiting for any nodes to be active..
				if *g3WaitNodesActive == "any" {
					boolLcs := (strings.Contains(nodeLifeCycleState, "ACTIVE"))
						if nodeLifeCycleState != "{}" {
							if boolLcs == true {
								nodeProvisioning = false
							}
						}
				}
					time.Sleep(15 * time.Second)
			}
		}

		// get nodepool details & create nodepool.json..
		getNodePool(ctx, c, *g3NodePoolId, configDirPath)

		// done, output config data..
		// if we are running as a terraform external data source, return only json data..
		if *g3TfExternalDs == "true" {

			// read nodepool.json..
			configFilePath := configDirPath + string(os.PathSeparator) + "nodepool.json"
			content, err := ioutil.ReadFile(configFilePath)
			if err != nil {
				fmt.Println("OKECTL :: No --nodePoolId flag provided, error reading nodepool.json at specified path :: Exiting..")
				fmt.Println(err)
				os.Exit(3)
			}

			// get worker node publicip from nodepool.json..
			jsonParsed, err := gabs.ParseJSON(content)
			nodeIps := (jsonParsed.Path("nodes.publicIp").String())
			nodeIp := findIP(nodeIps)

			jsonObj := gabs.New()
			jsonObj.Set(nodeIp, "workerNodeIp")

			fmt.Println(jsonObj.String())

		} else {
			// not running as terraform external data source, return verbose output..
			fmt.Println("")
			fmt.Println("OKECTL :: Get NodePool :: Complete ...")
			fmt.Println("-------------------------------------------------------")
			configFilePath := configDirPath + string(os.PathSeparator) + "nodepool.json"
				content, err := ioutil.ReadFile(configFilePath)
				if err != nil {
					fmt.Println(err)
				}
				strNodePool := string(content)
				fmt.Println(strNodePool)
		}
	}
}

// functions..
// configure local file system..
func configureFileSystem(configDir string, cleanUp bool) (configDirPath string) {

	// if default configDir..
	if configDir == ".okectl" {
		// find our okectl binary path..
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			fmt.Println(err)
		}
		// clean-up & create configDir..
		configDirPath = (dir + string(os.PathSeparator) + configDir)
		if cleanUp == true {
			err = os.RemoveAll(configDirPath)
			if err != nil {
				fmt.Println(err)
			}
		}
		if _, err := os.Stat(configDir); err != nil {
			err = os.MkdirAll(configDirPath, 0777)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	// if custom configDir..
	if configDir != ".okectl" {
		if _, err := os.Stat(configDir); err == nil {
		// specified configDir exists..
		configDirPath = configDir
		} else {
			// specified configDir does not exist..
			fmt.Println("OKECTL :: Directory --configDir not found :: Exiting ...")
			os.Exit(3)
		}
	}

	return configDirPath
}

// create cluster..
func createCluster(
	ctx context.Context,
	client containerengine.ContainerEngineClient,
	clusterName, vcnId, compartmentId, kubeVersion, subnet1Id, subnet2Id string) containerengine.CreateClusterResponse {

	req := containerengine.CreateClusterRequest{}
	req.Name = common.String(clusterName)
	req.CompartmentId = common.String(compartmentId)
	req.VcnId = common.String(vcnId)
	req.KubernetesVersion = common.String(kubeVersion)
	req.Options = &containerengine.ClusterCreateOptions{
		ServiceLbSubnetIds: []string{subnet1Id, subnet2Id},
		AddOns: &containerengine.AddOnOptions{
			IsKubernetesDashboardEnabled: common.Bool(true),
			IsTillerEnabled:              common.Bool(true),
		},
	}

	fmt.Println("OKECTL :: Create Cluster :: Submitted ...")
	resp, err := client.CreateCluster(ctx, req)
	helpers.FatalIfError(err)

	return resp
}

// delete cluster..
func deleteCluster(ctx context.Context, client containerengine.ContainerEngineClient, clusterId string) containerengine.DeleteClusterResponse {

	req := containerengine.DeleteClusterRequest{
		ClusterId: common.String(clusterId),
	}

	fmt.Println("OKECTL :: Delete Cluster :: Submitted ...")
	resp, err := client.DeleteCluster(ctx, req)
	helpers.FatalIfError(err)

	return resp
}

// create nodepool..
func createNodePool(
	ctx context.Context,
	client containerengine.ContainerEngineClient,
	compartmentId, clusterName, clusterId, kubeVersion, nodeImageName, nodeShape, nodeSshKey, subnet3Id, subnet4Id, subnet5Id string, quantityWkrSubnets, quantityPerSubnet int) containerengine.CreateNodePoolResponse {

	req := containerengine.CreateNodePoolRequest{}
	req.CompartmentId = common.String(compartmentId)
	req.Name = common.String(clusterName)
	req.ClusterId = common.String(clusterId)
	req.KubernetesVersion = common.String(kubeVersion)
	req.NodeImageName = common.String(nodeImageName)
	req.NodeShape = common.String(nodeShape)
	// worker node ssh key..
	if nodeSshKey != "Null" {
		req.SshPublicKey = common.String(nodeSshKey)
	}
	// worker subnets..
	if quantityWkrSubnets == 1 {
		req.SubnetIds = []string{subnet3Id}
	}
	if quantityWkrSubnets == 2 {
		req.SubnetIds = []string{subnet3Id, subnet4Id}
	}
	if quantityWkrSubnets == 3 {
		req.SubnetIds = []string{subnet3Id, subnet4Id, subnet5Id}
	}
	req.QuantityPerSubnet = common.Int(quantityPerSubnet)

	fmt.Println("OKECTL :: Create NodePool :: Submitted ...")
	resp, err := client.CreateNodePool(ctx, req)
	helpers.FatalIfError(err)

	return resp
}

// delete nodepool
func deleteNodePool(ctx context.Context, client containerengine.ContainerEngineClient, nodePoolID *string) {
	deleteReq := containerengine.DeleteNodePoolRequest{
		NodePoolId: nodePoolID,
	}

	client.DeleteNodePool(ctx, deleteReq)

	fmt.Println("deleting nodepool")
}

// get worker node lifecycle status..
func getNodeLifeCycleState(
	ctx context.Context,
	client containerengine.ContainerEngineClient,
	nodePoolId string) containerengine.GetNodePoolResponse {

	req := containerengine.GetNodePoolRequest{}
	req.NodePoolId = common.String(nodePoolId)

	resp, err := client.GetNodePool(ctx, req)
	helpers.FatalIfError(err)

	// marshal & parse json..
	nodePoolResp := resp.NodePool
	nodesJson, _ := json.Marshal(nodePoolResp)
	jsonParsed, _ := gabs.ParseJSON(nodesJson)
	nodeLifeCycleState = (jsonParsed.Path("nodes.lifecycleState").String())

	return resp
}

// get nodepool details & create nodepool.json..
func getNodePool(
	ctx context.Context,
	client containerengine.ContainerEngineClient,
	nodePoolId, configDirPath string) containerengine.GetNodePoolResponse {

	req := containerengine.GetNodePoolRequest{}
	req.NodePoolId = common.String(nodePoolId)

	if *g3TfExternalDs == "false" {
		fmt.Println("OKECTL :: Getting NodePool Data ...")
	}

	resp, err := client.GetNodePool(ctx, req)
	helpers.FatalIfError(err)

	// create output file..
	configFilePath := configDirPath + string(os.PathSeparator) + "nodepool.json"
	file, err := os.Create(configFilePath)
	if err != nil {
		fmt.Println("OKECTL :: Error Creating nodepool.json File:", err)
	}
	defer file.Close()

	// populate nodepool.json file..
	nodePoolResp := resp.NodePool
	nodesJsonIndent, _ := json.MarshalIndent(nodePoolResp, "", "\t")
	err = ioutil.WriteFile(configFilePath, nodesJsonIndent, 0666)
	if err != nil {
		fmt.Println("OKECTL :: Error Writing nodepool.json File:", err)
	}
	helpers.FatalIfError(err)

	return resp
}

// create kubeconfig..
func getKubeConfig(
	ctx context.Context,
	client containerengine.ContainerEngineClient,
	clusterId, configDirPath string) containerengine.CreateKubeconfigResponse {

	req := containerengine.CreateKubeconfigRequest{}
	req.ClusterId = common.String(clusterId)
	req.Expiration = common.Int(360)

	fmt.Println("OKECTL :: Getting kubeconfig Data ...")

	// create output file..
	configFilePath := configDirPath + string(os.PathSeparator) + "kubeconfig"
	file, err := os.Create(configFilePath)
	if err != nil {
		fmt.Println("Cannot create kubeconfig file", err)
	}
	defer file.Close()

	// populate output file..
	resp, err := client.CreateKubeconfig(ctx, req)
	_, err = io.Copy(file, resp.Content)
	if err != nil {
		fmt.Println("OKECTL :: Error Writing kubeconfig File:", err)
	}
	helpers.FatalIfError(err)

	return resp
}

// locate ip address..
func findIP(input string) string {
	numBlock := "(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])"
	regexPattern := numBlock + "\\." + numBlock + "\\." + numBlock + "\\." + numBlock

	regEx := regexp.MustCompile(regexPattern)
	return regEx.FindString(input)
}

// wait until work request finishes..
func waitUntilWorkRequestComplete(client containerengine.ContainerEngineClient, workReuqestID *string) containerengine.GetWorkRequestResponse {
	// retry GetWorkRequest call until TimeFinished is set..
	shouldRetryFunc := func(r common.OCIOperationResponse) bool {
		return r.Response.(containerengine.GetWorkRequestResponse).TimeFinished == nil
	}

	getWorkReq := containerengine.GetWorkRequestRequest{
		WorkRequestId:   workReuqestID,
		RequestMetadata: helpers.GetRequestMetadataWithCustomizedRetryPolicy(shouldRetryFunc),
	}

	getResp, err := client.GetWorkRequest(context.Background(), getWorkReq)
	helpers.FatalIfError(err)
	return getResp
}

// getResourceID return a resource ID based on the filter of resource actionType and entityType..
func getResourceID(resources []containerengine.WorkRequestResource, actionType containerengine.WorkRequestResourceActionTypeEnum, entityType string) *string {
	for _, resource := range resources {
		if resource.ActionType == actionType && strings.ToUpper(*resource.EntityType) == entityType {
			return resource.Identifier
		}
	}

	fmt.Println("OKECTL :: Unable to obtain Resource ID ...")
	return nil
}
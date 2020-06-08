[terraform]: https://terraform.io
[oci]: https://cloud.oracle.com/en_US/cloud-infrastructure
[go]: https://golang.org/dl/
[oke]: https://cloud.oracle.com/containers/kubernetes-engine
[oke-guide]: https://docs.cloud.oracle.com/iaas/Content/ContEng/Concepts/contengprerequisites.htm?tocpath=Services%7CContainer%20Engine%7CPreparing%20for%20Container%20Engine%20for%20Kubernetes%7C_____0
[go-sdk]: https://github.com/oracle/oci-go-sdk
[config-file]: https://docs.cloud.oracle.com/iaas/Content/API/Concepts/sdkconfig.htm#CLIConfiguration
[kubectl-guide]: https://kubernetes.io/docs/tasks/tools/install-kubectl/


# okectl: CLI utility for OKE cluster lifecycle management

## About

okectl is a CLI utility designed for use with [Oracle Container Engine for Kubernetes (OKE)][oke]. okectl provides a command-line interface for interaction with OKE, including Kubernetes cluster lifecycle.

okectl is designed as a stand-alone tool to automate the Kubernetes cluster creation process, but is most useful when used as part of an automation pipeline.

[Oracle Container Engine for Kubernetes][oke] is a developer friendly, container-native, and enterprise-ready managed Kubernetes service for running highly available clusters with the control, security, and predictable performance of Oracle’s Cloud Infrastructure.

okectl is built using the [Go SDK for Oracle Cloud Infrastructure][go-sdk].

### Supported Operations

 - `createOkeCluster`
    - Creates cluster control plane, node pool, worker nodes, & configuration data (kubeconfig & json cluster desctiption).
 - `deleteOkeCluster`
    - Deletes specified cluster.
 - `getOkeNodePool`
    - Retreives cluster, node poool, and node details for a specified node pool.
 - `createOkeKubeconfig`
    - Creates kubeconfig authentication artefact for kubectl.

## Usage

okectl requires configuration data via command-line arguments/flags. Command-line flags provide data relating to both the [OCI][oci] tenancy, and also OKE cluster configuration parameters.

### Example - Usage
```
$ ./okectl
$ usage: OKECTL [<flags>] <command> [<args> ...]
$
$ A command-line application for configuring Oracle OKE (Container Engine for Kubernetes.)
$
$ Flags:
$   --help                 Show context-sensitive help (also try --help-long and --help-man).
$   --configDir=".okectl"  Path where output files are created - e.g. kubeconfig file.
$   --version              Show application version.
$
$ Commands:
$   help [<command>...]
$     Show help.
$
$   createOkeCluster --vcnId=VCNID --compartmentId=COMPARTMENTID --subnet1Id=SUBNET1ID --subnet2Id=SUBNET2ID --subnet3Id=SUBNET3ID [<flags>]
$     Create new OKE Kubernetes cluster.
$
$   deleteOkeCluster --clusterId=CLUSTERID
$     Delete OKE Kubernetes cluster.
$
$   getOkeNodePool [<flags>]
$     Get cluster, node poool, and node details for a specified node pool.
$
$   createOkeKubeconfig --clusterId=CLUSTERID
$     Create kubeconfig authentication artefact for kubectl.
```

### Example - Create Cluster

#### Interactive Help

```
$ ./okectl createOkeCluster --help
$
$ usage: OKECTL createOkeCluster --vcnId=VCNID --compartmentId=COMPARTMENTID --subnet1Id=SUBNET1ID --subnet2Id=SUBNET2ID --subnet3Id=SUBNET3ID [<flags>]
$
$ Create new OKE Kubernetes cluster.
$
$ Flags:
$   --help                              Show context-sensitive help (also try --help-long and --help-man).
$   --configDir=".okectl"               Path where output files are created - e.g. kubeconfig file. Specify as absolute path.
$   --version                           Show application version.
$   --vcnId=VCNID                       OCI VCN-Id where cluster will be created.
$   --compartmentId=COMPARTMENTID       OCI Compartment-Id where cluster will be created.
$   --subnet1Id=SUBNET1ID               Cluster Control Plane LB Subnet 1.
$   --subnet2Id=SUBNET2ID               Cluster Control Plane LB Subnet 2.
$   --subnet3Id=SUBNET3ID               Worker Node Subnet 1.
$   --subnet4Id=SUBNET4ID               Worker Node Subnet 2.
$   --subnet5Id=SUBNET5ID               Worker Node Subnet 3.
$   --clusterName="dev-oke-001"         Kubernetes cluster name.
$   --kubeVersion="v1.10.3"             Kubernetes cluster version.
$   --nodeImageName="Oracle-Linux-7.4"  OS image used for Worker Node(s).
$   --nodeShape="VM.Standard1.1"        CPU/RAM allocated to Worker Node(s).
$   --nodeSshKey=NODESSHKEY             SSH key to provision to Worker Node(s) for remote access.
$   --quantityWkrSubnets=1              Number of subnets used to host Worker Node(s).
$   --quantityPerSubnet=1               Number of Worker Nodes per subnet.
$   --waitNodesActive="false"           If waitNodesActive=all, wait & return when all nodes in the pool are active.
                                        If waitNodesActive=any, wait & return when any of the nodes in the pool are active.
                                        If waitNodesActive=false, no wait & return when the node pool is active.
```

#### Create Cluster

```
$ ./okectl createOkeCluster \
$ --clusterName=OKE-Cluster-001 \
$ --kubernetesVersion=v1.10.3 \
$ --vcnId=ocid1.vcn.oc1.iad.aaaaaaaamg7tqzjpxbbibev7lhp3bhgtcmgkbbrxr7td4if5qa64bbekdxqa \
$ --compartmentId=ocid1.compartment.oc1..aaaaaaaa2id6dilongtlxxmufoeunasaxuv76xxcb4ewxcxxxw5eba \
$ --quantityWkrSubnets=1 \
$ --quantityPerSubnet=1 \
$ --subnet1Id=ocid1.subnet.oc1.iad.aaaaaaaagq5apzuwr2qnianczzie4ffo6t46rcjehnsyoymiuunxaauq7y7a \
$ --subnet2Id=ocid1.subnet.oc1.iad.aaaaaaaadxr6zl4jpmcaxd4izzlvbyq2pqss3pmotx6dnusmh3ijorrpbhva \
$ --subnet3Id=ocid1.subnet.oc1.iad.aaaaaaaabf6k3ufcjdsdb5xfzzc3ayplhpip2jxtnaqvfcpakxt3bhmhecxa \
$ --nodeImageName=Oracle-Linux-7.4 \
$ --nodeShape=VM.Standard1.1 \
$ --nodeSshKey="ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDsHX7RR0z+JSAf+5nfTO9kS4Y6HV2pPXoXTqUJH..." \
$ --waitNodesActive="all"
```
For the above request, okectl will provision:
 - Kubernetes Cluster (Control Plane)
       - Version will be as nominated via the `--kubeVersion` flag.
 - Node Pool
       - Node Pool will be created across the number of worker subnets as provided via `--quantityWkrSubnets` flag.
 - Nodes
       - Worker nodes will be provisioned to each of the nominated worker subnets. Number of worker nodes per subnet is determined by the `--quantityPerSubnet` flag.
 - Configuration Data
       - Provision to local filesystem a kubeconfig authentication artefact (kubeconfig) & json description of cluster configuration (nodeconfig.json).

Per the flag --waitNodesActive="all", okectl will return when cluster, node pool, and each of the nodes in the node pool are active.

Once completed, okectl will output the cluster, nodepool and node configuration data (stdout):

```
$ OKECTL :: Create Cluster :: Complete ...
$ --------------------------------------------------------------------------------------
$ {
$        "id": "ocid1.nodepool.oc1.iad.aaaaaaaaae3tonjqgftdiyrxha2gczrtgu3winbtgbsdszjqmnrdeodegu2t",
$        "compartmentId": "ocid1.compartment.oc1..aaaaaaaa2id6dilongtl6fmufoeunasaxuv76b6cb4ewxcw4juafe55w5eba",
$        "clusterId": "ocid1.cluster.oc1.iad.aaaaaaaaae2tgnlbmzrtknjygrrwmobsmvrwgnrsmnqtmzjygc2domtbgmyt",
$        "name": "oke-dev-001",
$        "kubernetesVersion": "v1.10.3",
$        "nodeImageId": "ocid1.image.oc1.iad.aaaaaaaajlw3xfie2t5t52uegyhiq2npx7bqyu4uvi2zyu3w3mqayc2bxmaa",
$        "nodeImageName": "Oracle-Linux-7.4",
$        "nodeShape": "VM.Standard1.1",
$        "initialNodeLabels": [],
$        "sshPublicKey": "",
$        "quantityPerSubnet": 1,
$        "subnetIds": [
$                "ocid1.subnet.oc1.iad.aaaaaaaajvfrxxawuwhvxnjliox7gzibonafqcyjkdozwie7q5po7qbawl4a"
$        ],
$        "nodes": [
$                {
$                        "id": "ocid1.instance.oc1.iad.abuwcljtayee6h7ttavqngewglsbe3b6my3n2eoqawhttgtswsu66lrjgi4q",
$                        "name": "oke-c2domtbgmyt-nrdeodegu2t-soxdncj6x5a-0",
$                        "availabilityDomain": "Ppri:US-ASHBURN-AD-3",
$                        "subnetId": "ocid1.subnet.oc1.iad.aaaaaaaattodyph6wco6cmusyza4kyz3naftwf6yjzvog5h2g6oxdncj6x5a",
$                        "nodePoolId": "ocid1.nodepool.oc1.iad.aaaaaaaaae3tonjqgftdiyrxha2gczrtgu3winbtgbsdszjqmnrdeodegu2t",
$                        "publicIp": "100.211.162.17",
$                        "nodeError": null,
$                        "lifecycleState": "UPDATING",
$                        "lifecycleDetails": "waiting for running compute instance"
$                }
$        ]
$ }
```

By default, okectl will create a sub-directory named ".okectl" within the same directory as the okectl binary. okectl will create x2 files within the ".okectl" directory:

 - `kubeconfig`
       - This file contains authentication and cluster connection information. It should be used with the `kubectl` command-line utility to access and configure the cluster.
 - `nodepool.json`
       - This file contains a detailed output of the cluster and node pool configuration in json format.

Output directory is configurable via the `--configDir` flag. Path provided to `--configDir` should be provided as an absolute path.

All clusters created using okectl will be provisioned with the additional options of the Kubernetes dashboard & Helm/Tiller as installed.

### Example - Get Node Pool

#### Interactive Help

```
$ ./okectl getOkeNodePool --help
$
$ usage: OKECTL getOkeNodePool [<flags>]
$
$ Get cluster, node pool, and node details for a specified node pool.
$
$ Flags:
$   --help                     Show context-sensitive help (also try --help-long and --help-man).
$   --configDir=".okectl"      Path where output files are created or referenced - e.g. kubeconfig file. Specify as absolute path.
$   --version                  Show application version.
$   --nodePoolId=NODEPOOLID    OKE Node Pool Id. If not specified, Id contained in nodepool.json will be used.
$   --tfExternalDs="false"     Run as a Terraform External Data Source, & provide json only response data.
$   --waitNodesActive="false"  If waitNodesActive=all, wait & return when all nodes in the pool are active. If waitNodesActive=any, wait & return when any of the nodes in the pool
$                              are active. If waitNodesActive=false, no wait & return when the node pool is active.
```

#### Get Node Pool

```
$ ./okectl getOkeNodePool \
$ --tfExternalDs="false" \
$ --waitNodesActive="all"
```

For the above request, okectl will provision:
 - Configuration Data
       - Provision to local filesystem a json description of cluster configuration (nodeconfig.json).

Per the flag --waitNodesActive="all", okectl will return when cluster, node pool, and each of the nodes in the node pool are active.

Once completed, okectl will output the cluster, nodepool and node configuration data (stdout):

```
OKECTL :: Get NodePool :: Complete ...
-------------------------------------------------------
{
        "id": "ocid1.nodepool.oc1.iad.aaaaaaaaafswgzjyguywemdcgbrtinzygaywmmjwg44tqntbgnzwmyzrgm3d",
        "compartmentId": "ocid1.compartment.oc1..aaaaaaaa2id6dilongtl6fmufoeunasaxuv76b6cb4ewxcw4juafe55w5eba",
        "clusterId": "ocid1.cluster.oc1.iad.aaaaaaaaae4tsyryg4zwkobvmyzdenzwgjsdiolbgyytcmrymc2wimbqg5rd",
        "name": "dev000-oke",
        "kubernetesVersion": "v1.11.1",
        "nodeImageId": "ocid1.image.oc1.iad.aaaaaaaa2tq67tvbeavcmioghquci6p3pvqwbneq3vfy7fe7m7geiga4cnxa",
        "nodeImageName": "Oracle-Linux-7.4",
        "nodeShape": "VM.Standard2.2",
        "initialNodeLabels": [],
        "sshPublicKey": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDHzATp/2KhhrVF0CiI6sHX7HA0z+JSAf+5JF5zdD7KnKsO9kS4Y6HV2vPPuV/z/IWIOLQeNOgXZQyC832oOdSAPu7/sag7PxpPXoXTqUJH+hc8zDUJ/WegX1dVhm3zZjU7TvvsjKJMUWO0c7TaRglebkcoMGzTMtU9WHF/7fJ8npOv4DSMC7Y7Ss1263vffpqnUpeBCsAHT6v+JuMsL6wEdYnQnY4GslmS3GTItQ1J2gNBlnMOyfVTOsyQNyw2sxE1AyvYvgxiZRZ1IYOth1al5uJQjEirjrb3llJgKQgMjwAX3zhPBa9E0UzyOx9YuaWJ2Yq8xP3OZ2Jh913KWlLT",
        "quantityPerSubnet": 1,
        "subnetIds": [
                "ocid1.subnet.oc1.iad.aaaaaaaaafa2y2dgywmjbtl6zyvgl2eucgkst3xfunfxm46lyrqg2jvdbjaq"
        ],
        "nodes": [
                {
                        "id": "ocid1.instance.oc1.iad.abuwcljswg6w4tl4mge46pwfmxjv3zdvkgh4fdu3umfdgpkkrwnymv76eypq",
                        "name": "oke-c2wimbqg5rd-nzwmyzrgm3d-rqg2jvdbjaq-0",
                        "availabilityDomain": "Ppri:US-ASHBURN-AD-1",
                        "subnetId": "ocid1.subnet.oc1.iad.aaaaaaaaafa2y2dgywmjbtl6zyvgl2eucgkst3xfunfxm46lyrqg2jvdbjaq",
                        "nodePoolId": "ocid1.nodepool.oc1.iad.aaaaaaaaafswgzjyguywemdcgbrtinzygaywmmjwg44tqntbgnzwmyzrgm3d",
                        "publicIp": "132.145.156.184",
                        "nodeError": null,
                        "lifecycleState": "ACTIVE",
                        "lifecycleDetails": ""
                }
        ]
}
```

Where the flag --tfExternalDs="true" is applied, okectl will run as a [Terraform external data source](https://www.terraform.io/docs/providers/external/data_source.html). The Terraform external data source allows an external program implementing a specific protocol to act as a data source, exposing arbitrary data for use elsewhere in the Terraform configuration.

In this circumstance, okectl will provide json only response data containing the public IP address of a cluster worker node:

```
$ ./okectl getOkeNodePool --tfExternalDs=true
$ {"workerNodeIp":"132.145.156.184"}
```

In combination with the --waitNodesActive flag, this provides the ability to have Terraform wait for worker nodes to be active, then proceed to call a remote-exec provisioner against the worker node via the public IP address returned (e.g. configure cluster or deploy workloads).

### Accessing a cluster

The Kubernetes cluster will be running after the okectl `createOkeCluster` operation completes.

#### Cluster Operations via CLI

To operate the cluster using the `kubectl` CLI, first ensure its installed per this [configuration guide][kubectl-guide]. You can then submit requests to the OKE kube api by invoking `kubectl` and specifying the path to the `kubeconfig` file:

```
$ kubectl cluster-info --kubeconfig=\path-to-oke-go\config\kubeconfig
```

#### Cluster Operations via Dashboard

To access the Kubernetes dashboard, ensure that you have kubectl installed & run the following command:

```
$ kubectl proxy --kubeconfig=\path-to-oke-go\config\kubeconfig
```
Open a web browser and request the following URL:
http://localhost:8001/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/

The kube dashboard will request authentication method - select _kubeconfig_ as the authentication method, & point to the local kubeconfig file generated by okectl.


## Configuration

Deploying an OKE cluster to OCI requires that certain configuration prerequisites be met on the host system that is running the utility, and in the target OCI tenancy.

### Dependencies

#### OCI

Deploying an OKE cluster to OCI requires that certain configuration prerequisites be met in the target OCI tenancy. These include a Compartment, VCN, Subnets, Internet Gateway, Route Table and Security Lists.
See the following guide which provides step-by-step instruction on configuring the dependencies: [Preparing for Container Engine for Kubernetes][oke-guide].

#### Environment

Basic configuration information (for example, user credentials and OCI tenancy OCID) is required in order for the utility to work. You can provide this information by configuring one of the following:

 - Using a configuration file:
   See the following [configuration guide][config-file] for instruction on building a configuration file.

 - Using environment variables:
  ```
    TF_VAR_user_ocid = (string)
        OCID of the user calling the API
    TF_VAR_fingerprint = (string)
        Fingerprint for the key pair being used
    TF_VAR_private_key_path = (string)
        Full path and filename of the private key
    TF_VAR_region = (string)
        Oracle Cloud Infrastructure region of your tenancy
    TF_VAR_tenancy_ocid = (string)
        OCID of your tenancy
  ```
okectl will automatically check for the presence of the configuration file & environment variables at runtime.

#### Debug

okectl will provide detailed debug information to stdout when specifying the environment varable:
  ```
    OCI_GO_SDK_DEBUG = 1
  ```

## Building okectl from source

### Dependencies

 - Install [Go programming language][go]
 - Install [Go SDK for Oracle Cloud Infrastructure][go-sdk]

### Build

```
$ go build okectl.go
```

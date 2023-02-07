# CNP-CIDR-Allocator

Used to allocate podCIDR subnets to nodes from a pool

## Description

CNP-CIDR-Allocator is a Kubernetes operator consisting of a CRD and Controller. The controller expects a `NodeCIDRAllocation` custom resource (CR) to be specified that will identify `addressPools` that will be used as a basis for CIDR allocation. The CR also expects a `NodeSelector` to be specified so that the controller can identify which nodes should be targeted for CIDR allocations.

For an example, please take a look at [config/samples](/config/samples/)

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

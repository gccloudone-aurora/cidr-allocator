## CIDR-Allocator

The CIDR-Allocator is a Kubernetes Operator that helps to implement dynamic IPAM irrespective of the Container Network Interface (CNI) being used.

At Statistics Canada, this operator is used to address an early design consideration for the Cloud Native Platform 2.0 (CNP2.0) related to our BGP route propagation solution.

In Kubernetes, a full PodCIDR **must be** allocated to a Node at creation-time since any modifications afterwards ti the `PodCIDR` or `PodCIDRs` fields are strictly prohibited.

This project follows the [`Kubernetes Operator Pattern`](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

### Architecture

![CIDR-Allocator Solution Architecture](./docs/media/cidr_allocator_solution_architecture.svg)


The controller watches for a [`NodeCIDRAllocation`](./api/v1alpha1/nodecidrallocation_types.go) custom resource (CR) that will identify blocks of IPv4 addresses that will be used during the allocation of a `PodCIDR` range to a Node. A `NodeSelector` is used to identify which `Node` resources should align with each `NodeCIDRAllocation` that is defined. This gives us the flexibility to manage Pod IP allocation with as much or as little granularity as desired.

> By default, the size of the assigned `PodCIDR` range will be equal to the `MaxPods` attribute on the `Node` resource

### Installation

Install `CIDR-Allocator` from the official StatCan Helm Chart

```bash
helm repo add statcan https://statcan.github.io/charts
helm repo update
helm install my-cidr-allocator statcan/cidr-allocator
```

> For an example configuration for the `NodeCIDRAllocation` CR, please take a look at [config/samples](/config/samples/)

### Changelog

Changes to this project are tracked in the [CHANGELOG](/CHANGELOG.md) which uses the [keepachangelog](https://keepachangelog.com/en/1.0.0/) format.

### Test It Out (locally)
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

______________________

## CIDR-Allocator

### Comment contribuer

Voir [CONTRIBUTING.md](CONTRIBUTING.md)

### Licence

Sauf indication contraire, le code source de ce projet est protégé par le droit d'auteur de la Couronne du gouvernement du Canada et distribué sous la [licence MIT](LICENSE).

Le mot-symbole « Canada » et les éléments graphiques connexes liés à cette distribution sont protégés en vertu des lois portant sur les marques de commerce et le droit d'auteur. Aucune autorisation n'est accordée pour leur utilisation à l'extérieur des paramètres du programme de coordination de l'image de marque du gouvernement du Canada. Pour obtenir davantage de renseignements à ce sujet, veuillez consulter les [Exigences pour l'image de marque](https://www.canada.ca/fr/secretariat-conseil-tresor/sujets/communications-gouvernementales/exigences-image-marque.html).

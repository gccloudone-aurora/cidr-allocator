# cidr-allocator

A Helm chart to deploy the STATCAN CIDR-Allocator Controller and CRDs to a Kubernetes Cluster

![Version: 2.0.1](https://img.shields.io/badge/Version-2.0.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v1.0.1](https://img.shields.io/badge/AppVersion-v1.0.1-informational?style=flat-square)

A Helm chart to deploy the STATCAN CIDR-Allocator Controller and CRDs to a Kubernetes Cluster

**Homepage:** <https://statcan.gc.ca>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Ben Sykes | <ben.sykes@statcan.gc.ca> |  |

## Source Code

* <https://github.com/StatCan/cidr-allocator>

## Requirements

Kubernetes: `>= 1.16.0-0`

## Installation

Install using Helm

```bash
helm repo add statcan-ca https://statcan.github.io/cidr-allocator
helm repo update
helm install cidr-allocator statcan-ca/cidr-allocator
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | specifies pod affinities and anti-affinities to apply when scheduling controller pods |
| envVars | list | `[]` | any additional environment vars to pass to container (manager) |
| fullnameOverride | string | `""` | override full name |
| image.pullPolicy | string | `"IfNotPresent"` | can be one of "Always", "IfNotPresent", "Never" |
| image.repository | string | `"statcan/cidr-allocator"` | the source image repository |
| imagePullSecrets | list | `[]` | specifies credentials for a private registry to pull source image |
| leaderElectionEnabled | bool | `true` | specifies whether or not to enable leader-election for the podtracker controller |
| nameOverride | string | `""` | override name |
| nodeCIDRAllocations | list | `[]` |  |
| nodeSelector | object | `{}` | specifies a selector for determining where the controller pods will be scheduled |
| podAnnotations | object | `{}` |  |
| podLabels | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| priorityClassName | string | `""` |  |
| prometheus.enabled | bool | `true` | resource. |
| prometheus.podmonitor.annotations | object | `{}` | Additional annotations to add to the PodMonitor. |
| prometheus.podmonitor.enabled | bool | `false` | Create a PodMonitor to add podtracker to Prometheus. |
| prometheus.podmonitor.endpointAdditionalProperties | object | `{}` | endpoint such as relabelings, metricRelabelings etc.  For example:  endpointAdditionalProperties:   relabelings:   - action: replace     sourceLabels:     - __meta_kubernetes_pod_node_name     targetLabel: instance  +docs:property |
| prometheus.podmonitor.honorLabels | bool | `false` | Keep labels from scraped data, overriding server-side labels. |
| prometheus.podmonitor.interval | string | `"60s"` | The interval to scrape metrics. |
| prometheus.podmonitor.labels | object | `{}` | Additional labels to add to the PodMonitor. |
| prometheus.podmonitor.path | string | `"/metrics"` | The path to scrape for metrics. |
| prometheus.podmonitor.prometheusInstance | string | `"default"` | different PodMonitors. |
| prometheus.podmonitor.scrapeTimeout | string | `"30s"` | The timeout before a metrics scrape fails. |
| prometheus.servicemonitor.annotations | object | `{}` | Additional annotations to add to the ServiceMonitor. |
| prometheus.servicemonitor.enabled | bool | `true` | Create a ServiceMonitor to add podtracker to Prometheus. |
| prometheus.servicemonitor.endpointAdditionalProperties | object | `{}` | endpoint such as relabelings, metricRelabelings etc.  For example:  endpointAdditionalProperties:   relabelings:   - action: replace     sourceLabels:     - __meta_kubernetes_pod_node_name     targetLabel: instance  +docs:property |
| prometheus.servicemonitor.honorLabels | bool | `false` | Keep labels from scraped data, overriding server-side labels. |
| prometheus.servicemonitor.interval | string | `"60s"` | The interval to scrape metrics. |
| prometheus.servicemonitor.labels | object | `{}` | Additional labels to add to the ServiceMonitor. |
| prometheus.servicemonitor.path | string | `"/metrics"` | The path to scrape for metrics. |
| prometheus.servicemonitor.prometheusInstance | string | `"default"` | different ServiceMonitors. |
| prometheus.servicemonitor.scrapeTimeout | string | `"30s"` | The timeout before a metrics scrape fails. |
| prometheus.servicemonitor.targetPort | int | `9003` | podtracker controller is listening on for metrics. |
| rbac.create | bool | `true` | Specifies whether RBAC resources should be created (recommended) |
| replicaCount | int | `2` | number of replicas to create for the controller |
| resources | object | `{}` | resource limits/requests for created resources |
| securityContext | object | `{"runAsNonRoot":true}` | the pod security context which defines privilege and access control settings for the controller Pod |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | If not set and create is true, a name is generated using the fullname template |
| tolerations | list | `[{"operator":"Exists"}]` | specifies which taints can be tolerated by the controller |
| topologySpreadConstraints | list | `[{"labelSelector":{"matchLabels":{"app.kubernetes.io/name":"cidr-allocator"}},"maxSkew":1,"nodeAffinityPolicy":"Honor","nodeTaintsPolicy":"Honor","topologyKey":"kubernetes.io/hostname","whenUnsatisfiable":"DoNotSchedule"}]` | specifies how pods should be scheduled across multiple nodes |

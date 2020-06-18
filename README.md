Observe Kubernetes events, collect metrics & trigger endpoint receivers

## Quick Start

## Configuration

### Kubeobserver configuration

Kubeobserver is configure throw environment variables. 

| Variable name | Mandatory | Description | Default |
| --- | --- | --- | --- |
| K8S_CLUSTER_NAME | true |  Host to listen on for the prometheus exporter | - |
| TELEMETRY_PORT | HTTP Port to listen on for the prometheus exporter | 8080 |

### Client settings

When kubeobserver is running inside k8s, client (pods, config-maps and so on) can define what to watch and which receviers they want to use.<br>
The configuration is made by using k8s controller annotations under the root template, for example:

```bash
...
 template:
    metadata:
      labels:
        app: {{ template "name" . }}
    annotations:
        pod-init-container-kubeobserver.io/watch: true
        kubeobserver.io/receivers "slack,alert-manager"
...        
```

<b>Note: if annotations are not defined, default values will be used basebased on kubeobserver configuration</b><br>


| Variable name | Mandatory | Description | Default |
| --- | --- | --- | --- |
| K8S_CLUSTER_NAME | true |  Host to listen on for the prometheus exporter | - |
| TELEMETRY_PORT | HTTP Port to listen on for the prometheus exporter | 8080 |

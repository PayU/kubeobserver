Observe Kubernetes events, collect metrics & trigger endpoint receivers

## Quick Start

### Configuration

#### Kubeobserver configuration
fdgfdgfd
fgfdgfdgdf
dfgfdgfdg

#### Client settings

| controller-name | k8s-annotation-name | value-type | description | default |
| :--- | :--- | :--- | :--- | :--- | :--- |
| pod-watcher | pod-init-container-kubeobserver.io/watch | boolean | pod watcher will trigger events for init containers related to the pod | false |
| *All* | kubeobserver.io/receivers | comma separated string | bla bla bla | default recevier is defined in kubeobserver using DEFAULT_RECEIVER env variable |

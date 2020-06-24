## 1.0.0 (June 25th, 2020)

FEATURES:

 * **Pod Watcher**: Listens on pod(s) events from k8s api servers  
 * **Slack Receiver**: Receiving watchers events and send them to a configurable slack channel  
 * **Slack User Mention - Crashloopback**: Enbales to mention a specific users when Crashloopback event occuers
 * **Init Containers Event Watch**: Enables to watch init containers events inside a pod
 * **Pod-Watcher - Ignore Update Event**: Enables to ingore update events inside pod watcher and only notify about Add/Delete events
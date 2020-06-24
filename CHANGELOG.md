## 1.0.0 (June 25th, 2020)

FEATURES:

 * **Pod Watcher**: Listens to pod(s) events from k8s api servers  
 * **Slack Receiver**: Receives watcher events and sends them to a configurable slack channel  
 * **Slack User Mention - Crashloopback**: Enables mentioning a specific user when CrashLoopBack event occurs
 * **Init Containers Event Watch**: Enables watching init-containers events inside a pod
 * **Pod-Watcher - Ignore Update Event**: Enables ignoring update events inside the Pod Watcher and only notify about Add/Delete events
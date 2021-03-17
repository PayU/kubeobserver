## 1.3.1 (March 17th, 2021)

FEATURES:
 * **Slack User Mention - HorizontalPodAutoScaler**: Enables mentioning a specific user when HPA event occurs

## 1.3.0 (March 16th, 2021)
 
CHANGES:
 * **Pod-Watcher - Watch Update**: Pod update events will not be watched as default

## 1.2.0 (December 22th, 2020)
 
FEATURES:
 * **Pod-Watcher - Ignore Pod Event**: Enables ignoring all events using annotation

## 1.1.0 (December 21th, 2020)
 
FEATURES:
 * **HPA Watcher**: Add Horizontal Pod Autoscaler events watcher.
 * **Configuration**: Add WATCHER_THREADS configuration - controller events can be process in parallel when value > 1

BUG FIXES:
 * Fixed panic for nil pointer reference on pod watcher controller

QUALITY:
 * Added end2end tests using kind cluster

## 1.0.0 (June 25th, 2020)

FEATURES:

 * **Pod Watcher**: Listens to pod(s) events from k8s api servers  
 * **Slack Receiver**: Receives watcher events and sends them to a configurable slack channel  
 * **Slack User Mention - Crashloopback**: Enables mentioning a specific user when CrashLoopBack event occurs
 * **Init Containers Event Watch**: Enables watching init-containers events inside a pod
 * **Pod-Watcher - Ignore Update Event**: Enables ignoring update events inside the Pod Watcher and only notify about Add/Delete events

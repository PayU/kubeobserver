## E2E Testing

End-to-end automated testing for the kubeobserver.

---

### How it works

The tests use the kind project [https://github.com/kubernetes-sigs/kind] that has features for setting up the necessary K8s resources and for interacting with a cluster.

### How to run tests

**prepare the kind cluster, metric server & test application**

```
./tests/scripts/env-setup.sh
```

**run tests**
```
./tests/e2e-tests.sh
```

**clear & delete local test environment**
```
./tests/scripts/env-destory.sh
```
---


# Multitenant controller
This controller emulates the behaviour of the `ovs-multitenant` pod network plugin.

Cluster users label their namespaces as follows:

```
...
metadata:
  labels:
    multitenant-pod-network/group: mygroup
```

Namespaces with the same group label can communicate with one another. Namespaces with group label `global` can access and are reachable from all namespaces.

Isolation occurs only at the namespace level. The generated NetworkPolicy objects use only a small subset of the capabilities of the NetworkPolicy specification. Avoiding the complexity of pod-level isolation and whitelisting is the main purpose of this controller.

## Run out of cluster
```
$ go build
$ ./multitenant-controller -kubeconfig=${HOME}/.kube/config
```

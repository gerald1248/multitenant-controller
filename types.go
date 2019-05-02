package main

import (
	"sync"
        "k8s.io/client-go/kubernetes"
        "k8s.io/client-go/tools/cache"
        "k8s.io/client-go/util/workqueue"
)

type Controller struct {
        indexer     cache.Indexer
        queue       workqueue.RateLimitingInterface
        informer    cache.Controller
        clientset   kubernetes.Interface
        mutex       *sync.Mutex
        state       map[string]string // map[NAMESPACE]GROUP
}

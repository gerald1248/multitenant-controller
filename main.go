package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"

	au "github.com/logrusorgru/aurora"
	rest "k8s.io/client-go/rest"
)

const labelPrefix = "multitenant-pod-network"
const labelNameGroup = "group"
const labelNameOwner = "owner"

// NewController acts as the central controller constructor
func NewController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller, clientset kubernetes.Interface, mutex *sync.Mutex, state map[string]string) *Controller {
	return &Controller{
		informer:    informer,
		indexer:     indexer,
		queue:       queue,
		clientset:   clientset,
		mutex:       mutex,
		state:       state,
	}
}

func (c *Controller) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.syncToStdout(key.(string))
	c.handleErr(err, key)
	return true
}

func (c *Controller) syncToStdout(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		log(fmt.Sprintf("%s: fetching object with key %s from store failed with %v", au.Red(au.Bold("ERROR")), key, err))
		return err
	}

	if !exists {
		log(fmt.Sprintf("%s: namespace %s deleted", au.Cyan(au.Bold("INFO")), key))
		c.mutex.Lock()
		delete(c.state, key)
		c.mutex.Unlock()
	} else {
		name := obj.(*v1.Namespace).GetName()
		label := obj.(*v1.Namespace).ObjectMeta.Labels[fmt.Sprintf("%s/%s", labelPrefix, labelNameGroup)]

		if len(label) > 0 {
			c.mutex.Lock()
			log(fmt.Sprintf("%s: processing namespace %s (label %s)", au.Cyan(au.Bold("INFO")), name, au.Bold(label)))
			c.state[name] = label
			c.mutex.Unlock()
		}
	}
	if c.queue.Len() == 0 {
		c.mutex.Lock()
		state := c.state
		c.mutex.Unlock()
		log(fmt.Sprintf("%s: multitenant state is as follows: %v", au.Cyan(au.Bold("INFO")), au.Bold(state)))
		err = apply(c.state)
		if err != nil {
			log(fmt.Sprintf("%s: can't apply state %v: %v", au.Red(au.Bold("ERROR")), state, err))
		}

	}
	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < 5 {
		log(fmt.Sprintf("%s: can't sync namespace %v: %v", au.Red(au.Bold("ERROR")), key, err))
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	runtime.HandleError(err)
	log(fmt.Sprintf("%s: dropping namespace %q out of the queue: %v", au.Cyan(au.Bold("INFO")), key, err))
}

// Run manages the controller lifecycle
func (c *Controller) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	defer c.queue.ShutDown()
	log(fmt.Sprintf("%s: starting namespace controller", au.Cyan(au.Bold("INFO"))))

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	log(fmt.Sprintf("%s: stopping namespace controller", au.Cyan(au.Bold("INFO"))))
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func main() {
	var kubeconfig string
	var master string

	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "", "master url")
	flag.Parse()

	if len(kubeconfig) == 0 {
		kubeconfig = os.Getenv("KUBECONFIG")
	}

	var config *rest.Config
	var configError error

	if len(kubeconfig) > 0 {
		config, configError = clientcmd.BuildConfigFromFlags(master, kubeconfig)
		if configError != nil {
			log(fmt.Sprintf("%s: %s", au.Bold(au.Red("ERROR")), configError))
			return
		}
	} else {
		config, configError = rest.InClusterConfig()
		if configError != nil {
			log(fmt.Sprintf("%s: %s", au.Bold(au.Red("ERROR")), configError))
			return
		}
	}

	// creates the connection
	config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		log(fmt.Sprintf("%s: %s", au.Bold(au.Red("ERROR")), err))
		return
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log(fmt.Sprintf("%s: %s", au.Bold(au.Red("ERROR")), err))
		return
	}

	var mutex = &sync.Mutex{}
	var state = map[string]string{}

	namespaceListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "namespaces", "", fields.Everything())

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	indexer, informer := cache.NewIndexerInformer(namespaceListWatcher, &v1.Namespace{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	}, cache.Indexers{})

	controller := NewController(queue, indexer, informer, clientset, mutex, state)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	select {}
}

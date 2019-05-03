package main

import (
	"fmt"
	kubernetes "k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	au "github.com/logrusorgru/aurora"
)

func prunePolicies(clientset kubernetes.Interface, state map[string]string) error {
	namespaceMap := map[string]int{}
	for namespace := range state {
		namespaceMap[namespace] = 1
	}

	// policies
	policies, err := clientset.NetworkingV1().NetworkPolicies("").List(metav1.ListOptions{})

	if err != nil {
		return err
	}

	ownerLabel := fmt.Sprintf("%s/%s", labelPrefix, labelNameOwner)
	for _, policy := range policies.Items {
		name := policy.GetObjectMeta().GetName()
		namespace := policy.GetObjectMeta().GetNamespace()
		owner := policy.GetObjectMeta().GetLabels()[ownerLabel]

		// skip if not one of ours
		if name != "multitenant" || owner != "multitenant-controller" {
			continue
		}

		// also skip if namespace in map of managed namespaces
		_, ok := namespaceMap[namespace]
		if ok {
			continue
		}

		// delete if object is one of ours, but not in a managed namespace
		deletePolicy := metav1.DeletePropagationForeground
		err = clientset.NetworkingV1().NetworkPolicies(namespace).Delete(name, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
		if err != nil {
			return err
		}
		log(fmt.Sprintf("%s: deleted NetworkPolicy %s in namespace %s", au.Bold(au.Red("WARN")), au.Bold(name), au.Bold(namespace)))
	}

	return nil
}
package main

import (
	"encoding/json"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/networking/v1"
	au "github.com/logrusorgru/aurora"
)

func applyPolicy(clientset kubernetes.Interface, policy string) error {
	b := []byte(policy)
	var obj v1.NetworkPolicy
	err := json.Unmarshal(b, &obj) //TODO: error handling
	if err != nil {
		return err
	}
	name := obj.GetObjectMeta().GetName()
	namespace := obj.GetObjectMeta().GetNamespace()

	_, err = clientset.NetworkingV1().NetworkPolicies(namespace).Create(&obj)
	if err == nil {
		log(fmt.Sprintf("%s: created network policy %s", au.Bold(au.Cyan("INFO")), au.Bold(name)))
		return nil
	}

	//attempt Update
	_, err = clientset.NetworkingV1().NetworkPolicies(namespace).Update(&obj)
	if err != nil {
		return err
	}
	log(fmt.Sprintf("%s: updated network policy %s", au.Bold(au.Cyan("INFO")), au.Bold(name)))
	return nil
}
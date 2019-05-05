package main

import (
	kubernetes "k8s.io/client-go/kubernetes"
)

func apply(clientset kubernetes.Interface, state map[string]string) error {
    policies, err := generatePolicies(state)
    if err != nil {
        return err
	}

	for _, policy := range policies {
		err = applyPolicy(clientset, policy)
		if err != nil {
			return err
		}
	}

	err = prunePolicies(clientset, state)
	if err != nil {
		return err
	}

	return nil
}

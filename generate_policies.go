package main

import (
	"fmt"
	"sort"
	"strings"
)

func generatePolicies(state map[string]string) ([]string, error) {

	// iterate over namespaces/map keys
	var groups []string
	var policies []string

	for _, group := range state {
		groups = append(groups, fmt.Sprintf("\"%s\"", group))
	}

	// if group is "global", permit ingress from all groups
	// if it isn't, permit ingress from own group and "global"
	for namespace, group := range state {
		var ingressGroups []string
		if group == "global" {
			ingressGroups = groups
		} else {
			ingressGroups = append(ingressGroups, fmt.Sprintf("\"%s\"", group), "\"global\"")
		}
		policy, err := generatePolicy(namespace, group, ingressGroups)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	return policies, nil
}

func generatePolicy(namespace string, group string, groups []string) (string, error) {
	manifest := fmt.Sprintf(`{
	"apiVersion": "networking.k8s.io/v1",
	"kind": "NetworkPolicy",
	"metadata": {
		"name": "multitenant",
		"namespace": "%s"
	},
	"spec": {
		"policyTypes": [ "Ingress" ],
		"ingress": [
			"from": [
				"namespaceSelector": {
					"matchExpressions": [{
						"key": "%s/%s",
						"operator": "In",
						"in": [ %s ]
					}]
				}
			]
		]	
	},
}`, namespace, labelPrefix, labelName, arrayToCSV(groups))
	return manifest, nil
}

func arrayToCSV(values []string) string {
	values = unique(values)
	sort.Strings(values)
	return strings.Join(values, ", ")
}

func unique(slice []string) []string {
	keys := make(map[string]struct{})
	list := []string{}
	for _, entry := range slice {
			if _, ok := keys[entry]; !ok {
					keys[entry] = struct{}{}
					list = append(list, entry)
			}
	}
	return list
}
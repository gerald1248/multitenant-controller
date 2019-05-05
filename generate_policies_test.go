package main

import (
	"encoding/json"
	"reflect"
	"testing"
	"k8s.io/api/networking/v1"
)

func TestGeneratePolicies(t *testing.T) {
	var tests = []struct {
		description            string
		input                  map[string]string
		expectedSelectorValues []string
	}{
		{"minimal", map[string]string{"namespaceA":"minimal"},[]string{"global", "minimal"}},
		{"frontend", map[string]string{"namespaceA":"frontend","namespaceB":"backend"},[]string{"frontend", "global"}},
		{"global", map[string]string{"namespaceGlobal":"global","namespaceA":"frontend"},[]string{"frontend", "global"}},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			output, _ := generatePolicies(test.input)

			var obj v1.NetworkPolicy
			b := []byte(output[0])
			err := json.Unmarshal(b, &obj)
			if err != nil {
				t.Errorf("Malformed JSON: %s", err.Error())
				return
			}

			selectorValues := obj.Spec.Ingress[0].From[0].NamespaceSelector.MatchExpressions[0].Values

			if !reflect.DeepEqual(selectorValues, test.expectedSelectorValues) {
				t.Errorf("Unexpected selector values for input '%s' - got '%s'; expected '%s'", test.input, selectorValues, test.expectedSelectorValues)
			}
		})
	}
}

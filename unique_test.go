package main

import (
	"reflect"
	"testing"
)

func TestUnique(t *testing.T) {
	var tests = []struct {
		description     string
		input           []string
		expected        []string
	}{
		{"unique_empty", []string{}, []string{}},
		{"unique_identity", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"unique_deduplicate", []string{"a", "a", "b", "c"}, []string{"a", "b", "c"}},
		{"unique_deduplicate_unsorted", []string{"c", "b", "a", "a"}, []string{"c", "b", "a"}},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			output := unique(test.input)
			if !reflect.DeepEqual(output, test.expected) {
				t.Errorf("Unexpected output %v from input %v", output, test.expected)
			}
		})
	}
}



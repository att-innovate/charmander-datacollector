package main

import (
	"reflect"
	"testing"
)

func TestExtractMachineReadable(t *testing.T) {
	var expected, result []string
	var mr bool

	// Not
	result = []string{"foo", "bar", "baz"}

	expected = []string{"foo", "bar", "baz"}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("bad: %#v", result)
	}

	if mr {
		t.Fatal("should not be mr")
	}

}
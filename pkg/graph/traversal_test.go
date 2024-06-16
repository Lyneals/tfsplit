package graph

// Test suite for the graph.traversal package

import (
	"slices"
	"testing"

	"github.com/awalterschulze/gographviz"
)

func TestGetChildrenBasic(t *testing.T) {
	graphString := `digraph G {
		1 -> 2;
		1 -> 3;
		2 -> 4;
		3 -> 4;
		3 -> 5;
		4 -> 6;
		5 -> 6;
	}`
	graph, _ := gographviz.Read([]byte(graphString))
	rootNodes := []string{"1", "3"}
	children := GetChildren("3", graph, rootNodes)
	expected := []string{"4", "5", "6"}
	if len(children) != len(expected) {
		t.Errorf("Expected %v, got %v", expected, children)
	}
	for _, v := range children {
		if !slices.Contains(expected, v) {
			t.Errorf("Expected %v, got %v", expected, children)
		}
	}
}

func TestRemoveDuplicateStr(t *testing.T) {
	strSlice := []string{"1", "2", "3", "1", "2", "3", "4"}
	expected := []string{"1", "2", "3", "4"}
	result := removeDuplicateStr(strSlice)
	if len(result) != len(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
	for _, v := range result {
		if !slices.Contains(expected, v) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	}
}

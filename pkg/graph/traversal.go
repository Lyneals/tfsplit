package graph

import (
	"github.com/awalterschulze/gographviz"
)

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func GetChildrenHelper(node string, graph *gographviz.Graph, visited map[string]bool) []string {
	children := []string{}
	if visited[node] {
		return children
	}
	visited[node] = true
	mEdges := graph.Edges.SrcToDsts[node]
	for _, edges := range mEdges {
		for _, edge := range edges {
			if visited[edge.Dst] {
				continue
			}
			children = append(children, edge.Dst)
			children = append(children, GetChildrenHelper(edge.Dst, graph, visited)...)
		}
	}
	return removeDuplicateStr(children)
}

func GetChildren(node string, graph *gographviz.Graph, rootNodes []string) []string {
	visited := make(map[string]bool)
	// Mark root nodes as visited
	// Will avoid discovering node that belong to another layer
	for _, n := range rootNodes {
		if n == node {
			continue
		}
		visited[n] = true
	}
	return GetChildrenHelper(node, graph, visited)
}

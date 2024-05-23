package graph

import (
	"github.com/awalterschulze/gographviz"
)

func GetChildrenHelper(node string, graph *gographviz.Graph, visited map[string]bool) []string {
	children := []string{}
	if visited[node] {
		return children
	}
	visited[node] = true
	mEdges := graph.Edges.SrcToDsts[node]
	for _, edges := range mEdges {
		for _, edge := range edges {
			children = append(children, edge.Dst)
			children = append(children, GetChildrenHelper(edge.Dst, graph, visited)...)
		}
	}
	return children
}

func GetChildren(node string, graph *gographviz.Graph, rootNodes []string) []string {
	visited := make(map[string]bool)
	// Mark root nodes as visited
	// Necessary this will avoid discovering node that belong to another layer
	for _, n := range rootNodes {
		if n == node {
			continue
		}
		visited[n] = true
	}
	return GetChildrenHelper(node, graph, visited)
}

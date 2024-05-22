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

func GetChildren(node string, graph *gographviz.Graph) []string {
	visited := make(map[string]bool)
	return GetChildrenHelper(node, graph, visited)
}

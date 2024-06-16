package graph

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/awalterschulze/gographviz"
)

func fixName(str string) string {

	name := strings.Split(str, " ")[1]
	// If the node name is a module, keep only the module named part
	if strings.HasPrefix(name, "module.") {
		name = "module." + strings.Split(name, ".")[1]
	}

	if strings.HasPrefix(name, "provider") {
		name, _ = strings.CutPrefix(name, "provider[\\\"registry.terraform.io/hashicorp/")
		kind := strings.Split(name, "\\")[0]

		if strings.HasSuffix(name, "\"") {
			name = name[:len(name)-1]
		}

		if strings.HasSuffix(name, "]") {
			name = "provider." + kind
		} else {
			spl := strings.Split(name, ".")
			alias := spl[len(spl)-1]
			name = "provider." + kind + alias
		}
		slog.Debug(
			"fixName.provider",
			"node", name,
		)
	}

	if strings.HasSuffix(name, "\"") {
		name = name[:len(name)-1]
	}

	return fmt.Sprintf("%s", name)
}

func shouldKeepEdge(edge *gographviz.Edge) bool {
	if edge.Src == edge.Dst {
		return false
	}
	return true
}

func pruneEdges(graph *gographviz.Graph, graphNg *gographviz.Graph) {
	seen := make(map[string]bool)
	for _, edge := range graph.Edges.Edges {
		edge.Src = fixName(edge.Src)
		edge.Dst = fixName(edge.Dst)
		key := fmt.Sprintf("%s%s", edge.Src, edge.Dst)
		if shouldKeepEdge(edge) && seen[key] == false {
			seen[fmt.Sprintf("%s%s", edge.Src, edge.Dst)] = true
			graphNg.Edges.Add(edge)
		}
	}
}

func rewriteNodes(graph *gographviz.Graph, graphNg *gographviz.Graph) {
	for _, node := range graph.Nodes.Nodes {
		node.Name = fixName(node.Name)
		node.Attrs.Add("label", node.Name)
		graphNg.Nodes.Add(node)
	}
}

func reworkGraph(graph *gographviz.Graph) *gographviz.Graph {
	gographNg := gographviz.NewGraph()
	gographNg.SetName("G")
	gographNg.SetDir(true)
	gographNg.AddAttr("G", "newrank", "true")
	gographNg.AddAttr("G", "compound", "true")

	pruneEdges(graph, gographNg)
	rewriteNodes(graph, gographNg)

	return gographNg
}

func LoadGraph(graph string) (*gographviz.Graph, error) {
	graphAst, _ := gographviz.ParseString(graph)

	gograph := gographviz.NewGraph()
	if err := gographviz.Analyse(graphAst, gograph); err != nil {
		return nil, fmt.Errorf("failed to parse graph: %w", err)
	}

	return reworkGraph(gograph), nil
}

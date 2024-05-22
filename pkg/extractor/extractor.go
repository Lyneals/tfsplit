package extractor

import (
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
)

func GetIds(name string, state *tfjson.State) map[string]string {
	m := make(map[string]string)
	all := state.Values.RootModule.Resources
	for _, child := range state.Values.RootModule.ChildModules {
		all = append(all, child.Resources...)
	}
	for _, r := range all {
		if r.Mode == "data" {
			continue
		}
		// Will create conflict if there are multiple resources with the same prefix
		// eg. aws_instance.foo and aws_instance.foo_bar
		// Might fix later, idk
		if strings.HasPrefix(r.Address, name) {
			m[r.Address] = r.AttributeValues["id"].(string)
		}
	}
	return m
}

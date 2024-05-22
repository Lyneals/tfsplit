package writer

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const (
	IMPORT_FILENAME = "tfsplit_imports.tf"
	HCL_IMPORT      = `variable "tfsplit_imports" {
  type 			= map(string)
  description 	= "Resources to import"
  default 		= {}
}

import {
  for_each = var.tfsplit_imports

  to = each.key
  id = each.value
}
`
)

func WriteLayer(path string, nodes []string, hcl map[string]map[string]string) {
	slog.Debug(
		"WriteLayer",
		"path", path,
		"nodes", nodes,
		"hcl", hcl,
	)

	// If directory already exists, return
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		slog.Info(
			"Directory already exist, assuming the layer is already written. Skipping.",
			"path", path,
		)
		return
	}

	err := os.MkdirAll(path, os.ModePerm)

	if err != nil {
		panic(err)
	}

	var main strings.Builder
	var data strings.Builder
	var variables strings.Builder
	var outputs strings.Builder
	var providers strings.Builder

	for _, node := range nodes {
		if strings.HasPrefix(node, "module.") {
			name, _ := strings.CutPrefix(node, "module.")
			main.WriteString(hcl["modules"][name] + "\n")
		}
		if strings.HasPrefix(node, "resource.") {
			name, _ := strings.CutPrefix(node, "resource.")
			main.WriteString(hcl["resources"][name] + "\n")
		}
		if strings.HasPrefix(node, "data.") {
			name, _ := strings.CutPrefix(node, "data.")
			data.WriteString(hcl["data"][name] + "\n")
		}
		if strings.HasPrefix(node, "var.") {
			name, _ := strings.CutPrefix(node, "var.")
			variables.WriteString(hcl["variables"][name] + "\n")
		}
		if strings.HasPrefix(node, "output.") {
			name, _ := strings.CutPrefix(node, "output.")
			outputs.WriteString(hcl["outputs"][name] + "\n")
		}
		if strings.HasPrefix(node, "provider") {
			name, _ := strings.CutPrefix(node, "provider.")
			providers.WriteString(hcl["providers"][name] + "\n")
		}
	}

	os.WriteFile(filepath.Join(path, "main.tf"), []byte(main.String()), 0644)
	os.WriteFile(filepath.Join(path, "data.tf"), []byte(data.String()), 0644)
	os.WriteFile(filepath.Join(path, "variables.tf"), []byte(variables.String()), 0644)
	os.WriteFile(filepath.Join(path, "outputs.tf"), []byte(outputs.String()), 0644)
	os.WriteFile(filepath.Join(path, "providers.tf"), []byte(providers.String()), 0644)
	os.WriteFile(filepath.Join(path, IMPORT_FILENAME), []byte(HCL_IMPORT), 0644)
}

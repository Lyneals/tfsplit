package writer

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"tfsplit/pkg/terraform"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/exp/maps"
)

func WriteLayer(path string, nodes []string, hcl map[string]map[string]string, layerName string) {
	slog.Debug(
		"WriteLayer",
		"path", path,
		"nodes", nodes,
		"hcl", hcl,
	)

	basePath := filepath.Join(path, "tfsplit", layerName)

	// If directory already exists, return
	if _, err := os.Stat(basePath); !os.IsNotExist(err) {
		slog.Info(
			"Directory already exist, assuming the layer is already written. Skipping.",
			"path", basePath,
		)
		return
	}

	err := os.MkdirAll(basePath, os.ModePerm)

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
		/* Don't handle providers for now, alias make it difficult
		if strings.HasPrefix(node, "provider") {
			name, _ := strings.CutPrefix(node, "provider.")
			providers.WriteString(hcl["providers"][name] + "\n")
		}
		*/
	}

	providers.WriteString(strings.Join(maps.Values(hcl["providers"]), "/n"))

	if main.Len() != 0 {
		os.WriteFile(filepath.Join(basePath, "main.tf"), []byte(main.String()), 0644)
	}
	if data.Len() != 0 {
		os.WriteFile(filepath.Join(basePath, "data.tf"), []byte(data.String()), 0644)
	}
	if variables.Len() != 0 {
		os.WriteFile(filepath.Join(basePath, "variables.tf"), []byte(variables.String()), 0644)
	}
	if outputs.Len() != 0 {
		os.WriteFile(filepath.Join(basePath, "outputs.tf"), []byte(outputs.String()), 0644)
	}
	if providers.Len() != 0 {
		os.WriteFile(filepath.Join(basePath, "providers.tf"), []byte(providers.String()), 0644)
	}

	os.WriteFile(filepath.Join(basePath, "terraform.tf"), []byte(hcl["terraform"]["root"]), 0644)
}

func WriteVars(path string, fileName string, nodes []string, imports map[string]string, layerName string) {
	vars, err := terraform.ReadTfvars(filepath.Join(path, fileName))
	if err != nil {
		panic(fmt.Errorf("Failed to read var file: %s", err))
	}

	slog.Debug(
		"WriteVars",
		"path", path,
		"filename", fileName,
		"nodes", nodes,
		"vars", vars,
	)

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	// Build variables files with required variables value
	for _, node := range nodes {
		if strings.HasPrefix(node, "var.") {
			name, _ := strings.CutPrefix(node, "var.")
			rootBody.SetAttributeValue(name, vars[name])
		}
	}

	// Build imports map
	importNeeded := make(map[string]bool)
	for _, node := range nodes {
		if strings.HasPrefix(node, "resource.") || strings.HasPrefix(node, "module.") {
			importNeeded[node] = true
		}
	}

	// Get only the imports needed
	mImports := make(map[string]cty.Value)
	for k, v := range imports {
		slog.Debug(
			"Filter imports",
			"key", k,
		)
		sp := strings.Split(k, ".")
		t := sp[0]
		n := strings.Split(sp[1], "[")[0]
		if !importNeeded[t+"."+n] {
			continue
		}
		mImports[k] = cty.StringVal(v)
	}

	varsPath := filepath.Join(path, "tfsplit", layerName, fileName)
	dir := filepath.Dir(varsPath)

	os.MkdirAll(dir, os.ModePerm)

	err = os.WriteFile(varsPath, f.Bytes(), 0644)
	if err != nil {
		panic(err)
	}
}

func WriteBackendConfig(path string, fileName string, layerName string) {
	backend, err := terraform.ReadTfvars(filepath.Join(path, fileName))
	if err != nil {
		panic(fmt.Errorf("Failed to read backend config file: %s", err))
	}

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	backend["key"] = cty.StringVal(fmt.Sprintf("%s/%s", layerName, backend["key"].AsString()))
	for k, v := range backend {
		rootBody.SetAttributeValue(k, v)
	}

	varsPath := filepath.Join(path, "tfsplit", layerName, fileName)
	dir := filepath.Dir(varsPath)

	os.MkdirAll(dir, os.ModePerm)

	err = os.WriteFile(varsPath, f.Bytes(), 0644)
	if err != nil {
		panic(err)
	}
}

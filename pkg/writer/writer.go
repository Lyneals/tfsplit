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

var (
	mFileName = map[string]string{
		"resource": "main.tf",
		"module":   "main.tf",
		"output":   "outputs.tf",
		"var":      "variables.tf",
		"provider": "providers.tf",
		"local":    "locals.tf",
	}
	kindKeys = [8]string{"data", "module", "output", "var", "provider", "resource", "provider", "local"}
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

	m := make(map[string]*strings.Builder)
	for _, kind := range kindKeys {
		m[kind] = &strings.Builder{}
	}

	seen := make(map[string]bool)
	for _, node := range nodes {
		if strings.HasPrefix(node, "provider") {
			continue
		}

		split := strings.Split(node, ".")
		kind := split[0]
		name := strings.Join(split[1:], ".")

		if seen[kind+name] {
			continue
		}

		slog.Debug(
			"WriteLayer",
			"kind", kind,
			"name", name,
		)

		if kind == "local" {
			continue
		}

		// Handle resources
		if m[kind] == nil {
			kind = "resource"
			name = node
		}

		if hcl[kind][name] == "" {
			panic(fmt.Errorf("Resource %s %s not found in HCL", kind, name))
		}
		m[kind].WriteString(hcl[kind][name] + "\n")
		seen[kind+name] = true
	}
	m["provider"].WriteString(strings.Join(maps.Values(hcl["providers"]), "/n"))

	slog.Debug(
		"WriteLayer",
		"m", m,
	)

	for name, sb := range m {
		os.WriteFile(filepath.Join(basePath, mFileName[name]), []byte(sb.String()), 0644)
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

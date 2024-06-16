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
)

var (
	mFileName = map[string]string{
		"resource":  "main.tf",
		"module":    "main.tf",
		"data":      "data.tf",
		"terraform": "providers.tf",
		"output":    "outputs.tf",
		"variable":  "variables.tf",
		"provider":  "providers.tf",
		"locals":    "locals.tf",
	}
	kindKeys = [10]string{"data", "module", "output", "var", "provider", "resource", "provider", "locals", "terraform", "variable"}
)

func WriteLayer(path string, nodes []string, hcl map[string]map[string]interface{}, layerName string) {
	basePath := filepath.Join(path, "tfsplit", layerName)
	slog.Debug(
		"WriteLayer",
		"path", basePath,
		"nodes", nodes,
	)

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

	for _, node := range nodes {

		if strings.HasPrefix(node, "data.") {
			k := strings.Split(node, ".")[0]
			t := strings.Split(node, ".")[1]
			name := strings.Join(strings.Split(node, ".")[2:], ".")

			slog.Debug(
				"Writing data",
				"kind", k,
				"type", t,
				"name", name,
			)

			m[k].WriteString(hcl[k][t].(map[string]string)[name])
		} else if strings.HasPrefix(node, "module.") || strings.HasPrefix(node, "output.") {
			k := strings.Split(node, ".")[0]
			name := strings.Split(node, ".")[1]

			slog.Debug(
				"Writing module/output/var",
				"kind", k,
				"name", name,
			)

			m[k].WriteString(hcl[k][name].(string))
		} else if strings.HasPrefix(node, "var.") {
			k := "variable"
			name := strings.Split(node, ".")[1]

			slog.Debug(
				"Writing variable",
				"kind", k,
				"name", name,
			)

			m[k].WriteString(hcl[k][name].(string))
		} else if strings.HasPrefix(node, "local.") {
			k := "locals"
			name := strings.Split(node, ".")[1]

			slog.Debug(
				"Writing local",
				"kind", k,
				"name", name,
			)

			m[k].WriteString(hcl[k][name].(string))
		} else if strings.HasPrefix(node, "provider") {
			k := "provider"
			t := strings.Split(node, ".")[1]

			slog.Debug(
				"Writing provider",
				"type", t,
			)

			if hcl[k][t] == nil {
				slog.Info(
					"WriteLayer.provider",
					"provider", t,
					"message", "provider not found - can be ignored if provider doesn't require configuration",
				)
				continue
			}
			m[k].WriteString(hcl[k][t].(string))
		} else {
			k := "resource"
			t := strings.Split(node, ".")[0]
			name := strings.Join(strings.Split(node, ".")[1:], ".")

			slog.Debug(
				"Writing resource",
				"kind", k,
				"type", t,
				"name", name,
			)

			m[k].WriteString(hcl[k][t].(map[string]string)[name])
		}
	}

	if m["locals"].Len() > 0 {
		var newSb strings.Builder
		newSb.WriteString(fmt.Sprintf(`locals {
			%s
		}`, m["locals"].String()))
		m["locals"] = &newSb
	}

	for name, sb := range m {
		f, _ := os.OpenFile(filepath.Join(basePath, mFileName[name]), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		f.WriteString(sb.String())
		f.Close()
	}
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

	/*
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
		}	*/

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

package terraform

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

var (
	KEYS = [7]string{"resources", "data", "modules", "outputs", "variables", "providers", "terraform"}
)

func providerParser(path string) map[string]string {
	m, err := filepath.Glob(path + "/*.tf")

	if err != nil {
		panic(err)
	}

	res := make(map[string]string)

	for _, file := range m {
		dat, _ := os.ReadFile(file)
		lines := strings.Split(string(dat), "\n")

		for i, line := range lines {
			if strings.HasPrefix(line, "provider") {
				name := strings.Split(line, " ")[1]
				unq, _ := strconv.Unquote(name)
				res[unq] = customParser(file, i+1)
			}
		}
	}
	return res
}

func terraformParser(path string) string {
	m, err := filepath.Glob(path + "/*.tf")

	if err != nil {
		panic(err)
	}

	var res strings.Builder

	for _, file := range m {
		dat, _ := os.ReadFile(file)
		lines := strings.Split(string(dat), "\n")

		for i, line := range lines {
			if strings.HasPrefix(line, "terraform") {
				res.WriteString(customParser(file, i+1) + "\n")
			}
		}
	}
	return res.String()
}

func OpenTerraformFiles(path string) *tfconfig.Module {
	module, _ := tfconfig.LoadModule(path)
	return module
}

func FromConfig(module *tfconfig.Module) map[string]map[string]string {
	slog.Debug(
		"FromConfig",
		"module", fmt.Sprintf("%+v", module),
	)
	hclCode := make(map[string]map[string]string)
	for _, key := range KEYS {
		hclCode[key] = make(map[string]string)
	}

	for _, item := range module.ManagedResources {
		hclCode["resources"][item.Name] = customParser(item.Pos.Filename, item.Pos.Line)
	}
	for _, item := range module.DataResources {
		hclCode["data"][item.Name] = customParser(item.Pos.Filename, item.Pos.Line)
	}
	for _, item := range module.ModuleCalls {
		hclCode["modules"][item.Name] = customParser(item.Pos.Filename, item.Pos.Line)
	}
	for _, item := range module.Outputs {
		hclCode["outputs"][item.Name] = customParser(item.Pos.Filename, item.Pos.Line)
	}
	for _, item := range module.Variables {
		hclCode["variables"][item.Name] = customParser(item.Pos.Filename, item.Pos.Line)
	}
	for _, item := range providerParser(module.Path) {
		hclCode["providers"][item] = item
	}
	hclCode["terraform"]["root"] = terraformParser(module.Path)
	return hclCode
}

func customParser(path string, pos int) string {
	dat, _ := os.ReadFile(path)
	lines := strings.Split(string(dat), "\n")
	var sb strings.Builder

	for i := pos - 1; i < len(lines); i++ {
		sb.WriteString(lines[i] + "\n")
		if strings.HasPrefix(lines[i], "}") {
			break
		}
	}
	return sb.String()
}

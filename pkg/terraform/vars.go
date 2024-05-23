package terraform

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func ReadTfvars(path string) (map[string]cty.Value, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	file, diags := hclsyntax.ParseConfig(content, "", hcl.InitialPos)
	if diags.HasErrors() {
		return nil, diags
	}

	attrs, diags := file.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}

	vals := make(map[string]cty.Value, len(attrs))
	for name, attr := range attrs {
		vals[name], diags = attr.Expr.Value(nil)
		if diags.HasErrors() {
			return nil, diags
		}
	}
	return vals, nil
}

func ReadTfModule(path string) (*hcl.File, error) {
	content, err := filepath.Glob(path + "/*.tf")
	if err != nil {
		return nil, err
	}

	var acc strings.Builder
	for _, m := range content {
		content, err := os.ReadFile(m)
		if err != nil {
			return nil, err
		}
		acc.WriteString(string(content))
	}

	file, diags := hclsyntax.ParseConfig([]byte(acc.String()), "", hcl.InitialPos)
	if diags.HasErrors() {
		return nil, diags
	}

	return file, nil
}

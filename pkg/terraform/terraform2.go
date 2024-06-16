package terraform

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// Why the fuck do I even need this
func getClosingBrace(start int, f []byte) int {
	count := 0
	for i := start; i < len(f); i++ {
		switch f[i] {
		case '{':
			count++
		case '}':
			count--
		}
		if count == 0 {
			return i
		}
	}

	return start
}

func ParseFolder(folder string) (map[string]map[string]string, error) {
	result := map[string]map[string]string{
		"data":     {},
		"resource": {},
		"module":   {},
		"locals":   {},
		"provider": {},
	}

	parser := hclparse.NewParser()

	m, _ := filepath.Glob(folder + "/*.tf")
	for _, file := range m {
		tfFile, _ := parser.ParseHCLFile(file)
		src, err := os.ReadFile(file)

		if err != nil {
			return result, fmt.Errorf("failed to read %s: %s", file, err)
		}

		content, _, diags := tfFile.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{Type: "data", LabelNames: []string{"type", "name"}},
				{Type: "resource", LabelNames: []string{"type", "name"}},
				{Type: "module", LabelNames: []string{"name"}},
				{Type: "locals"},
				{Type: "provider", LabelNames: []string{"type"}},
			},
		})

		if diags.HasErrors() {
			slog.Debug(
				"terraform.ParseFolder",
				"diags", diags,
			)
		}

		for _, block := range content.Blocks {
			blockType := block.Type

			switch blockType {
			case "data", "resource":
				blockKind := block.Labels[0]
				blockName := block.Labels[1]
				result[blockType][blockKind+blockName] = string(src[block.DefRange.Start.Byte:getClosingBrace(block.DefRange.End.Byte+1, src)])
			case "provider":
				blockName := block.Labels[0]
				slog.Debug(
					"ParseFolder",
					"provider", block,
				)
				// Get alias if it exists as attribute
				attributes, _ := block.Body.JustAttributes()
				if alias, ok := attributes["alias"]; ok {
					// Should always be static, so we just get the value and ignore diags
					v, _ := alias.Expr.Value(nil)
					blockName += v.AsString()
				}
				result[blockType][blockName] = string(src[block.DefRange.Start.Byte:getClosingBrace(block.DefRange.End.Byte+1, src)])
			case "module":
				blockName := block.Labels[0]

				result[blockType][blockName] = string(src[block.DefRange.Start.Byte:getClosingBrace(block.DefRange.End.Byte+1, src)])
			case "locals":
				attributes, _ := block.Body.JustAttributes()

				for name, attr := range attributes {
					result[blockType][name] = string(src[attr.Range.Start.Byte:attr.Range.End.Byte])
				}
			default:
				slog.Warn(
					"ParseFolder",
					"unsuported block type: %s", blockType,
					"block", block,
				)
				continue
			}
		}

	}

	slog.Debug(
		"terraform.ParseFolder",
		"result", result,
	)

	return result, nil
}
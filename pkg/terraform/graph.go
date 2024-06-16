// Based on https://github.com/busser/tfautomv/blob/main/pkg/terraform/plan.go
package terraform

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-exec/tfexec"
)

func GetGraph(ctx context.Context, path string, tfBin string, backendConfig string) (string, error) {
	tf, err := tfexec.NewTerraform(path, tfBin)
	if err != nil {
		return "", fmt.Errorf("failed to create Terraform executor: %w", err)
	}

	err = tf.Init(ctx, tfexec.BackendConfig(backendConfig))
	if err != nil {
		return "", fmt.Errorf("failed to initialize Terraform: %w", err)
	}

	graph, err := tf.Graph(ctx, tfexec.DrawCycles(true), tfexec.GraphType("plan-refresh-only"))
	if err != nil {
		return "", fmt.Errorf("failed to compute Terraform graph: %w", err)
	}

	return graph, nil
}

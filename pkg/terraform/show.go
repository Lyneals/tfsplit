// Based on https://github.com/busser/tfautomv/blob/main/pkg/terraform/plan.go
package terraform

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

func GetState(ctx context.Context, path string, tfBin string) (*tfjson.State, error) {
	tf, err := tfexec.NewTerraform(path, tfBin)
	if err != nil {
		return nil, fmt.Errorf("failed to create Terraform executor: %w", err)
	}

	err = tf.Init(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Terraform: %w", err)
	}

	state, err := tf.Show(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve terraform state: %w", err)
	}

	return state, nil
}

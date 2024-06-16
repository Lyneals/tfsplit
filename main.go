package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"slices"
	"strings"

	"tfsplit/pkg/config"
	"tfsplit/pkg/extractor"
	"tfsplit/pkg/graph"
	"tfsplit/pkg/logger"
	"tfsplit/pkg/terraform"
	"tfsplit/pkg/writer"

	"github.com/urfave/cli/v2"
)

func main() {
	opts := logger.PrettyHandlerOptions{
		SlogOpts: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}
	handler := logger.NewPrettyHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(handler))

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Usage:    "Path to the tfsplit configuration file",
				Aliases:  []string{"c"},
				FilePath: "./.tfsplit.yaml",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "path",
				Usage:    "Path to the Terraform module",
				Aliases:  []string{"p"},
				Value:    "./",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "backend-config",
				Usage: "Terraform init backend config",
			},
			&cli.StringFlag{
				Name:  "var-file",
				Usage: "Terraform plan var-file",
			},
		},
		Name:   "tfsplit",
		Usage:  "Read the Terraform module and split it into smaller layers based on the dependencies between the resources",
		Action: appHandler,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func appHandler(c *cli.Context) error {
	slog.Info("Starting tfsplit")

	ctx := context.Background()

	config_path := c.String("config")
	config, err := config.LoadConfig(config_path)
	if err != nil {
		return fmt.Errorf("Failed to load config: %s", err)
	}
	slog.Debug(
		"Config loaded",
		"config", config,
	)

	tfPath := c.String("path")
	backendConfig := c.String("backend-config")
	varFile := c.String("var-file")

	tfGraph, err := terraform.GetGraph(ctx, tfPath, "terraform", backendConfig)
	if err != nil {
		return fmt.Errorf("Failed to get graph: %s", err)
	}

	gograph, err := graph.LoadGraph(tfGraph)
	if err != nil {
		return fmt.Errorf("Failed to load graph: %s", err)
	}

	hclCode, err := terraform.ParseFolder(tfPath)

	if err != nil {
		return fmt.Errorf("Failed to parse folder: %s", err)

	}

	state, err := terraform.GetState(ctx, tfPath, "terraform", backendConfig)
	if err != nil {
		return fmt.Errorf("Failed to get state: %s", err)
	}

	// Build usedNodes list
	var usedNodes []string
	for _, layer := range config.Layers {
		usedNodes = append(usedNodes, layer.RootNode)
	}

	// For each layer, get the children of the root node
	// extract the ids of the resources
	// write the layer
	// write the vars
	for _, layer := range config.Layers {
		childs := graph.GetChildren(layer.RootNode, gograph, usedNodes)
		for _, child := range childs {
			if strings.HasPrefix(child, "data") || strings.HasPrefix(child, "provider") || strings.HasPrefix(child, "local") || strings.HasPrefix(child, "var") {
				continue
			}
			usedNodes = append(usedNodes, child)
		}
		ids := extractor.GetIds(layer.RootNode, state)

		requiredNodes := append(childs, layer.RootNode)
		slices.Sort(requiredNodes)

		slog.Debug(
			"Processing layer",
			"layer", layer.Name,
			"rootNode", layer.RootNode,
			"ids", ids,
			"childs", childs,
		)

		writer.WriteLayer(tfPath, requiredNodes, hclCode, layer.Name)

		if varFile != "" {
			writer.WriteVars(tfPath, varFile, requiredNodes, ids, layer.Name)
		}
		if backendConfig != "" {
			writer.WriteBackendConfig(tfPath, backendConfig, layer.Name)
		}
	}

	return nil
}

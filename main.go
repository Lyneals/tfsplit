package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

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

	tf_path := c.String("path")
	tfGraph, err := terraform.GetGraph(ctx, tf_path, "terraform")
	if err != nil {
		return fmt.Errorf("Failed to get graph: %s", err)
	}

	gograph, err := graph.LoadGraph(tfGraph)
	if err != nil {
		return fmt.Errorf("Failed to load graph: %s", err)
	}

	module := terraform.OpenTerraformFiles(tf_path)
	hclCode := terraform.FromConfig(module)

	state, err := terraform.GetState(ctx, tf_path, "terraform")
	if err != nil {
		return fmt.Errorf("Failed to get state: %s", err)
	}

	for _, layer := range config.Layers {
		childs := graph.GetChildren(layer.RootNode, gograph)
		ids := extractor.GetIds(layer.RootNode, state)

		slog.Debug(
			"Processing layer",
			"layer", layer.Name,
			"rootNode", layer.RootNode,
			"ids", ids,
		)

		writer.WriteLayer(filepath.Join(tf_path, "tfsplit", layer.Name), childs, hclCode)
	}

	return nil
}

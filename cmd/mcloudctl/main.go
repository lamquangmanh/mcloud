package mcloudctl

import (
	"mcloud/pkg/logger"
	"os"

	"github.com/urfave/cli/v2"
)

// main is the entry point for the mcloudctl CLI application.
// It sets up the CLI app, defines available commands, and runs the app.
//
// Logic:
//   1. Create a new cli.App instance with name and usage description
//   2. Register the 'init' command for cluster initialization
//   3. Parse command-line arguments and execute the selected command
//   4. Log any errors encountered during execution
//
// Example Input (Command Line):
//   $ mcloudctl init --name test-cluster
//
// Example Output (Success):
//   [INFO] 2026-01-03 10:30:45 Initializing mcloud cluster: test-cluster
//   ... (see InitCommand for full output)
//
// Example Output (Error - Missing Name):
//   [ERROR] 2026-01-03 10:30:45 flag --name is required
//   ...existing code...
func main() {
	app := &cli.App{
		Name:  "mcloud",
		Usage: "Mini cloud bootstrap tool",
		Commands: []*cli.Command{
			{
				Name:   "init",
				Usage:  "Initialize a new mcloud cluster",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Aliases:  []string{"n"},
						Usage:    "Cluster name",
						Required: true,
					},
				},
				Action: InitCommand, // See cmd/mcloudctl/init.go for full logic
			},
		},
	}

	// Run the CLI app and handle errors
	if err := app.Run(os.Args); err != nil {
		logger.Error("%v", err)
	}
}

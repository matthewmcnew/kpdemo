package main

import (
	"errors"
	"fmt"
	"github.com/matthewmcnew/build-service-visualization/logs"
	"github.com/matthewmcnew/build-service-visualization/populate"
	"github.com/matthewmcnew/build-service-visualization/rebase"
	"github.com/matthewmcnew/build-service-visualization/server"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "A tool to demo build service & kpack",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to the Build Service Demo")
	},
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func init() {
	rootCmd.AddCommand(populateCmd(),
		serveCmd(),
		updateRunImageCmd(),
		cleanupCmd(),
		logsCmd(),
	)
}

func populateCmd() *cobra.Command {
	var registry string
	var cacheSize string
	var count int32
	var cmd = &cobra.Command{
		Use:     "populate",
		Aliases: []string{"setup"},
		Short:   "Populate Build Service with Images",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("using existing builder and run image")
			populate.Populate(count, registry, cacheSize)
			return nil
		},
	}
	cmd.Flags().StringVarP(&cacheSize, "cache-size", "s", "500Mi", "the cache size to use for build service images")

	cmd.Flags().StringVarP(&registry, "registry", "r", "", "registry to deploy images into")
	_ = cmd.MarkFlagRequired("registry")

	cmd.Flags().Int32VarP(&count, "count", "c", 0, "the number of images to populate in build service")
	_ = cmd.MarkFlagRequired("count")

	return cmd
}

func serveCmd() *cobra.Command {
	var port string
	var cmd = &cobra.Command{
		Use:     "visualization",
		Aliases: []string{"serve", "ui"},
		Short:   "Setup a local web server build service visualization ",
		RunE: func(cmd *cobra.Command, args []string) error {
			go func() {
				fmt.Println("Starting Up")
				url := fmt.Sprintf("http://localhost:%s", port)
				fmt.Printf("Open up a browser to %s\n", url)

				server.OpenBrowser(url)
			}()

			server.Serve(port)

			return nil
		},
	}

	cmd.Flags().StringVarP(&port, "port", "p", "8080", "registry to deploy images into")

	return cmd
}

func updateRunImageCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "update-stack",
		Aliases: []string{"rebase", "update-run-image"},
		Short:   "Demo an update by pushing an updated stack run image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return rebase.UpdateRunImage()
		},
	}

	return cmd
}

func cleanupCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "cleanup",
		Short: "Remove build service demo images",
		RunE: func(cmd *cobra.Command, args []string) error {
			return populate.Cleanup()
		},
	}

	return cmd
}

func logsCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "logs",
		Short:   "Stream build logs from an image",
		Example: "pbdemo logs <image-name>",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("no image name provided")
			}

			image := args[0]

			return logs.Logs(image)
		},
	}

	return cmd
}

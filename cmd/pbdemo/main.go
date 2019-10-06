package main

import (
	"fmt"
	"github.com/matthewmcnew/build-service-visualization/populate"
	"github.com/matthewmcnew/build-service-visualization/relocatebuilder"
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
	rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(populateCmd())
	rootCmd.AddCommand(serveCmd())
}

func populateCmd() *cobra.Command {
	var registry string
	var count int32
	var cmd = &cobra.Command{
		Use:   "populate",
		Short: "Populate Build Service with Images",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Relocating Builder and Run Image. This will take a moment.")

			relocated, err := relocatebuilder.Relocate(registry)
			if err != nil {
				return err
			}

			populate.Populate(count, relocated.BuilderImage, registry)
			return nil
		},
	}
	cmd.Flags().StringVarP(&registry, "registry", "r", "", "registry to deploy images into")
	_ = cmd.MarkFlagRequired("registry")

	cmd.Flags().Int32VarP(&count, "count", "c", 0, "the number of images to populate in build service")
	_ = cmd.MarkFlagRequired("count")

	return cmd
}

func serveCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "visualization",
		Aliases: []string{"serve", "ui"},
		Short:   "Setup a local web server build service visualization ",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Starting Up")
			fmt.Println("Open up a browser to http://localhost:8081/")

			server.Serve()

			return nil
		},
	}

	return cmd
}

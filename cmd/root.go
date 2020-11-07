package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "wfs",
		Short: "Webfilesystem",
		Long:  `Access the web through your file system.`,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

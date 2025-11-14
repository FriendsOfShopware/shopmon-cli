package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "shopmon-cli",
	Short: "Shopware Monitoring CLI",
	Long:  `A CLI tool for monitoring and managing Shopware applications.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
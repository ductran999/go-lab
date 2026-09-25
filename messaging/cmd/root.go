/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "xplr-distributed-mq",
	Short: "Experimental distributed message queue platform",
	Long: `xplr-distributed-mq is an experimental platform for exploring
distributed message queue architectures.

It supports multiple example such as:
  - kafka producer with sarama, kafkago, etc.
Use subcommands to start specific components.
	`,
	SilenceUsage:  true,
	SilenceErrors: false,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

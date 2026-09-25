/*
Copyright © 2026
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	GitCommit = "none"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(
			"xplr-distributed-mq\nversion: %s\ncommit:  %s\nbuild:   %s\n",
			Version,
			GitCommit,
			BuildDate,
		)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

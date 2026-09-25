/*
Copyright © 2026 Duc Ngo
*/
package cmd

import (
	"fmt"
	"go-lab/messaging/examples/producer/kafka"

	"github.com/spf13/cobra"
)

var lib string

// kafkaCmd represents the kafka command
var kafkaCmd = &cobra.Command{
	Use:   "kafka",
	Short: "Run Kafka producer/consumer example",
	Long: `Run Kafka example with different Go client libraries.

Supported libraries:
  - sarama
  - kafkago
  - franzgo
  - confluent
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch lib {
		case "kafkago":
			kafka.RunKafkaGoExample()
		case "confluent":
			kafka.RunConfluentKafkaGoExample()
		case "franzgo":
			kafka.RunFranzGoExample()
		default:
			kafka.RunSaramaExample()
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(kafkaCmd)

	kafkaCmd.Flags().StringVarP(
		&lib,
		"lib",
		"l",
		"",
		"kafka client library to use (sarama | kafkago | confluent | franzgo)",
	)

	kafkaCmd.RegisterFlagCompletionFunc(
		"lib",
		cobra.FixedCompletions(
			[]string{"sarama", "kafkago", "confluent", "franzgo"},
			cobra.ShellCompDirectiveNoFileComp),
	)

	kafkaCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		switch lib {
		case "sarama", "kafkago", "confluent", "franzgo":
			return nil
		default:
			return fmt.Errorf(
				"invalid value for --lib %q (allowed: sarama, kafkago, confluent, franzgo)",
				lib,
			)
		}
	}
}

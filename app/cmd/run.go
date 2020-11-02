// Package command contains cli commands.
package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/METALmasterKS/simplinic/aggregator"
	"github.com/METALmasterKS/simplinic/app/definition"
	"github.com/METALmasterKS/simplinic/generator"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Bus command.
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run command",
	RunE: func(_ *cobra.Command, _ []string) (err error) {
		var (
			ctx               = diContainer.Get(definition.DefContextName).(context.Context)
			generatorFactory  = diContainer.Get(definition.DefGeneratorFactoryName).(generator.GeneratorFactory)
			aggregatorFactory = diContainer.Get(definition.DefAggregatorFactoryName).(aggregator.AggregatorFactory)
		)

		ctxCancel, cancel := context.WithCancel(ctx)

		var aggregators []aggregator.Options
		if err := viper.UnmarshalKey("aggregators", &aggregators); err != nil {
			return err
		}
		for _, opts := range aggregators {
			if _, err := aggregatorFactory.CreateAggregator(ctxCancel, opts); err != nil {
				return err
			}
		}

		var generators []generator.Options
		if err := viper.UnmarshalKey("generators", &generators); err != nil {
			return err
		}
		for _, genOpts := range generators {
			if _, err := generatorFactory.CreateGenerator(ctxCancel, genOpts); err != nil {
				return err
			}
		}

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		<-signalChan
		log.Info().Msg("5 seconds to stopping...")
		cancel()
		<-time.After(5 * time.Second)

		return nil
	},
}

// Command init function.
func init() {
	rootCmd.AddCommand(runCmd)
}

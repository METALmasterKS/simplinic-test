// Package command contains cli commands.
package cmd

import (
	"context"
	"github.com/METALmasterKS/simplinic/app/definition"
	"github.com/METALmasterKS/simplinic/generator"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Bus command.
var busCmd = &cobra.Command{
	Use:   "run",
	Short: "Run command",
	RunE: func(_ *cobra.Command, _ []string) (err error) {
		var ctx = diContainer.Get(definition.DefContextName).(context.Context)
		var generatorFactory = diContainer.Get(definition.DefGeneratorFactoryName).(generator.GeneratorFactory)
		var generators []generator.GeneratorOptions

		if err := viper.UnmarshalKey("generators", &generators); err != nil {
			return err
		}

		for _, genOpts := range generators {
			if _, err := generatorFactory.CreateGenerator(ctx, genOpts); err != nil {
				return err
			}
		}

		//// wait for stop application
		//signalChan := make(chan os.Signal, 1)
		//signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		//<-signalChan

		<-ctx.Done()
		log.Info().Msg("stopping...")

		return nil
	},
}

// Command init function.
func init() {
	rootCmd.AddCommand(busCmd)
}

package cmd

import (
	"fmt"
	"github.com/METALmasterKS/simplinic/app/definition"
	di "github.com/sarulabs/di/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

var (
	// Used for flags.
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "app [command]",
		Short: "application",
		Long:  `Application for test simplinic`,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	diContainer di.Container
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "json config file")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		builder, err := di.NewBuilder()
		if err != nil {
			return err
		}

		err = builder.Add(
			di.Def{
				Build: func(ctn di.Container) (interface{}, error) {
					return rootCmd.Context(), nil
				},
				Name: definition.DefContextName,
			},
			definition.DefBus(),
			definition.DefGeneratorFactory(),
		)

		diContainer = builder.Build()

		return err
	}

}
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error: %w", err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		var (
			appPath string
			err     error
		)
		if appPath, err = filepath.Abs(filepath.Dir(os.Args[0])); err != nil {
			fmt.Println("Error: %w", err)
			os.Exit(1)
		}

		viper.AddConfigPath(appPath)
		viper.SetConfigName("config")
		viper.SetConfigType("json")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

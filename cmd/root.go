package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/treethought/lyuba/ui"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "masto",
	Short: "mastodon tui",
	Run: func(cmd *cobra.Command, args []string) {
		app := ui.NewApp()
		// app := ui.New()
		app.Start()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.masto.yaml)")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		_, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".masto" (without extension).
		// viper.AddConfigPath(home)
		viper.SetConfigName(".lyuba")
		viper.SetConfigName("lyuba")         // name of config file (without extension)
		viper.SetConfigType("yaml")          // REQUIRED if the config file does not have the extension in the name
		viper.AddConfigPath("$HOME/.config") // call multiple times to add many search paths
		viper.AddConfigPath("$HOME")         // call multiple times to add many search paths
		viper.AddConfigPath(".")             // optionally look for config in the working directory
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				viper.SafeWriteConfig()
			} else {
				panic(fmt.Errorf("Fatal error config file: %s \n", err))
			}
		}

	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

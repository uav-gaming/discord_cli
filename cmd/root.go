/*
Copyright © 2022 Tianjiao Huang <tjhu@tjhu.dev>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const CONFIG_FILENAME string = ".discord_cli"

// CLI configurations/flags.
type Configuration struct {
	// Path to the configuration file in used.
	ConfigFile    string
	DiscordToken  string
	ApplicationID discord.AppID
	GuildID       discord.GuildID
	UserID        discord.UserID
}

// Globals
var (
	config Configuration
	client *api.Client
)

// Helper for child commands to invoke the ancestors' `PersistentPreRun`
// https://github.com/spf13/cobra/issues/216#issuecomment-703846787
func callPersistentPreRun(cmd *cobra.Command, args []string) {
	if parent := cmd.Parent(); parent != nil {
		if parent.PersistentPreRun != nil {
			parent.PersistentPreRun(parent, args)
		}
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "discord_cli",
	Short: "A CLI tool for interacting with the Discord REST APIs.",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		callPersistentPreRun(cmd, args)
		token := viper.GetString("discord_token")
		application_id := discord.AppID(viper.GetUint64("application_id"))

		if token == "" || !application_id.IsValid() {
			logrus.Fatal("Required configs are not set. Either set them with `discord_cli config` or through command line flags")
		}
		client = api.NewClient(token)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}

func init() {
	cobra.OnInitialize(initConfig)

	// A list of flags that can be set in the config file.
	rootCmd.PersistentFlags().StringVarP(&config.DiscordToken, "discord_token", "t", "", "Discord token.")
	rootCmd.PersistentFlags().Uint64VarP((*uint64)(&config.ApplicationID), "application_id", "a", 0, "Discord application ID.")
	// No shorthand for the 'guild_id' flag because it will conflict with the 'guild' command.
	rootCmd.PersistentFlags().Uint64Var((*uint64)(&config.GuildID), "guild_id", 0, "Discord guild ID.")
	rootCmd.PersistentFlags().Uint64VarP((*uint64)(&config.GuildID), "user_id", "u", 0, "Discord user ID.")
	viper.BindPFlags(rootCmd.PersistentFlags())

	// Flags that cannot be set in the config file.
	rootCmd.PersistentFlags().StringVar(&config.ConfigFile, "config", "", fmt.Sprintf("config file (default is $HOME/%s)", CONFIG_FILENAME))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if config.ConfigFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(config.ConfigFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory and the current directory with name ".discord_cli" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(CONFIG_FILENAME)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		config.ConfigFile = viper.ConfigFileUsed()
		fmt.Fprintln(os.Stderr, "Using config file:", config.ConfigFile)
	}
}

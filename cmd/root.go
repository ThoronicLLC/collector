package cmd

import (
  log "github.com/sirupsen/logrus"
  "github.com/spf13/cobra"
  "github.com/spf13/pflag"
  "github.com/spf13/viper"
  "os"
  "strings"
)

var cfgPath string

// Build variables - passed in ldflags on build
var ApplicationName string
var BuildBranch string
var BuildDate string
var BuildEnv string
var BuildRevision string
var BuildVersion string
var logVerbose = false

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
  Use:   "collector",
  Short: "A generic log collector that can be used as a CLI binary or an imported package.",
  PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
    if logVerbose {
      log.SetLevel(log.DebugLevel)
    }
    return nil
  },
  Version: BuildVersion,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
  rootCmd.SetVersionTemplate(printHumanVersionInfo())
  err := rootCmd.Execute()
  if err != nil {
    os.Exit(1)
  }
}

func init() {
  cobra.OnInitialize(func() {
    initConfig()
    postInitCommands(rootCmd.Commands())
  })

  // Persistent Flags
  rootCmd.PersistentFlags().BoolVarP(&logVerbose, "verbose", "v", false, "log debug level messages")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
  replacer := strings.NewReplacer(".", "_", "-", "_")
  viper.SetEnvPrefix("COLLECTOR")
  viper.SetEnvKeyReplacer(replacer)
  viper.AutomaticEnv() // read in environment variables that match
}

func postInitCommands(commands []*cobra.Command) {
  for _, cmd := range commands {
    presetRequiredFlags(cmd)
    if cmd.HasSubCommands() {
      postInitCommands(cmd.Commands())
    }
  }
}

func presetRequiredFlags(cmd *cobra.Command) {
  _ = viper.BindPFlags(cmd.Flags())
  cmd.Flags().VisitAll(func(f *pflag.Flag) {
    if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
      _ = cmd.Flags().Set(f.Name, viper.GetString(f.Name))
    }
  })
}

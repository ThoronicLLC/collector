package cmd

import (
  "fmt"
  "github.com/ThoronicLLC/collector/internal/cli"
  log "github.com/sirupsen/logrus"
  "path/filepath"

  "github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
  Use:     "serve",
  Aliases: []string{"server", "run", "start"},
  Short:   "Run a set of log collectors.",
  Long: `The collector will run a set of instances that will collect data from
an input, process that data, and then output it to the configured plugins. Ex:

collector run --config /opt/collector`,
  Run: func(cmd *cobra.Command, args []string) {
    if cfgPath == "" {
      cobra.CheckErr(fmt.Errorf("missing config"))
    }

    if !cli.DirectoryExists(cfgPath) {
      cobra.CheckErr(fmt.Errorf("supplied config directory does not exist"))
    }

    cfgPath, err := filepath.Abs(cfgPath)
    if err != nil {
      cobra.CheckErr(fmt.Errorf("issue getting absoulte path: %s", err))
    }

    err = cli.Run(cfgPath)
    if err != nil {
      log.Errorf("%s", err)
    }
  },
}

func init() {
  rootCmd.AddCommand(serveCmd)
  serveCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "config directory")
  _ = serveCmd.MarkPersistentFlagRequired("config")
}

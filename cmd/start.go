package cmd

import (
	"fmt"
	"github.com/ThoronicLLC/collector/internal/cli"
	log "github.com/sirupsen/logrus"
	"path/filepath"

	"github.com/spf13/cobra"
)

// startCmd represents the serve command
var startCmd = &cobra.Command{
	Use:     "start",
	Aliases: []string{"server", "run", "serve"},
	Short:   "Run a set of log collectors.",
	Long: `The collector will run a set of instances that will collect data from
an input, process that data, and then output it to the configured plugins. 

Example Command:
collector start --config /etc/collector`,
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
	rootCmd.AddCommand(startCmd)
	startCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "config directory")
	_ = startCmd.MarkPersistentFlagRequired("config")
}

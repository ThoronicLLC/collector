package cmd

import (
  "encoding/json"
  "fmt"
  "github.com/spf13/cobra"
  "runtime"
  "strconv"
  "time"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
  Use:   "version",
  Short: "Print version and environment info",
  Long:  `Print version and environment info. This is useful in bug reports.`,
  Run: func(cmd *cobra.Command, args []string) {
    buildInfo := ""
    if result, _ := cmd.Flags().GetBool("machine"); result {
      buildInfo = printMachineVersionInfo()
    } else {
      buildInfo = printHumanVersionInfo()
    }
    fmt.Printf("%s\n", buildInfo)
  },
}

func init() {
  rootCmd.AddCommand(versionCmd)

  // Persistent Flags

  // Local Flags
  versionCmd.Flags().BoolP("machine", "m", false, "print version and build info in a machine readable format (json)")
}

func printMachineVersionInfo() string {
  buildTimeString, _ := strconv.ParseInt(BuildDate, 10, 64)
  buildTime := time.Unix(buildTimeString, 0)
  buildInfo := make(map[string]string, 0)
  buildInfo["version"] = BuildVersion
  buildInfo["environment"] = BuildEnv
  buildInfo["branch"] = BuildBranch
  buildInfo["date"] = buildTime.Format(time.RFC3339)
  buildInfo["commit"] = BuildRevision
  buildInfoBytes, err := json.Marshal(buildInfo)
  if err != nil {
    return ""
  }
  return string(buildInfoBytes)
}

func printHumanVersionInfo() string {
  buildTimeString, _ := strconv.ParseInt(BuildDate, 10, 64)
  buildTime := time.Unix(buildTimeString, 0)
  buildInfo := ""
  buildInfo += fmt.Sprintf("%s - version %s\n", ApplicationName, BuildVersion)
  buildInfo += fmt.Sprintf("  branch: %s\n", BuildBranch)
  buildInfo += fmt.Sprintf("  revision: %s\n", BuildRevision)
  buildInfo += fmt.Sprintf("  build date: %s\n", buildTime.Format(time.RFC3339))
  buildInfo += fmt.Sprintf("  build env: %s\n", BuildEnv)
  buildInfo += fmt.Sprintf("  go version: %s\n", runtime.Version())
  return buildInfo
}

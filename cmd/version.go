package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
    version = "dev"
	buildTime = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Wyświetla wersję aplikacji",
	Annotations: map[string]string{
		"SkipSetup": "true",
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ksef-cli version:", version)
		fmt.Println("Build time:", buildTime)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

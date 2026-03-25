package cmd

import (
	"log/slog"
	"os"
	"path/filepath"
)

func initConfig() {
	verbosity, err := rootCmd.PersistentFlags().GetCount("verbose")
	if err != nil {
		slog.Error("getting verbosity level failed", "err", err)
		return
	}
	setupLogger(verbosity)

	//configFile, err := rootCmd.PersistentFlags().GetString("configFile")
	//if err != nil {
	//	slog.Error("Cound not parse 'cofig' option", "err", err)
	//	return
	//}
	//slog.Debug("initializing configuration", "configFile", configFile)

}

func defaultConfigFile() string {
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "ksef-cli/config.yaml")
}

func defaultCacheDir() string {
	dir, _ := os.UserCacheDir()
	return filepath.Join(dir, "ksef-cli")
}

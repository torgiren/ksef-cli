package cmd

import (
	"context"
	"log/slog"
	"os"
	"fmt"

	"github.com/spf13/cobra"
	//"github.com/spf13/viper"

	"github.com/torgiren/ksef-cli/pkg/ksef"
	"github.com/torgiren/ksef-cli/internal/profile"
)

var (
	client  ksef.Client
	//nip     string
	ctx     context.Context
	config  profile.Config
	currentProfile profile.Profile
)

var rootCmd = &cobra.Command{
	Use:   "ksef-cli",
	Short: "Klient tekstowy KSeF",
	Long: `Klient tekstowy dla Krajowego Systemu e-Faktur (KSeF).

API produkcyjne: https://api.ksef.mf.gov.pl/v2
API testowe:     https://api-test.ksef.mf.gov.pl/v2`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx = context.Background()
		slog.Debug("load config")
		profileConfig, err := profile.LoadConfig(cmd.Flag("configFile").Value.String())
		if err != nil {
			slog.Warn("could not load config", "err", err)
			slog.Debug("creating default config")
			profileConfig = profile.Config{}
			profile.SaveConfig(profileConfig, cmd.Flag("config").Value.String())

		}
		config = profileConfig
		slog.Debug("config loaded", "currentProfile", profileConfig.CurrentProfile, "profilesCount", len(profileConfig.Profiles))
		selectedProfile := profileConfig.CurrentProfile
		if cmd.Flag("profile").Value.String() != "" {
			slog.Debug("profile provided via flag, overriding current profile", "profile", cmd.Flag("profile").Value.String())
			selectedProfile = cmd.Flag("profile").Value.String()
		}
		currentProfile = profileConfig.Profiles[selectedProfile]
		slog.Debug("current profile loaded", "name", currentProfile.Name, "nip", currentProfile.Nip, "api", currentProfile.Api)
		if cmd.Flag("nip").Value.String() != "" {
			slog.Debug("NIP provided via flag, overriding current profile NIP", "nip", cmd.Flag("nip").Value.String())
			currentProfile.Nip = cmd.Flag("nip").Value.String()
		}

		slog.Debug("pre-run check", "command", cmd.Name(), "parent", cmd.Parent().Name())


		if cmd.Annotations["NipRequired"] == "true" {
			slog.Debug("NIP is required for this command")
			if currentProfile.Nip == "" {
				slog.Error("NIP is required but not set", "command", cmd.Name())
				return fmt.Errorf("NIP is required but not set. Set it using 'ksef-cli profile set <NIP>' or use the --nip flag")
			}
			slog.Debug("NIP validation passed", "nip", currentProfile.Nip)
		}
		return initClient(cmd)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", defaultConfigFile(), "config file")
	rootCmd.PersistentFlags().String("configFile", defaultConfigFile(), "config file")
	rootCmd.PersistentFlags().StringP("api", "a", "https://api.ksef.mf.gov.pl/v2", "API endpoint for KSeF")
	rootCmd.PersistentFlags().String("nip", "", "NIP number")
	rootCmd.PersistentFlags().String("profile", "", "Profile name")
	rootCmd.PersistentFlags().String("cache-dir", defaultCacheDir(), "Directory to store cache files")
	rootCmd.PersistentFlags().CountP("verbose", "v", "Increase verbosity level (can be used multiple times). Levels: -v (INFO), -vv (DEBUG), -vvv (TRACE), -vvvv (SECRET)")
	rootCmd.PersistentFlags().Bool("test", false, "Use test API endpoint (https://api-test.ksef.mf.gov.pl/v2)")
	rootCmd.PersistentFlags().String("output", "text", "Output format: text, json")

	//viper.BindPFlag("currentNip", rootCmd.PersistentFlags().Lookup("nip"))
	//viper.BindPFlag("cacheDir", rootCmd.PersistentFlags().Lookup("cache-dir"))
	//viper.BindPFlag("api", rootCmd.PersistentFlags().Lookup("api"))
}

package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	//"github.com/spf13/viper"
	"github.com/torgiren/ksef-cli/internal/profile"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Logowanie do KSeF",
	Annotations: map[string]string{
		"NipRequired": "true",
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Logowanie do KSeF...")

		var ksefToken string
		if cmd.Flag("token").Value.String() != "" {
			slog.Debug("using token from command flag")
			ksefToken = cmd.Flag("token").Value.String()
		} else {
			t := client.GetTokens()
			if t == nil || t.KsefToken == "" {
				fmt.Fprintln(os.Stderr, "Nie podano tokenu KSeF. Użyj flagi --token lub zaloguj się ponownie bez tej flagi, aby pobrać token z cache'u.")
				slog.Error("no KSeF token available for login")
				return
			}
			slog.Debug("using KSeF token from client tokens", "ksefToken", t.KsefToken)
			ksefToken = t.KsefToken
		}

		tokens, err := client.Login(ctx, currentProfile.Nip, ksefToken)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Błąd logowania:", err)
			slog.Error("login failed", "err", err)
			return
		}

		if cmd.Flag("save-token").Value.String() == "true" {
			slog.Debug("saving KSeF token in tokens struct")
			tokens.KsefToken = ksefToken
		}

		fmt.Println("Zalogowano pomyślnie!")
		slog.Debug("login tokens", "accessTokenExpiry", tokens.AccessTokenExpiry, "refreshTokenExpiry", tokens.RefreshTokenExpiry)

		err = profile.SaveTokens(currentProfile.Name, tokens, cmd.Flag("cache-dir").Value.String())
		if err != nil {
			fmt.Fprintln(os.Stderr, "Błąd zapisywania tokenów:", err)
			slog.Error("saving tokens failed", "err", err)
			return
		}
		slog.Info("tokens saved to cache", "nip", currentProfile.Nip)

		slog.Debug("current config", "config", config, "currentProfile", currentProfile)
		config.Profiles[currentProfile.Name] = currentProfile
		slog.Debug("profile added to config", "profileName", currentProfile.Name)
		config.CurrentProfile = currentProfile.Name
		slog.Debug("current profile set in config", "currentProfile", config.CurrentProfile)
		err = profile.SaveConfig(config, cmd.Flag("configFile").Value.String())
		if err != nil {
			fmt.Fprintln(os.Stderr, "Błąd zapisywania konfiguracji:", err)
			slog.Error("saving config failed", "err", err)
			return
		}
		slog.Info("config saved", "currentProfile", config.CurrentProfile, "profilesCount", len(config.Profiles))

		//if err = viper.WriteConfig(); err != nil {
		//	fmt.Fprintln(os.Stderr, "Błąd zapisywania konfiguracji:", err)
		//	slog.Error("writing config failed", "err", err)
		//	return
		//}
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		//cmd.Parent().PersistentPreRunE(cmd, args) // Call the parent command's PreRunE to ensure client initialization
		slog.Debug("login command PreRunE", "command", cmd.Name(), "parent", cmd.Parent().Name())
		if cmd.Flag("profile").Value.String() == "" {
			fmt.Fprintln(os.Stderr, "Nie wybrano profilu. Użyj flagi --profile, aby wybrać profil do logowania.")
			slog.Error("no profile selected for login")
			os.Exit(1)
		}
		currentProfile.Name = cmd.Flag("profile").Value.String()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().Bool("save-token", true, "Zapisz KSeF token w cache'u (domyślnie: true)")
	loginCmd.Flags().String("token", "", "Token do KSeF (jeśli nie podano, zostanie użyty token z cache'u jeśli jest dostępny)")
}

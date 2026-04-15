package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
	//"github.com/spf13/viper"

	"github.com/torgiren/ksef-cli/internal/profile"
	"github.com/torgiren/ksef-cli/pkg/ksef"
)

//func validateNip(cmd *cobra.Command) error {
//	nipFlag := cmd.Flag("nip").Value.String()
//	if nipFlag != "" {
//		currentProfile.Nip = nipFlag
//		slog.Debug("NIP set from flag", "nip", nipFlag)
//	}
//	if currentProfile.Nip == "" && cmd.Annotations["NipRequired"] == "true" {
//		fmt.Fprintln(os.Stderr, "Numer NIP jest wymagany. Ustaw go za pomocą polecenia 'ksef-cli profile set <NIP>' albo użyj flagi --nip")
//		return fmt.Errorf("NIP is required but not set")
//	}
//	return nil
//}

func initClient(cmd *cobra.Command) error {
	if cmd.Annotations["skipClientInit"] == "true" {
		slog.Debug("skipping client initialization")
		return nil
	}

	if cmd.Flag("test").Value.String() == "true" {
		slog.Debug("test mode enabled, using test API endpoint")
		currentProfile.Api = "https://api-test.ksef.mf.gov.pl/v2"
	}
	slog.Debug("API endpoint from profile", "currentProfileApi", currentProfile.Api, "testFlag", cmd.Flag("test").Value.String(), "apiFlag", cmd.Flag("api").Value.String())
	if cmd.Flag("api").Value.String() != "" && (cmd.Flag("api").Changed || currentProfile.Api == "") {
		slog.Debug("API endpoint provided via flag, overriding current profile API", "api", cmd.Flag("api").Value.String())
		currentProfile.Api = cmd.Flag("api").Value.String()
	}
	slog.Debug("API endpoint", "endpoint", currentProfile.Api)

	var err error
	client, err = ksef.NewClient(currentProfile.Api)
	if err != nil {
		slog.Error("client creation failed", "err", err)
		fmt.Fprintln(os.Stderr, "Błąd tworzenia klienta:", err)
		return err
	}

	return loadAndRefreshTokens(cmd)
}

func loadAndRefreshTokens(cmd *cobra.Command) error {
	cacheDir := cmd.Flag("cache-dir").Value.String()
	tokens, err := profile.LoadTokens(currentProfile.Name, cacheDir)
	if err != nil {
		slog.Debug("no cached tokens found", "err", err)
		return nil
	}

	client.SetTokens(tokens)
	slog.Debug("cached tokens loaded", "accessTokenExpiry", tokens.AccessTokenExpiry, "refreshTokenExpiry", tokens.RefreshTokenExpiry)

	if tokens.RefreshTokenExpiry.Before(time.Now()) {
		slog.Warn("refresh token expired, login required")
		return nil
	}

	slog.Debug("checking access token expiry", "accessTokenExpiry", tokens.AccessTokenExpiry, "currentTime", time.Now())
	if tokens.AccessTokenExpiry.Before(time.Now()) {
		slog.Warn("access token expired, refreshing")
		if _, err := client.RefreshTokens(ctx); err != nil {
			slog.Error("token refresh failed", "err", err)
			return fmt.Errorf("error refreshing tokens: %w", err)
		}
		slog.Info("access token refreshed")
		profile.SaveTokens(currentProfile.Name, client.GetTokens(), cacheDir)
		//profile.SaveTokens(viper.GetString("currentNip"), client.GetTokens(), cacheDir)
	}

	return nil
}

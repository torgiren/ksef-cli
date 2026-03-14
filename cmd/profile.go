package cmd

import (
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Zarządzanie profilami",
	Long:  `Polecenia do tworzenia, listowania i usuwania profili konfiguracyjnych.`,
}

func init() {
	rootCmd.AddCommand(profileCmd)
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "Wylistuj profile",
	Long:  `Wyświetla listę wszystkich zdefiniowanych profili konfiguracyjnych.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("profileList called")
	},
}

func init() {
	profileCmd.AddCommand(profileListCmd)
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var profileSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Ustaw aktywny profil",
	Long:  `Ustawia wskazany profil jako aktywny (używany domyślnie przez wszystkie polecenia).`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("profileSet called")
	},
}

func init() {
	profileCmd.AddCommand(profileSetCmd)
}

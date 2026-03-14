package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var profileDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Usuń profil",
	Long:  `Usuwa wskazany profil konfiguracyjny wraz z powiązanymi tokenami.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("profileDelete called")
	},
}

func init() {
	profileCmd.AddCommand(profileDeleteCmd)
}

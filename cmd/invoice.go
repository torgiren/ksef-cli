package cmd

import (
	"github.com/spf13/cobra"
)

var invoiceCmd = &cobra.Command{
	Use:   "invoice",
	Short: "Zarządzanie fakturami",
	Long:  `Polecenia do pobierania i zarządzania fakturami z KSeF.`,
}

func init() {
	rootCmd.AddCommand(invoiceCmd)
}

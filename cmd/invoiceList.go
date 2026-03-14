package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/torgiren/ksef-cli/pkg/ksef"
)

var invoiceListCmd = &cobra.Command{
	Use:   "list",
	Short: "Pobierz listę faktur",
	Long:  `Pobiera listę faktur z KSeF dla podanego profilu.`,
	Annotations: map[string]string{
		"NipRequired": "true", // This annotation can be used in the root command's PersistentPreRunE to enforce NIP check for this subcommand
	},
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("fetching invoices")
		query := ksef.InvoiceQuery{
			From:        time.Now().AddDate(0, -1, 0),
			To:          time.Now(),
			SubjectType: ksef.Subject2,
		}
		page := ksef.InvoicePage{
			PageSize:   15,
			PageOffset: 0,
		}
		invoices, err := client.GetInvoices(ctx, query, page)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Błąd pobierania faktur:", err)
			slog.Error("fetching invoices failed", "err", err)
			return
		}
		slog.Debug("invoices retrieved", "count", len(invoices))

		slog.Info("output format", "format", cmd.Flag("output").Value.String())
		switch cmd.Flag("output").Value.String() {
		case "json":
			printAsJSON(invoices)
		case "text":
			printAsText(invoices)
		default:
			fmt.Fprintln(os.Stderr, "Nieznany format wyjścia:", cmd.Flag("output").Value.String())
			slog.Warn("unknown output format", "format", cmd.Flag("output").Value.String())
		}
	},
	//PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
	//	cmd.Parent().PersistentPreRunE(cmd, args) // Call the parent command's PersistentPreRunE
	//	fmt.Println("Running invoice list pre-run checks")
	//	return nil
	//},
	//	Annotations: map[string]string{
	//		"skipNipCheck": "true", // This annotation can be used in the root command's PersistentPreRunE to skip NIP check for this subcommand
	//	},
}

func init() {
	invoiceCmd.AddCommand(invoiceListCmd)
}

func printAsJSON(invoices []ksef.Invoice) {
	jsonData, err := json.MarshalIndent(invoices, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Błąd konwersji faktur do JSON:", err)
		return
	}
	fmt.Println(string(jsonData))
}

func printAsText(invoices []ksef.Invoice) {
	for _, invoice := range invoices {
		fmt.Printf("Invoice: %s, Date: %s, Gross: %.2f, Net: %.2f, Seller: %s\n",
			invoice.InvoiceNumber,
			invoice.PermanentStorageDate.Format("2006-01-02"),
			invoice.GrossAmount,
			invoice.NetAmount,
			invoice.SellerName,
		)
	}
}

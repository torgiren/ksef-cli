package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/torgiren/ksef-cli/pkg/ksef"

	"github.com/jedib0t/go-pretty/v6/table"
)

func cmdToQuery(cmd *cobra.Command) (ksef.InvoiceQuery, error) {
	shortForm := "2006-01-02"
	fromStr, err := cmd.Flags().GetString("from")
	if err != nil {
		return ksef.InvoiceQuery{}, fmt.Errorf("invalid 'from' date: %w", err)
	}
	slog.Debug("parsing 'from' date", "fromStr", fromStr)
	from, err := time.Parse(shortForm, fromStr)
	if err != nil {
		return ksef.InvoiceQuery{}, fmt.Errorf("invalid 'from' date format: %w", err)
	}
	toStr, err := cmd.Flags().GetString("to")
	if err != nil {
		return ksef.InvoiceQuery{}, fmt.Errorf("invalid 'to' date: %w", err)
	}
	slog.Debug("parsing 'to' date", "toStr", toStr)
	to, err := time.Parse(shortForm, toStr)
	if err != nil {
		return ksef.InvoiceQuery{}, fmt.Errorf("invalid 'to' date format: %w", err)
	}
	subjectTypeStr, err := cmd.Flags().GetString("subject")
	if err != nil {
		return ksef.InvoiceQuery{}, fmt.Errorf("invalid 'subject' type: %w", err)
	}
	slog.Debug("parsing 'subject' type", "subjectTypeStr", subjectTypeStr)
	var subjectType ksef.SubjectType
	switch subjectTypeStr {
	case "subject1":
		subjectType = ksef.Subject1
	case "subject2":
		subjectType = ksef.Subject2
	case "subject3":
		subjectType = ksef.Subject3
	case "authorized":
		subjectType = ksef.SubjectAuthorized
	default:
		return ksef.InvoiceQuery{}, fmt.Errorf("invalid subject type: %s", subjectTypeStr)
	}
	return ksef.InvoiceQuery{
		From:        from,
		To:          to,
		SubjectType: subjectType,
	}, nil
}

func cmdToPage(cmd *cobra.Command) (ksef.InvoicePage, error) {
	pageSize, err := cmd.Flags().GetInt32("pagesize")
	if err != nil {
		return ksef.InvoicePage{}, fmt.Errorf("invalid page size: %w", err)
	}
	pageOffset, err := cmd.Flags().GetInt32("pageoffset")
	if err != nil {
		return ksef.InvoicePage{}, fmt.Errorf("invalid page offset: %w", err)
	}
	return ksef.InvoicePage{
		PageSize:   pageSize,
		PageOffset: pageOffset,
	}, nil
}

var invoiceListCmd = &cobra.Command{
	Use:   "list",
	Short: "Pobierz listę faktur",
	Long:  `Pobiera listę faktur z KSeF dla podanego profilu.`,
	Annotations: map[string]string{
		"NipRequired": "true", // This annotation can be used in the root command's PersistentPreRunE to enforce NIP check for this subcommand
	},
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("fetching invoices")
		query, err := cmdToQuery(cmd)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Błąd tworzenia zapytania:", err)
			slog.Error("creating invoice query failed", "err", err)
			return
		}
		page, err := cmdToPage(cmd)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Błąd tworzenia paginacji:", err)
			slog.Error("creating invoice page failed", "err", err)
			return
		}
		slog.Debug("invoice query and page created", "query", query, "page", page)
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

	invoiceListCmd.Flags().String("from", time.Now().AddDate(0, -3, 0).Format("2006-01-02"), "Data początkowa (format: RRRR-MM-DD)")
	invoiceListCmd.Flags().String("to", time.Now().Format("2006-01-02"), "Data końcowa (format: RRRR-MM-DD)")
	invoiceListCmd.Flags().String("subject", "subject2", "Typ podmiotu (subject1, subject2, subject3, authorized)")
	invoiceListCmd.Flags().Int32("pagesize", 100, "Liczba faktur na stronę")
	invoiceListCmd.Flags().Int32("pageoffset", 0, "Numer strony (offset)")
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
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Data", "Faktura", "Brutto", "Netto", "Kontrahent"})
	for _, invoice := range invoices {
		t.AppendRow(table.Row{
			invoice.PermanentStorageDate.Format("2006-01-02"),
			invoice.InvoiceNumber,
			fmt.Sprintf("%.2f", invoice.GrossAmount),
			fmt.Sprintf("%.2f", invoice.NetAmount),
			invoice.SellerName,
		})
	}
	t.Render()
}

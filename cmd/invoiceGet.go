package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	//"github.com/torgiren/ksef-cli/pkg/ksef"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// invoiceGetCmd represents the invoiceGet command
var invoiceGetCmd = &cobra.Command{
	Use:        "get [flags] <invoiceNumber>",
	Short:      "Pobierz szczegóły konkretnej faktury z KSeF",
	Long:       `Pobiera szczegóły konkretnej faktury z KSeF dla podanego profilu.`,
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"invoiceNumber"},
	Annotations: map[string]string{
		"NipRequired": "true",
	},
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("invoiceGet command executed")
		invoiceNumber := args[0]
		slog.Info("fetching invoice details", "invoiceNumber", invoiceNumber)

		invoice, err := client.GetInvoiceDetails(cmd.Context(), invoiceNumber)
		if err != nil {
			slog.Error("failed to fetch invoice details", "error", err)
			fmt.Printf("Błąd podczas pobierania szczegółów faktury: %v\n", err)
			return
		}

		style := table.StyleDouble

		th := table.NewWriter()
		th.SetOutputMirror(os.Stdout)
		th.SetStyle(style)
		th.AppendRow(table.Row{fmt.Sprintf("FAKTURA VAT  │ %s | %s", invoice.InvoiceNumber, invoice.DateOfIssue)})
		th.AppendRow(table.Row{fmt.Sprintf("KSeF: %s", invoice.KsefNumber)})
		th.AppendRow(table.Row{fmt.Sprintf("Data dostawy/sprzedaży: %s", invoice.Content.DateOfSale)})
		th.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, Align: text.AlignCenter, WidthMin: 60},
		})
		th.Render()

		tp := table.NewWriter()
		tp.SetOutputMirror(os.Stdout)
		tp.SetStyle(style)
		tp.AppendHeader(table.Row{"SPRZEDAWCA", "NABYWCA"})
		tp.AppendRows([]table.Row{
			{invoice.Seller.Name, invoice.Buyer.Name},
			{fmt.Sprintf("NIP: %s", invoice.Seller.Nip), fmt.Sprintf("NIP: %s", invoice.Buyer.Nip)},
			{invoice.Seller.Address, invoice.Buyer.Address},
		})
		tp.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, Align: text.AlignLeft, WidthMin: 40, WidthMax: 60},
			{Number: 2, Align: text.AlignLeft, WidthMin: 40, WidthMax: 60},
		})
		tp.Render()

		ti := table.NewWriter()
		ti.SetOutputMirror(os.Stdout)
		ti.SetStyle(style)
		ti.AppendHeader(table.Row{"LP.", "NAZWA TOWARU/USŁUGI", "J.M.", "ILOŚĆ", "CENA NETTO", "WARTOŚĆ NETTO", "STAWKA VAT", "KWOTA VAT", "CENA BRUTTO", "WARTOŚĆ BRUTTO"})
		for _, item := range invoice.Content.Items {
			vatRate := fmt.Sprintf("%.0f%%", item.VatRate)
			if item.VatRate == -1 {
				vatRate = "ZW"
			}
			ti.AppendRow(table.Row{
				item.OrderNumber,
				item.Description,
				item.Unit,
				item.Quantity,
				item.NetPrice,
				item.NetAmount,
				vatRate,
				item.VatAmount,
				item.GrossPrice,
				item.GrossAmount,
			})
		}
		ti.AppendSeparator()
		if invoice.Content.NettoExempt > 0 {
			ti.AppendRow(table.Row{"", "", "", "", "", "", "", "", "Netto ZW", invoice.Content.NettoExempt})
		}
		if invoice.Content.Netto0 > 0 {
			ti.AppendRow(table.Row{"", "", "", "", "", "", "", "", "Netto 0%", invoice.Content.Netto0})
		}
		if invoice.Content.Netto5 > 0 {
			ti.AppendRow(table.Row{"", "", "", "", "", "", "", "", "Netto 5%", invoice.Content.Netto5})
			ti.AppendRow(table.Row{"", "", "", "", "", "", "", "", "Vat 5%", invoice.Content.Vat5})
		}
		if invoice.Content.Netto8 > 0 {
			ti.AppendRow(table.Row{"", "", "", "", "", "", "", "", "Netto 8%", invoice.Content.Netto8})
			ti.AppendRow(table.Row{"", "", "", "", "", "", "", "", "Vat 8%", invoice.Content.Vat8})
		}
		if invoice.Content.Netto23 > 0 {
			ti.AppendRow(table.Row{"", "", "", "", "", "", "", "", "Netto 23%", invoice.Content.Netto23})
			ti.AppendRow(table.Row{"", "", "", "", "", "", "", "", "Vat 23%", invoice.Content.Vat23})
		}

		ti.AppendSeparator()
		ti.AppendRow(table.Row{"", "", "", "", "", "", "", "", "Razem Brutto", invoice.Content.GrossAmount})
		ti.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, Align: text.AlignCenter, WidthMin: 5},
			{Number: 2, Align: text.AlignLeft, WidthMin: 30},
			{Number: 3, Align: text.AlignCenter, WidthMin: 5},
			{Number: 4, Align: text.AlignRight, WidthMin: 10},
			{Number: 5, Align: text.AlignRight, WidthMin: 15},
			{Number: 6, Align: text.AlignRight, WidthMin: 15},
			{Number: 7, Align: text.AlignCenter, WidthMin: 10},
			{Number: 8, Align: text.AlignRight, WidthMin: 15},
			{Number: 9, Align: text.AlignRight, WidthMin: 15},
			{Number: 10, Align: text.AlignRight, WidthMin: 15},
		})
		ti.Render()

		tp = table.NewWriter()
		tp.SetOutputMirror(os.Stdout)
		tp.SetStyle(style)
		tp.AppendRow(table.Row{"Termin płatności: " + invoice.Content.Payment.Date})
		tp.AppendRow(table.Row{"Sposób płatności: " + invoice.Content.Payment.Method})
		tp.AppendRow(table.Row{"Konto bankowe: " + invoice.Content.Payment.BankAccount})
		tp.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, Align: text.AlignLeft, WidthMin: 40, WidthMax: 60},
		})
		tp.Render()

	},
}

func init() {
	invoiceCmd.AddCommand(invoiceGetCmd)
}

package ksef

import (
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/torgiren/ksef-cli/internal/ksefapi"
)

type Invoice struct {
	GrossAmount          float64   `json:"gross_amount"`
	NetAmount            float64   `json:"net_amount"`
	InvoiceNumber        string    `json:"invoice_number"`
	KsefNumber           string    `json:"ksef_number"`
	PermanentStorageDate time.Time `json:"permanent_storage_date"`
	SellerNIP            string    `json:"seller_nip"`
	SellerName           string    `json:"seller_name"`
}

type InvoiceDetailsParty struct {
	Nip     string `json:"nip"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type InvoiceDetailsElement struct {
	OrderNumber string  `json:"orderNumber"`
	Description string  `json:"description"`
	Unit        string  `json:"unit"`
	Quantity    float64 `json:"quantity"`
	NetPrice    float64 `json:"netPrice"`
	GrossPrice  float64 `json:"grossPrice"`
	NetAmount   float64 `json:"netAmount"`
	GrossAmount float64 `json:"grossAmount"`
	VatAmount   float64 `json:"vatAmount"`
	VatRate     float64 `json:"vatRate"`
}

type InvoiceDetailsPayment struct {
	Method      string `json:"method"`
	Date        string `json:"date"`
	BankAccount string `json:"bankAccount"`
}

type InvoiceDetailContent struct {
	Netto0      float64 `json:"netto0"`
	Vat0        float64 `json:"vat0"`
	Netto5      float64 `json:"netto5"`
	Vat5        float64 `json:"vat5"`
	Netto8      float64 `json:"netto8"`
	Vat8        float64 `json:"vat8"`
	NettoExempt float64 `json:"nettoExempt"`
	Netto23     float64 `json:"netto23"`
	Vat23       float64 `json:"vat23"`

	GrossAmount float64                 `json:"grossAmount"`
	DateOfSale  string                  `json:"dateOfSale"`
	Items       []InvoiceDetailsElement `json:"elements"`
	Payment     InvoiceDetailsPayment   `json:"payment"`
}

type InvoiceDetails struct {
	Seller        InvoiceDetailsParty  `json:"seller"`
	Buyer         InvoiceDetailsParty  `json:"buyer"`
	InvoiceNumber string               `json:"invoiceNumber"`
	KsefNumber    string               `json:"ksefNumber"`
	DateOfIssue   string               `json:"dateOfIssue"`
	Content       InvoiceDetailContent `json:"content"`
}

func subjectTypeToAPI(subjectType SubjectType) ksefapi.InvoiceQuerySubjectType {
	switch subjectType {
	case Subject1:
		return ksefapi.Subject1
	case Subject2:
		return ksefapi.Subject2
	case Subject3:
		return ksefapi.Subject3
	case SubjectAuthorized:
		return ksefapi.SubjectAuthorized
	default:
		return ksefapi.Subject2
	}
}

func (c *ksefClient) GetInvoices(ctx context.Context, query InvoiceQuery, page InvoicePage) ([]Invoice, error) {
	tokens := c.GetTokens()
	if tokens == nil || tokens.AccessToken == "" {
		return nil, fmt.Errorf("no access token available, please login first")
	}
	accessToken := tokens.AccessToken
	apiQuery := ksefapi.PostInvoicesQueryMetadataJSONRequestBody{
		SubjectType: subjectTypeToAPI(query.SubjectType),
		DateRange: ksefapi.InvoiceQueryDateRange{
			DateType: "PermanentStorage",
			From:     query.From,
			To:       &query.To,
		},
	}
	apiParams := &ksefapi.PostInvoicesQueryMetadataParams{
		PageSize:   &page.PageSize,
		PageOffset: &page.PageOffset,
	}
	slog.Log(ctx, LevelTrace, "API query body", "query", apiQuery, "params", apiParams)
	slog.Log(ctx, LevelSecret, "access token for invoices", "token", accessToken)
	invoices, err := c.api.PostInvoicesQueryMetadataWithResponse(ctx, apiParams, apiQuery, bearerTokenFn(accessToken))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch invoices: %w", err)
	}
	if invoices.StatusCode() == 400 {
		errorMessage := JSON400ToString(invoices.JSON400)
		return nil, fmt.Errorf("bad request: %s", errorMessage)
	}
	if invoices.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", invoices.StatusCode())
	}
	if invoices.JSON200 == nil {
		return nil, fmt.Errorf("empty response body")
	}

	var result []Invoice

	for _, item := range invoices.JSON200.Invoices {
		result = append(result, Invoice{
			GrossAmount:          item.GrossAmount,
			NetAmount:            item.NetAmount,
			InvoiceNumber:        item.InvoiceNumber,
			KsefNumber:           item.KsefNumber,
			PermanentStorageDate: item.PermanentStorageDate,
			SellerNIP:            item.Seller.Nip,
			SellerName:           *item.Seller.Name,
		})
	}
	slog.Debug("invoices fetched", "count", len(result))
	return result, nil
}

func (c *ksefClient) GetInvoiceDetailsRaw(ctx context.Context, invoiceNumber string) ([]byte, error) {
	tokens := c.GetTokens()
	if tokens == nil || tokens.AccessToken == "" {
		return nil, fmt.Errorf("no access token available, please login first")
	}
	accessToken := tokens.AccessToken
	slog.Log(ctx, LevelSecret, "access token for invoice details", "token", accessToken)
	invoiceResp, err := c.api.GetInvoicesKsefKsefNumberWithResponse(ctx, invoiceNumber, bearerTokenFn(accessToken))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch invoice details: %w", err)
	}
	if invoiceResp.StatusCode() == 400 {
		errorMessage := JSON400ToString(invoiceResp.JSON400)
		return nil, fmt.Errorf("bad request: %s", errorMessage)
	}
	if invoiceResp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", invoiceResp.StatusCode())
	}
	if invoiceResp.XML200 == nil {
		return nil, fmt.Errorf("empty response body")
	}
	return invoiceResp.Body, nil
}

func stringToFloat(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse float from string '%s': %w", s, err)
	}
	return f, nil
}

func paymentMethodToString(method string) string {
	switch method {
	case "1":
		return "Gotówka"
	case "2":
		return "Karta płatnicza"
	case "3":
		return "Bon"
	case "4":
		return "Czek"
	case "5":
		return "Kredyt"
	case "6":
		return "Przelew"
	case "7":
		return "Mobilna"
	default:
		return method
	}
}

func bankAccountFormat(account string) string {
	if len(account) == 26 {
		return fmt.Sprintf("%s %s %s %s %s %s", account[0:2], account[2:6], account[6:10], account[10:14], account[14:18], account[18:26])
	}
	return account
}

func invoiceXMLToDetails(invoiceXML []byte) (*InvoiceDetails, error) {
	var invoice FakturaXML
	err := xml.Unmarshal(invoiceXML, &invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML response: %w", err)
	}
	// Different formats. with Z and without
	issueDate, err := time.Parse("2006-01-02", invoice.Naglowek.DataWytworzeniaFa[:10])

	if err != nil {
		return nil, fmt.Errorf("failed to parse issue date: %w", err)
	}
	var dateOfSale string
	if invoice.Fa.DataDostawy != "" {
		dateOfSale = invoice.Fa.DataDostawy
	}
	if invoice.Fa.OkresFa.Od != "" && invoice.Fa.OkresFa.Do != "" {
		dateOfSale = fmt.Sprintf("od %s do %s", invoice.Fa.OkresFa.Od, invoice.Fa.OkresFa.Do)
	}
	items := make([]InvoiceDetailsElement, len(invoice.Fa.Wiersze))
	for i, item := range invoice.Fa.Wiersze {
		quantity, err := stringToFloat(item.Ilosc)
		if err != nil {
			return nil, fmt.Errorf("failed to parse quantity for item %d: %w", i+1, err)
		}
		netPrice, err := stringToFloat(item.Cena)
		if err != nil {
			return nil, fmt.Errorf("failed to parse net price for item %d: %w", i+1, err)
		}
		grossPrice, err := stringToFloat(item.CenaB)
		if err != nil {
			return nil, fmt.Errorf("failed to parse gross price for item %d: %w", i+1, err)
		}
		netAmount, err := stringToFloat(item.Netto)
		if err != nil {
			return nil, fmt.Errorf("failed to parse net amount for item %d: %w", i+1, err)
		}
		grossAmount, err := stringToFloat(item.Brutto)
		if err != nil {
			return nil, fmt.Errorf("failed to parse gross amount for item %d: %w", i+1, err)
		}
		vatAmount, err := stringToFloat(item.Vat)
		if err != nil {
			return nil, fmt.Errorf("failed to parse VAT amount for item %d: %w", i+1, err)
		}
		vatRate, err := stringToFloat(item.StVAT)
		if err != nil {
			return nil, fmt.Errorf("failed to parse VAT rate for item %d: %w", i+1, err)
		}
		items[i] = InvoiceDetailsElement{
			OrderNumber: item.Nr,
			Description: item.Nazwa,
			Unit:        item.Jm,
			Quantity:    quantity,
			NetPrice:    netPrice,
			GrossPrice:  grossPrice,
			NetAmount:   netAmount,
			GrossAmount: grossAmount,
			VatAmount:   vatAmount,
			VatRate:     vatRate,
		}
	}

	nettoExempt, err := stringToFloat(invoice.Fa.KwotaZW)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Netto ZW amount: %w", err)
	}
	netto0, err := stringToFloat(invoice.Fa.KwotaNetto0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Netto 0%% amount: %w", err)
	}
	netto5, err := stringToFloat(invoice.Fa.KwotaNetto5)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Netto 5%% amount: %w", err)
	}
	netto8, err := stringToFloat(invoice.Fa.KwotaNetto8)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Netto 8%% amount: %w", err)
	}
	netto23, err := stringToFloat(invoice.Fa.KwotaNetto23)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Netto 23%% amount: %w", err)
	}
	vat5, err := stringToFloat(invoice.Fa.KwotaVAT5)
	if err != nil {
		return nil, fmt.Errorf("failed to parse VAT 5%% amount: %w", err)
	}
	vat8, err := stringToFloat(invoice.Fa.KwotaVAT8)
	if err != nil {
		return nil, fmt.Errorf("failed to parse VAT 8%% amount: %w", err)
	}
	vat23, err := stringToFloat(invoice.Fa.KwotaVAT23)
	if err != nil {
		return nil, fmt.Errorf("failed to parse VAT 2%3% amount: %w", err)
	}
	grossAmount, err := stringToFloat(invoice.Fa.KwotaBrutto)
	if err != nil {
		return nil, fmt.Errorf("failed to parse gross amount: %w", err)
	}

	invoiceDetails := &InvoiceDetails{
		InvoiceNumber: invoice.Fa.NumerFaktury,
		KsefNumber:    "asd",
		DateOfIssue:   issueDate.Format("2006-01-02"),
		Seller: InvoiceDetailsParty{
			Nip:     invoice.Podmiot1.NIP,
			Name:    invoice.Podmiot1.Nazwa,
			Address: fmt.Sprintf("%s\n%s", invoice.Podmiot1.Adres.AdresL1, invoice.Podmiot1.Adres.AdresL2),
		},
		Buyer: InvoiceDetailsParty{
			Nip:     invoice.Podmiot2.NIP,
			Name:    invoice.Podmiot2.Nazwa,
			Address: fmt.Sprintf("%s\n%s", invoice.Podmiot2.Adres.AdresL1, invoice.Podmiot2.Adres.AdresL2),
		},
		Content: InvoiceDetailContent{
			DateOfSale:  dateOfSale,
			Items:       items,
			NettoExempt: nettoExempt,
			Netto0:      netto0,
			Netto5:      netto5,
			Netto8:      netto8,
			Netto23:     netto23,
			Vat5:        vat5,
			Vat8:        vat8,
			Vat23:       vat23,
			GrossAmount: grossAmount,
			Payment: InvoiceDetailsPayment{
				Method:      paymentMethodToString(invoice.Fa.Platnosc.FormaPlatnosci),
				Date:        invoice.Fa.Platnosc.TerminPlatnosci.Termin,
				BankAccount: bankAccountFormat(invoice.Fa.Platnosc.RachunekBankowy.NrRB),
			},
		},
	}
	return invoiceDetails, nil
}

func (c *ksefClient) GetInvoiceDetails(ctx context.Context, invoiceNumber string) (*InvoiceDetails, error) {
	invoiceRaw, err := c.GetInvoiceDetailsRaw(ctx, invoiceNumber)
	//fmt.Printf("Raw invoice XML:\n%s\n", string(invoiceRaw))
	if err != nil {
		return nil, fmt.Errorf("failed to get raw invoice details: %w", err)
	}
	slog.Debug("raw invoice details fetched", "length", len(invoiceRaw))
	invoiceDetails, err := invoiceXMLToDetails(invoiceRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse invoice XML: %w", err)
	}
	invoiceDetails.KsefNumber = invoiceNumber
	return invoiceDetails, nil
}

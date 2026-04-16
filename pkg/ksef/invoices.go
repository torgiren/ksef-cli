package ksef

import (
	"context"
	"fmt"
	"log/slog"
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

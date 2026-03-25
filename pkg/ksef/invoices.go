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

func (c *ksefClient) GetInvoices(ctx context.Context, query InvoiceQuery, page InvoicePage) ([]Invoice, error) {
	tokens := c.GetTokens()
	if tokens == nil || tokens.AccessToken == "" {
		return nil, fmt.Errorf("no access token available, please login first")
	}
	accessToken := tokens.AccessToken
	from := time.Now().AddDate(0, -1, 0)
	to := time.Now()
	apiQuery := ksefapi.PostInvoicesQueryMetadataJSONRequestBody{
		SubjectType: "Subject2",
		DateRange: ksefapi.InvoiceQueryDateRange{
			DateType: "PermanentStorage",
			From:     from,
			To:       &to,
		},
	}
	apiParams := &ksefapi.PostInvoicesQueryMetadataParams{
		PageSize:   &page.PageSize,
		PageOffset: &page.PageOffset,
	}
	slog.Debug("querying invoices", "from", from, "to", to, "pageSize", page.PageSize)
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

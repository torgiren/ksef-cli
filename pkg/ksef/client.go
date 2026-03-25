package ksef

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/torgiren/ksef-cli/internal/ksefapi"
)

type Tokens struct {
	AccessToken        string    `json:"access_token"`
	AccessTokenExpiry  time.Time `json:"access_token_expiry"`
	RefreshToken       string    `json:"refresh_token"`
	RefreshTokenExpiry time.Time `json:"refresh_token_expiry"`
	KsefToken          string    `json:"ksef_token,omitempty"`
}

type Client interface {
	Login(ctx context.Context, nip string, token string) (*Tokens, error)
	GetInvoices(ctx context.Context, query InvoiceQuery, page InvoicePage) ([]Invoice, error)
	RefreshTokens(ctx context.Context) (*Tokens, error)

	SetTokens(tokens *Tokens)
	GetTokens() *Tokens
}

type ksefClient struct {
	api *ksefapi.ClientWithResponses

	Tokens *Tokens
}

func NewClient(serverURL string) (Client, error) {
	slog.Debug("creating API client", "endpoint", serverURL)
	apiClient, err := ksefapi.NewClientWithResponses(serverURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	return &ksefClient{
		api: apiClient,
	}, nil
}

func (c *ksefClient) SetTokens(tokens *Tokens) {
	c.Tokens = tokens
	slog.Log(context.Background(), LevelSecret, "tokens set in client", "accessToken", tokens.AccessToken, "refreshToken", tokens.RefreshToken)
}

func (c *ksefClient) GetTokens() *Tokens {
	if c.Tokens == nil {
		slog.Warn("attempting to get tokens from client but no tokens are set")
		return nil
	}
	return c.Tokens
}

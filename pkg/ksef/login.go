package ksef

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/torgiren/ksef-cli/internal/ksefapi"
)

type challengeResponse struct {
	Challenge   string `json:"challenge"`
	TimestampMs int64  `json:"timestamp"`
}

func (c *ksefClient) Login(ctx context.Context, nip string, token string) (*Tokens, error) {
	challenge, err := c.getChallenge(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get challenge: %w", err)
	}
	slog.Log(ctx, LevelTrace, "challenge received", "challenge", challenge.Challenge, "timestampMs", challenge.TimestampMs)

	cert, err := c.getKsefTokenEncryptionKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}
	slog.Log(ctx, LevelSecret, "encryption certificate received", "cert", cert)

	encryptedToken, err := encryptToken(token, cert, challenge.TimestampMs)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt token: %w", err)
	}
	slog.Log(ctx, LevelSecret, "token encrypted", "encryptedToken", encryptedToken)

	authToken, err := c.getAuthToken(ctx, nip, encryptedToken, challenge.Challenge)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}
	slog.Log(ctx, LevelTrace, "auth token received", "referenceNumber", authToken.ReferenceNumber)

	slog.Info("waiting for KSeF authorization")
	err = c.waitForAuthorization(ctx, authToken)
	if err != nil {
		return nil, fmt.Errorf("authorization failed: %w", err)
	}
	tokens, err := c.redeemToken(ctx, authToken)
	if err != nil {
		return nil, fmt.Errorf("failed to redeem auth token: %w", err)
	}
	slog.Debug("login tokens received", "accessTokenExpiry", tokens.AccessTokenExpiry, "refreshTokenExpiry", tokens.RefreshTokenExpiry)
	slog.Log(ctx, LevelSecret, "login tokens received", "accessToken", tokens.AccessToken, "refreshToken", tokens.RefreshToken)
	slog.Info("login completed", "nip", nip)

	return tokens, nil
}

func (c *ksefClient) getChallenge(ctx context.Context) (challengeResponse, error) {
	challengeResp, err := c.api.PostAuthChallengeWithResponse(ctx)
	if err != nil {
		return challengeResponse{}, fmt.Errorf("failed to get challenge: %w", err)
	}
	if challengeResp.JSON200 == nil {
		return challengeResponse{}, fmt.Errorf("unexpected response format: %v", challengeResp)
	}

	// Extract the challenge string from the response

	challenge := challengeResponse{
		Challenge:   challengeResp.JSON200.Challenge,
		TimestampMs: challengeResp.JSON200.TimestampMs,
	}

	return challenge, nil
}

func (c *ksefClient) getKsefTokenEncryptionKey(ctx context.Context) ([]byte, error) {
	certs, err := c.api.GetSecurityPublicKeyCertificatesWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}
	if certs.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response format: %v", certs)
	}
	if len(*certs.JSON200) == 0 {
		return nil, fmt.Errorf("no certificates found in response")
	}

	cert, err := findEncryptionCert(certs.JSON200)
	if err != nil {
		return nil, fmt.Errorf("failed to find encryption certificate: %w", err)
	}
	return cert, nil
}

type authTokenResponse struct {
	Token           string
	ReferenceNumber string
}

func (c *ksefClient) getAuthToken(ctx context.Context, nip string, encryptedToken []byte, challange string) (authTokenResponse, error) {
	slog.Debug("requesting auth token", "nip", nip)
	authResp, err := c.api.PostAuthKsefTokenWithResponse(ctx, ksefapi.PostAuthKsefTokenJSONRequestBody{
		ContextIdentifier: ksefapi.AuthenticationContextIdentifier{
			Type:  ksefapi.AuthenticationContextIdentifierTypeNip,
			Value: nip,
		},
		EncryptedToken: encryptedToken,
		Challenge:      challange,
	})
	if err != nil {
		return authTokenResponse{}, fmt.Errorf("failed to get auth token: %w, status: %d", err, authResp.HTTPResponse.StatusCode)
	}
	if authResp.JSON400 != nil {
		return authTokenResponse{}, fmt.Errorf("bad request: %v", JSON400ToString(authResp.JSON400))
	}
	if authResp.JSON202 == nil {
		return authTokenResponse{}, fmt.Errorf("unexpected response format: %v", authResp)
	}
	tmpToken := authResp.JSON202.AuthenticationToken.Token
	slog.Log(ctx, LevelSecret, "temporary auth token received", "token", tmpToken)

	return authTokenResponse{
		Token:           tmpToken,
		ReferenceNumber: authResp.JSON202.ReferenceNumber,
	}, nil
}

func (c *ksefClient) waitForAuthorization(ctx context.Context, authToken authTokenResponse) error {
	maxRetries := 60
	for i := 0; i < maxRetries; i++ {
		authStatusResp, err := c.api.GetAuthReferenceNumberWithResponse(ctx, authToken.ReferenceNumber, bearerTokenFn(authToken.Token))
		if err != nil {
			return fmt.Errorf("failed to get auth token status: %w", err)
		}
		if authStatusResp.JSON200 == nil {
			return fmt.Errorf("unexpected response format: %v", authStatusResp)
		}
		slog.Debug("authorization status", "code", authStatusResp.JSON200.Status.Code, "description", authStatusResp.JSON200.Status.Description)
		slog.Log(ctx, LevelTrace, "authorization status details", "details", authStatusResp.JSON200.Status.Details)

		done, err := interpretAuthorizationStatus(authStatusResp.JSON200.Status.Code, authStatusResp.JSON200.Status.Description)
		if err != nil {
			return fmt.Errorf("authorization failed: %w", err)
		}
		if done {
			slog.Info("authorization successful")
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("authorization timed out after %d retries", maxRetries)
}

func (c *ksefClient) redeemToken(ctx context.Context, authToken authTokenResponse) (*Tokens, error) {
	tokens, err := c.api.PostAuthTokenRedeemWithResponse(ctx, bearerTokenFn(authToken.Token))
	if err != nil {
		return nil, fmt.Errorf("failed to redeem token: %w", err)
	}
	if tokens.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response format: %v", tokens)
	}
	tokensData := Tokens{
		AccessToken:        tokens.JSON200.AccessToken.Token,
		AccessTokenExpiry:  tokens.JSON200.AccessToken.ValidUntil,
		RefreshToken:       tokens.JSON200.RefreshToken.Token,
		RefreshTokenExpiry: tokens.JSON200.RefreshToken.ValidUntil,
	}
	slog.Debug("tokens redeemed", "accessTokenExpiry", tokensData.AccessTokenExpiry, "refreshTokenExpiry", tokensData.RefreshTokenExpiry)
	slog.Log(ctx, LevelSecret, "redeemed tokens", "accessToken", tokensData.AccessToken, "refreshToken", tokensData.RefreshToken)
	return &tokensData, nil
}

func (c *ksefClient) RefreshTokens(ctx context.Context) (*Tokens, error) {
	slog.Debug("refreshing access token")
	currentTokens := c.GetTokens()
	if currentTokens == nil {
		return nil, fmt.Errorf("no tokens available to refresh")
	}
	refreshResp, err := c.api.PostAuthTokenRefreshWithResponse(ctx, bearerTokenFn(currentTokens.RefreshToken))
	if err != nil {
		return nil, fmt.Errorf("failed to refresh tokens: %w", err)
	}
	if refreshResp.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response format: %v", refreshResp)
	}
	newTokens := &Tokens{
		AccessToken:        refreshResp.JSON200.AccessToken.Token,
		AccessTokenExpiry:  refreshResp.JSON200.AccessToken.ValidUntil,
		RefreshToken:       currentTokens.RefreshToken,
		RefreshTokenExpiry: currentTokens.RefreshTokenExpiry,
	}
	if currentTokens.KsefToken != "" {
		newTokens.KsefToken = currentTokens.KsefToken
	}
	c.SetTokens(newTokens)
	slog.Debug("access token refreshed", "accessTokenExpiry", newTokens.AccessTokenExpiry)
	slog.Log(ctx, LevelSecret, "refreshed tokens", "accessToken", newTokens.AccessToken, "refreshToken", newTokens.RefreshToken)
	return newTokens, nil
}

package ksef

import (
	"context"
	"net/http"
	"fmt"
	"crypto/rand"
	"crypto/x509"
	"crypto/sha256"
	"crypto/rsa"
	"log/slog"

	"github.com/torgiren/ksef-cli/internal/ksefapi"
)

func bearerTokenFn(token string) ksefapi.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}
}

func encryptToken(token string, cert []byte, timestamp int64) ([]byte, error) {
	x509Cert, err := x509.ParseCertificate(cert)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	pubKey, ok := x509Cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("certificate does not contain an RSA public key")
	}

	combinedData := []byte(token + "|" + fmt.Sprintf("%d", timestamp))
	slog.Log(context.Background(), LevelSecret, "data to encrypt", "data", string(combinedData))
	encryptedData, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, combinedData, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt token: %w", err)
	}

	return encryptedData, nil
}

func findEncryptionCert(certs *[]ksefapi.PublicKeyCertificate) ([]byte, error) {
	for _, cert := range *certs {
		if cert.Usage != nil {
			for _, usage := range cert.Usage {
				if usage == ksefapi.KsefTokenEncryption {
					return cert.Certificate, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("encryption certificate not found")
}

func interpretAuthorizationStatus(status int32, description string) (bool, error) {
	if status == 200 {
		return true, nil
	}
	if status != 100 {
		return false, fmt.Errorf("authorization failed with status %d: %s", status, description)
	}
	return false, nil
}

func JSON400ToString(errorResponse *ksefapi.ExceptionResponse) string {
	separator := ""
	error_message := ""
	for _, exception := range *errorResponse.Exception.ExceptionDetailList {
		for _, message := range *exception.Details {
			error_message += separator + message
			separator = ";"
		}
	}
	return error_message
}

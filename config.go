package main

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultIrmaServerURL = "http://localhost:8088"
	defaultCredentialID  = "irma-demo.IRMATube.member"
)

type config struct {
	irmaServerURL string
	irmaAPIToken  string
	credentialID  string
	publicKey     *rsa.PublicKey
}

func loadConfig() (config, error) {
	keyPath := os.Getenv("IRMA_SERVER_PUBLICKEY")
	if keyPath == "" {
		return config{}, errors.New("IRMA_SERVER_PUBLICKEY is not set")
	}

	pemBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return config{}, fmt.Errorf("read IRMA server public key: %w", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pemBytes)
	if err != nil {
		return config{}, fmt.Errorf("parse IRMA server public key: %w", err)
	}

	return config{
		irmaServerURL: strings.TrimSuffix(envOr("IRMA_SERVER_URL", defaultIrmaServerURL), "/"),
		irmaAPIToken:  os.Getenv("IRMA_SERVER_API_TOKEN"),
		credentialID:  envOr("IRMATUBE_CREDENTIAL_ID", defaultCredentialID),
		publicKey:     publicKey,
	}, nil
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}

	return fallback
}

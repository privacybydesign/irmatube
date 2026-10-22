package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"
)

const (
	issuanceContext   = "https://irma.app/ld/request/issuance/v2"
	disclosureContext = "https://irma.app/ld/request/disclosure/v2"

	sessionPath = "/session"

	headerAuthorization = "Authorization"
	headerContentType   = "Content-Type"
	contentTypeJSON     = "application/json"

	credentialValidityMonths = 6

	// Membership numbers have 11 digits and no leading zero.
	memberIDMin  = 10_000_000_000
	memberIDSpan = 90_000_000_000

	irmaRequestTimeout   = 15 * time.Second
	maxIrmaResponseBytes = 1 << 20
)

type issuanceRequest struct {
	Context     string              `json:"@context"`
	Credentials []credentialRequest `json:"credentials"`
}

type credentialRequest struct {
	Credential string            `json:"credential"`
	Validity   int64             `json:"validity"`
	Attributes map[string]string `json:"attributes"`
}

type disclosureRequest struct {
	Context  string       `json:"@context"`
	Disclose [][][]string `json:"disclose"`
}

// irmaClient starts sessions on the IRMA server. The session responses
// are returned unparsed, because the browser needs them.
type irmaClient struct {
	serverURL    string
	apiToken     string
	credentialID string
	http         *http.Client
}

func newIrmaClient(cfg config) *irmaClient {
	return &irmaClient{
		serverURL:    cfg.irmaServerURL,
		apiToken:     cfg.irmaAPIToken,
		credentialID: cfg.credentialID,
		http:         &http.Client{Timeout: irmaRequestTimeout},
	}
}

func (c *irmaClient) startIssuance(ctx context.Context) ([]byte, error) {
	id, err := newMemberID()
	if err != nil {
		return nil, err
	}

	return c.startSession(ctx, issuanceRequest{
		Context: issuanceContext,
		Credentials: []credentialRequest{{
			Credential: c.credentialID,
			Validity:   time.Now().AddDate(0, credentialValidityMonths, 0).Unix(),
			Attributes: map[string]string{typeAttribute: memberTypeRegular, idAttribute: id},
		}},
	})
}

// startVerification asks for the membership type and, when age is above
// zero, for an attribute that proves the holder is at least that old.
func (c *irmaClient) startVerification(ctx context.Context, age int) ([]byte, error) {
	disclose := [][][]string{{{membershipAttributeID(c.credentialID)}}}

	if age > 0 {
		var alternatives [][]string
		for _, id := range ageAttributeIDs(age) {
			alternatives = append(alternatives, []string{id})
		}

		disclose = append(disclose, alternatives)
	}

	return c.startSession(ctx, disclosureRequest{Context: disclosureContext, Disclose: disclose})
}

func (c *irmaClient) startSession(ctx context.Context, request any) ([]byte, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.serverURL+sessionPath, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerContentType, contentTypeJSON)
	if c.apiToken != "" {
		req.Header.Set(headerAuthorization, c.apiToken)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IRMA server answered %s", resp.Status)
	}

	return io.ReadAll(io.LimitReader(resp.Body, maxIrmaResponseBytes))
}

func newMemberID() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(memberIDSpan))
	if err != nil {
		return "", err
	}

	return n.Add(n, big.NewInt(memberIDMin)).String(), nil
}

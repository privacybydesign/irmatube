package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testCredentialID = "pbdf-staging.pbdf.yivitube"
	testAPIToken     = "secret-token"
	testSessionReply = `{"sessionPtr":{"u":"https://irma.example/irma/session/abc","irmaqr":"disclosing"},"token":"abc"}`
)

var (
	signingKey      = sync.OnceValue(newTestKey)
	otherSigningKey = sync.OnceValue(newTestKey)
)

func newTestKey() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	return key
}

func publicKeyPEM(key *rsa.PrivateKey) []byte {
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		panic(err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
}

// fakeIrma stands in for the IRMA server and records the session requests.
type fakeIrma struct {
	*httptest.Server
	status   int
	requests []recordedRequest
}

type recordedRequest struct {
	method        string
	path          string
	authorization string
	contentType   string
	body          map[string]any
}

func newFakeIrma(t *testing.T) *fakeIrma {
	t.Helper()

	f := &fakeIrma{status: http.StatusOK}
	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)

		f.requests = append(f.requests, recordedRequest{
			method:        r.Method,
			path:          r.URL.Path,
			authorization: r.Header.Get("Authorization"),
			contentType:   r.Header.Get("Content-Type"),
			body:          body,
		})

		w.WriteHeader(f.status)
		_, _ = io.WriteString(w, testSessionReply)
	}))
	t.Cleanup(f.Close)

	return f
}

// newTestHandler serves the real embedded static files and videos, and
// verifies tokens against signingKey.
func newTestHandler(t *testing.T, irmaURL string) http.Handler {
	t.Helper()

	cfg := config{
		irmaServerURL: irmaURL,
		irmaAPIToken:  testAPIToken,
		credentialID:  testCredentialID,
		publicKey:     &signingKey().PublicKey,
	}

	videos, err := loadVideos(videoFiles)
	if err != nil {
		t.Fatal(err)
	}

	static, err := fs.Sub(staticFiles, "static")
	if err != nil {
		t.Fatal(err)
	}

	return newServer(newIrmaClient(cfg), newVerifier(cfg, videos), static)
}

type attr struct {
	id  string
	raw any
}

func memberAttr(memberType string) attr {
	return attr{testCredentialID + ".type", memberType}
}

func ageAttr(id string) attr {
	return attr{id, "yes"}
}

// resultClaimsFor builds the claims of a valid, fresh disclosure result JWT.
func resultClaimsFor(attrs ...attr) jwt.MapClaims {
	now := time.Now()
	disclosed := []map[string]any{}
	for _, a := range attrs {
		disclosed = append(disclosed, map[string]any{"id": a.id, "rawvalue": a.raw, "status": "PRESENT"})
	}

	return jwt.MapClaims{
		"iat":         now.Unix(),
		"exp":         now.Add(2 * time.Minute).Unix(),
		"sub":         "disclosing_result",
		"token":       "abc",
		"status":      "DONE",
		"type":        "disclosing",
		"proofStatus": "VALID",
		"disclosed":   [][]map[string]any{disclosed},
	}
}

func signRS256(t *testing.T, key *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()

	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	if err != nil {
		t.Fatal(err)
	}

	return token
}

func do(h http.Handler, method, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	return rec
}

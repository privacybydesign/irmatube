package main

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	testEnvServerURL    = "IRMA_SERVER_URL"
	testEnvAPIToken     = "IRMA_SERVER_API_TOKEN"
	testEnvCredentialID = "IRMATUBE_CREDENTIAL_ID"
	testEnvPublicKey    = "IRMA_SERVER_PUBLICKEY"

	testDefaultIrmaURL = "http://localhost:8088"
)

func writeTempFile(t *testing.T, content []byte) string {
	t.Helper()

	name := filepath.Join(t.TempDir(), "key.pem")
	if err := os.WriteFile(name, content, 0o600); err != nil {
		t.Fatal(err)
	}

	return name
}

func clearEnv(t *testing.T) {
	t.Helper()

	for _, name := range []string{testEnvServerURL, testEnvAPIToken, testEnvCredentialID, testEnvPublicKey} {
		t.Setenv(name, "")
	}
}

func TestLoadConfig(t *testing.T) {
	clearEnv(t)
	t.Setenv(testEnvPublicKey, writeTempFile(t, publicKeyPEM(signingKey())))
	t.Setenv(testEnvServerURL, "https://irma.example/")
	t.Setenv(testEnvAPIToken, "token")
	t.Setenv(testEnvCredentialID, "pbdf.pbdf.yivitube")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.irmaServerURL != "https://irma.example" || cfg.irmaAPIToken != "token" || cfg.credentialID != "pbdf.pbdf.yivitube" {
		t.Errorf("cfg = %+v", cfg)
	}
	if !cfg.publicKey.Equal(&signingKey().PublicKey) {
		t.Errorf("public key was not loaded from IRMA_SERVER_PUBLICKEY")
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	clearEnv(t)
	t.Setenv(testEnvPublicKey, writeTempFile(t, publicKeyPEM(signingKey())))

	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.irmaServerURL != testDefaultIrmaURL || cfg.irmaAPIToken != "" || cfg.credentialID != "irma-demo.IRMATube.member" {
		t.Errorf("cfg = %+v", cfg)
	}
}

func TestLoadConfigNeedsPublicKey(t *testing.T) {
	clearEnv(t)

	tests := map[string]string{
		"unset":        "",
		"missing file": filepath.Join(t.TempDir(), "absent.pem"),
		"not a PEM":    writeTempFile(t, []byte("hello")),
		"garbage PEM":  writeTempFile(t, []byte("-----BEGIN PUBLIC KEY-----\nAAAA\n-----END PUBLIC KEY-----\n")),
	}

	for name, path := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv(testEnvPublicKey, path)

			if _, err := loadConfig(); err == nil {
				t.Errorf("loadConfig succeeded without a usable public key")
			}
		})
	}
}

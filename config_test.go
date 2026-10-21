package main

import (
	"os"
	"path/filepath"
	"testing"
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

	for _, name := range []string{"IRMA_SERVER_URL", "IRMA_SERVER_API_TOKEN", "IRMATUBE_CREDENTIAL_ID", "IRMA_SERVER_PUBLICKEY"} {
		t.Setenv(name, "")
	}
}

func TestLoadConfig(t *testing.T) {
	clearEnv(t)
	t.Setenv("IRMA_SERVER_PUBLICKEY", writeTempFile(t, publicKeyPEM(signingKey())))
	t.Setenv("IRMA_SERVER_URL", "https://irma.example/")
	t.Setenv("IRMA_SERVER_API_TOKEN", "token")
	t.Setenv("IRMATUBE_CREDENTIAL_ID", "pbdf.pbdf.yivitube")

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
	t.Setenv("IRMA_SERVER_PUBLICKEY", writeTempFile(t, publicKeyPEM(signingKey())))

	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.irmaServerURL != "http://localhost:8088" || cfg.irmaAPIToken != "" || cfg.credentialID != "irma-demo.IRMATube.member" {
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
			t.Setenv("IRMA_SERVER_PUBLICKEY", path)

			if _, err := loadConfig(); err == nil {
				t.Errorf("loadConfig succeeded without a usable public key")
			}
		})
	}
}

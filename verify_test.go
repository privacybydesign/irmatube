package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const verifyPath = "/api/verifysession"

func verifyBody(t *testing.T, token, videoID string) string {
	t.Helper()

	body, err := json.Marshal(map[string]string{"token": token, "videoid": videoID})
	if err != nil {
		t.Fatal(err)
	}

	return string(body)
}

func decodeVerify(t *testing.T, rec *httptest.ResponseRecorder) verifyResponse {
	t.Helper()

	if ct := rec.Header().Get(headerContentType); ct != contentTypeJSON {
		t.Errorf("Content-Type = %q, want application/json (body %q)", ct, rec.Body.String())
	}

	var resp verifyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("body is not JSON: %q", rec.Body.String())
	}

	return resp
}

func TestVerifyGrantsAccess(t *testing.T) {
	over18 := "pbdf.pbdf.ageLimits.over18"

	tests := []struct {
		name      string
		video     string
		youtubeID string
		attrs     []attr
	}{
		{"regular member, no age limit", "up", "HWEW_qTLSEE", []attr{memberAttr("regular")}},
		{"premium member, no age limit", "up", "HWEW_qTLSEE", []attr{memberAttr("premium")}},
		{"member with 18+ proof", "django", "_iH0UBYDI4g", []attr{memberAttr("regular"), ageAttr(over18)}},
		{"age answer in Dutch", "django", "_iH0UBYDI4g", []attr{memberAttr("regular"), {over18, "ja"}}},
		{"age answer in capitals", "django", "_iH0UBYDI4g", []attr{memberAttr("regular"), {over18, "Yes"}}},
		{"nijmegen age proof", "django", "_iH0UBYDI4g", []attr{memberAttr("regular"), ageAttr("pbdf.nijmegen.ageLimits.over18")}},
		{"gemeente age proof", "django", "_iH0UBYDI4g", []attr{memberAttr("regular"), ageAttr("pbdf.gemeente.personalData.over18")}},
		{"amsterdam id card age proof", "django", "_iH0UBYDI4g", []attr{memberAttr("regular"), ageAttr("pbdf.pilot-amsterdam.idcard.over18")}},
		{"amsterdam passport age proof", "django", "_iH0UBYDI4g", []attr{memberAttr("regular"), ageAttr("pbdf.pilot-amsterdam.passport.over18")}},
		{"age proof listed before membership", "django", "_iH0UBYDI4g", []attr{ageAttr(over18), memberAttr("regular")}},
	}

	h := newTestHandler(t, testUnusedIrmaURL)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := signRS256(t, signingKey(), resultClaimsFor(tt.attrs...))

			rec := do(h, http.MethodPost, verifyPath, verifyBody(t, token, tt.video))

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200 (body %q)", rec.Code, rec.Body.String())
			}
			resp := decodeVerify(t, rec)
			if !resp.Success || resp.YoutubeID != tt.youtubeID {
				t.Errorf("response = %+v, want success with youtubeId %q", resp, tt.youtubeID)
			}
		})
	}
}

func TestVerifyDeniesAccess(t *testing.T) {
	over18 := "pbdf.pbdf.ageLimits.over18"
	member := memberAttr("regular")

	tests := []struct {
		name  string
		video string
		token func(t *testing.T) string
	}{
		{"not a member", "up", func(t *testing.T) string {
			return signRS256(t, signingKey(), resultClaimsFor(attr{"irma-demo.other.attr", "x"}))
		}},
		{"unknown membership type", "up", func(t *testing.T) string {
			return signRS256(t, signingKey(), resultClaimsFor(memberAttr("guest")))
		}},
		{"membership type is null", "up", func(t *testing.T) string {
			return signRS256(t, signingKey(), resultClaimsFor(attr{testCredentialID + ".type", nil}))
		}},
		{"no age proof for 18+ video", "django", func(t *testing.T) string {
			return signRS256(t, signingKey(), resultClaimsFor(member))
		}},
		{"age answer is no", "django", func(t *testing.T) string {
			return signRS256(t, signingKey(), resultClaimsFor(member, attr{over18, "no"}))
		}},
		{"age answer is null", "django", func(t *testing.T) string {
			return signRS256(t, signingKey(), resultClaimsFor(member, attr{over18, nil}))
		}},
		{"over 16 proof for 18+ video", "django", func(t *testing.T) string {
			return signRS256(t, signingKey(), resultClaimsFor(member, ageAttr("pbdf.pbdf.ageLimits.over16")))
		}},
		{"age proof from unknown issuer", "django", func(t *testing.T) string {
			return signRS256(t, signingKey(), resultClaimsFor(member, ageAttr("irma-demo.evil.ageLimits.over18")))
		}},
		{"no proofStatus", "up", func(t *testing.T) string {
			claims := resultClaimsFor(member)
			delete(claims, "proofStatus")
			return signRS256(t, signingKey(), claims)
		}},
		{"proofStatus INVALID", "up", func(t *testing.T) string {
			claims := resultClaimsFor(member)
			claims["proofStatus"] = "INVALID"
			return signRS256(t, signingKey(), claims)
		}},
		{"proofStatus in lower case", "up", func(t *testing.T) string {
			claims := resultClaimsFor(member)
			claims["proofStatus"] = "valid"
			return signRS256(t, signingKey(), claims)
		}},
		{"expired token", "up", func(t *testing.T) string {
			claims := resultClaimsFor(member)
			claims["iat"] = time.Now().Add(-10 * time.Minute).Unix()
			claims["exp"] = time.Now().Add(-5 * time.Minute).Unix()
			return signRS256(t, signingKey(), claims)
		}},
		{"expired 90 s ago, past the leeway", "up", func(t *testing.T) string {
			claims := resultClaimsFor(member)
			claims["exp"] = time.Now().Add(-90 * time.Second).Unix()
			return signRS256(t, signingKey(), claims)
		}},
		{"no expiry", "up", func(t *testing.T) string {
			claims := resultClaimsFor(member)
			delete(claims, "exp")
			return signRS256(t, signingKey(), claims)
		}},
		{"issued in the future", "up", func(t *testing.T) string {
			claims := resultClaimsFor(member)
			claims["iat"] = time.Now().Add(10 * time.Minute).Unix()
			return signRS256(t, signingKey(), claims)
		}},
		{"signed with another key", "up", func(t *testing.T) string {
			return signRS256(t, otherSigningKey(), resultClaimsFor(member))
		}},
		{"algorithm none", "up", func(t *testing.T) string {
			token, err := jwt.NewWithClaims(jwt.SigningMethodNone, resultClaimsFor(member)).SignedString(jwt.UnsafeAllowNoneSignatureType)
			if err != nil {
				t.Fatal(err)
			}
			return token
		}},
		{"HS256 keyed with the public key", "up", func(t *testing.T) string {
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, resultClaimsFor(member)).SignedString(publicKeyPEM(signingKey()))
			if err != nil {
				t.Fatal(err)
			}
			return token
		}},
		{"RS384 with the right key", "up", func(t *testing.T) string {
			token, err := jwt.NewWithClaims(jwt.SigningMethodRS384, resultClaimsFor(member)).SignedString(signingKey())
			if err != nil {
				t.Fatal(err)
			}
			return token
		}},
		{"not a JWT", "up", func(*testing.T) string { return "not-a-jwt" }},
	}

	h := newTestHandler(t, testUnusedIrmaURL)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(h, http.MethodPost, verifyPath, verifyBody(t, tt.token(t), tt.video))

			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403 (body %q)", rec.Code, rec.Body.String())
			}
			if resp := decodeVerify(t, rec); resp.Success || resp.YoutubeID != "" {
				t.Errorf("response = %+v, want success=false without youtubeId", resp)
			}
		})
	}
}

func TestVerifyTokenWithinLeeway(t *testing.T) {
	claims := resultClaimsFor(memberAttr("regular"))
	claims["exp"] = time.Now().Add(-30 * time.Second).Unix()
	token := signRS256(t, signingKey(), claims)

	rec := do(newTestHandler(t, testUnusedIrmaURL), http.MethodPost, verifyPath, verifyBody(t, token, "up"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rec.Code, rec.Body.String())
	}
}

func TestVerifyRejectsBadRequests(t *testing.T) {
	validToken := signRS256(t, signingKey(), resultClaimsFor(memberAttr("regular")))

	tests := []struct {
		name    string
		body    string
		status  int
		message string
	}{
		{"empty object", `{}`, http.StatusBadRequest, "No token provided"},
		{"missing token", `{"videoid":"up"}`, http.StatusBadRequest, "No token provided"},
		{"empty token", `{"token":"","videoid":"up"}`, http.StatusBadRequest, "No token provided"},
		{"missing videoid", `{"token":"anything"}`, http.StatusBadRequest, "No videoid provided"},
		{"empty videoid", `{"token":"anything","videoid":""}`, http.StatusBadRequest, "No videoid provided"},
		{"body is not JSON", `token=abc`, http.StatusBadRequest, "Invalid request body"},
		{"empty body", ``, http.StatusBadRequest, "Invalid request body"},
		{"token is not a string", `{"token":5,"videoid":"up"}`, http.StatusBadRequest, "Invalid request body"},
		{"videoid is not a string", `{"token":"x","videoid":["up"]}`, http.StatusBadRequest, "Invalid request body"},
		{"unknown video", verifyBody(t, validToken, "nope"), http.StatusBadRequest, ""},
		{"path traversal in videoid", verifyBody(t, validToken, "../data/demo-privkey"), http.StatusBadRequest, ""},
		{"video file extension", verifyBody(t, validToken, "up.json"), http.StatusBadRequest, ""},
	}

	h := newTestHandler(t, testUnusedIrmaURL)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(h, http.MethodPost, verifyPath, tt.body)

			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d (body %q)", rec.Code, tt.status, rec.Body.String())
			}
			resp := decodeVerify(t, rec)
			if resp.Success || resp.Message != tt.message {
				t.Errorf("response = %+v, want success=false and message %q", resp, tt.message)
			}
		})
	}
}

func TestVerifyVideoBeforeToken(t *testing.T) {
	rec := do(newTestHandler(t, testUnusedIrmaURL), http.MethodPost, verifyPath, verifyBody(t, "not-a-jwt", "nope"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestVerifyOnlyAcceptsPost(t *testing.T) {
	h := newTestHandler(t, testUnusedIrmaURL)

	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		if rec := do(h, method, verifyPath, `{}`); rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s status = %d, want 405", method, rec.Code)
		}
	}

	if rec := do(h, http.MethodGet, verifyPath, ""); rec.Code == http.StatusOK {
		t.Errorf("GET status = 200, want an error")
	}
}

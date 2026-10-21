package main

import (
	"net/http"
	"regexp"
	"testing"
	"time"
)

var memberIDPattern = regexp.MustCompile(`^[1-9][0-9]{10}$`)

func startSession(t *testing.T, irma *fakeIrma, method, query string) (int, string) {
	t.Helper()

	rec := do(newTestHandler(t, irma.URL), method, "/api/session?"+query, "")

	return rec.Code, rec.Body.String()
}

func TestSessionIssuance(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		t.Run(method, func(t *testing.T) {
			irma := newFakeIrma(t)
			start := time.Now()

			status, body := startSession(t, irma, method, "type=issuance&0.123")

			if status != http.StatusOK || body != testSessionReply {
				t.Fatalf("got %d %q, want the IRMA server response passed through", status, body)
			}
			if len(irma.requests) != 1 {
				t.Fatalf("IRMA server got %d requests, want 1", len(irma.requests))
			}

			req := irma.requests[0]
			if req.method != http.MethodPost || req.path != "/session" {
				t.Errorf("IRMA request = %s %s, want POST /session", req.method, req.path)
			}
			if req.authorization != testAPIToken {
				t.Errorf("Authorization = %q, want the API token", req.authorization)
			}
			if req.contentType != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", req.contentType)
			}
			if req.body["@context"] != "https://irma.app/ld/request/issuance/v2" {
				t.Errorf("@context = %v", req.body["@context"])
			}

			credentials := req.body["credentials"].([]any)
			if len(credentials) != 1 {
				t.Fatalf("got %d credentials, want 1", len(credentials))
			}
			credential := credentials[0].(map[string]any)
			attributes := credential["attributes"].(map[string]any)
			if credential["credential"] != testCredentialID {
				t.Errorf("credential = %v, want %s", credential["credential"], testCredentialID)
			}
			if attributes["type"] != "regular" {
				t.Errorf("type = %v, want regular", attributes["type"])
			}
			if id, _ := attributes["id"].(string); !memberIDPattern.MatchString(id) {
				t.Errorf("id = %v, want 11 digits without leading zero", attributes["id"])
			}

			validity := time.Unix(int64(credential["validity"].(float64)), 0)
			want := start.AddDate(0, 6, 0)
			if validity.Before(want.Add(-time.Minute)) || validity.After(want.Add(time.Minute)) {
				t.Errorf("validity = %v, want about %v", validity, want)
			}
		})
	}
}

func TestSessionIssuesDifferentMemberIDs(t *testing.T) {
	irma := newFakeIrma(t)

	startSession(t, irma, http.MethodGet, "type=issuance")
	startSession(t, irma, http.MethodGet, "type=issuance")

	ids := map[any]bool{}
	for _, req := range irma.requests {
		credential := req.body["credentials"].([]any)[0].(map[string]any)
		ids[credential["attributes"].(map[string]any)["id"]] = true
	}
	if len(ids) != 2 {
		t.Errorf("got %d distinct ids in 2 sessions, want 2", len(ids))
	}
}

func TestSessionVerification(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantAge    string
		wantAnyAge bool
	}{
		{"no age", "type=verification", "", false},
		{"empty age", "type=verification&age=", "", false},
		{"age zero", "type=verification&age=0", "", false},
		{"age 16", "type=verification&age=16", "16", true},
		{"age 18 with cache buster", "type=verification&age=18&0.5", "18", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			irma := newFakeIrma(t)

			status, body := startSession(t, irma, http.MethodGet, tt.query)

			if status != http.StatusOK || body != testSessionReply {
				t.Fatalf("got %d %q, want the IRMA server response passed through", status, body)
			}
			req := irma.requests[0]
			if req.body["@context"] != "https://irma.app/ld/request/disclosure/v2" {
				t.Errorf("@context = %v", req.body["@context"])
			}

			disclose := req.body["disclose"].([]any)
			membership := disclose[0].([]any)
			if len(membership) != 1 || membership[0].([]any)[0] != testCredentialID+".type" {
				t.Errorf("first disclose condition = %v, want only %s.type", membership, testCredentialID)
			}

			if !tt.wantAnyAge {
				if len(disclose) != 1 {
					t.Errorf("disclose = %v, want membership only", disclose)
				}
				return
			}

			if len(disclose) != 2 {
				t.Fatalf("disclose = %v, want membership and age", disclose)
			}
			var got []string
			for _, alternative := range disclose[1].([]any) {
				got = append(got, alternative.([]any)[0].(string))
			}
			want := []string{
				"pbdf.pbdf.ageLimits.over" + tt.wantAge,
				"pbdf.nijmegen.ageLimits.over" + tt.wantAge,
				"pbdf.gemeente.personalData.over" + tt.wantAge,
				"pbdf.pilot-amsterdam.idcard.over" + tt.wantAge,
				"pbdf.pilot-amsterdam.passport.over" + tt.wantAge,
			}
			if len(got) != len(want) {
				t.Fatalf("age alternatives = %v, want %v", got, want)
			}
			for i := range want {
				if got[i] != want[i] {
					t.Errorf("age alternative %d = %s, want %s", i, got[i], want[i])
				}
			}
		})
	}
}

func TestSessionRejectsBadRequests(t *testing.T) {
	queries := []string{
		"",
		"type=",
		"type=unknown",
		"type=Issuance",
		"type=verification&age=16a",
		"type=verification&age=-1",
		"type=verification&age=%2B16",
		"type=verification&age=1.5",
		"type=verification&age=16%20",
		"type=verification&age=%3Cscript%3E",
		"type=verification&age=999",
		"type=verification&age=99999999999999999999",
	}

	for _, query := range queries {
		t.Run(query, func(t *testing.T) {
			irma := newFakeIrma(t)

			status, _ := startSession(t, irma, http.MethodGet, query)

			if status != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", status)
			}
			if len(irma.requests) != 0 {
				t.Errorf("IRMA server got %d requests, want none", len(irma.requests))
			}
		})
	}
}

func TestSessionFailsWhenIrmaServerFails(t *testing.T) {
	irma := newFakeIrma(t)
	irma.status = http.StatusInternalServerError

	status, body := startSession(t, irma, http.MethodGet, "type=issuance")

	if status != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", status)
	}
	if body == testSessionReply {
		t.Errorf("body passed through an IRMA server error")
	}
}

func TestSessionFailsWhenIrmaServerIsDown(t *testing.T) {
	irma := newFakeIrma(t)
	irma.Close()

	status, _ := startSession(t, irma, http.MethodGet, "type=verification")

	if status != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", status)
	}
}

func TestSessionOmitsEmptyAPIToken(t *testing.T) {
	irma := newFakeIrma(t)
	cfg := config{irmaServerURL: irma.URL, credentialID: testCredentialID}

	if _, err := newIrmaClient(cfg).startIssuance(t.Context()); err != nil {
		t.Fatal(err)
	}

	if got := irma.requests[0].authorization; got != "" {
		t.Errorf("Authorization = %q, want none", got)
	}
}

package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"strconv"
)

const (
	paramType = "type"
	paramAge  = "age"

	sessionTypeIssuance     = "issuance"
	sessionTypeVerification = "verification"

	maxRequestBytes = 1 << 20
)

type server struct {
	irma     *irmaClient
	verifier *verifier
	static   fs.FS
}

type verifyRequest struct {
	Token   string `json:"token"`
	VideoID string `json:"videoid"`
}

type verifyResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	YoutubeID string `json:"youtubeId,omitempty"`
}

func newServer(irma *irmaClient, verifier *verifier, static fs.FS) http.Handler {
	s := &server{irma: irma, verifier: verifier, static: static}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/session", s.handleSession)
	mux.HandleFunc("POST /api/session", s.handleSession)
	mux.HandleFunc("POST /api/verifysession", s.handleVerify)
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.Handle("GET /", http.FileServerFS(filesOnly{s.static}))

	return mux
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFileFS(w, r, s.static, "index.html")
}

func (s *server) handleSession(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var session []byte
	var err error
	switch query.Get(paramType) {
	case sessionTypeIssuance:
		session, err = s.irma.startIssuance(r.Context())
	case sessionTypeVerification:
		age, parseErr := parseAge(query.Get(paramAge))
		if parseErr != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		session, err = s.irma.startVerification(r.Context(), age)
	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Printf("start session: %v", err)
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	if _, err := w.Write(session); err != nil {
		log.Printf("write session: %v", err)
	}
}

func (s *server) handleVerify(w http.ResponseWriter, r *http.Request) {
	var req verifyRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBytes)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, verifyResponse{Message: "Invalid request body"})
		return
	}
	if req.Token == "" {
		writeJSON(w, http.StatusBadRequest, verifyResponse{Message: "No token provided"})
		return
	}
	if req.VideoID == "" {
		writeJSON(w, http.StatusBadRequest, verifyResponse{Message: "No videoid provided"})
		return
	}

	youtubeID, err := s.verifier.youtubeID(req.Token, req.VideoID)
	if errors.Is(err, errUnknownVideo) {
		writeJSON(w, http.StatusBadRequest, verifyResponse{})
		return
	}
	if err != nil {
		log.Printf("verify session: %v", err)
		writeJSON(w, http.StatusForbidden, verifyResponse{})
		return
	}

	writeJSON(w, http.StatusOK, verifyResponse{Success: true, YoutubeID: youtubeID})
}

// parseAge accepts digits only. An empty or zero age means no age check.
func parseAge(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}

	age, err := strconv.ParseUint(raw, 10, 8)

	return int(age), err
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write response: %v", err)
	}
}

// filesOnly hides directories, so that the file server never lists one.
type filesOnly struct{ fs.FS }

func (f filesOnly) Open(name string) (fs.File, error) {
	file, err := f.FS.Open(name)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil || info.IsDir() {
		_ = file.Close()
		return nil, fs.ErrNotExist
	}

	return file, nil
}

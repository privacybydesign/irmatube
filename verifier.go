package main

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// The result JWT lives 120 s by default; the leeway only covers clock skew.
	jwtLeeway = 60 * time.Second

	proofStatusValid = "VALID"

	ageAnswerYes = "yes"
	ageAnswerJa  = "ja"
)

var (
	errUnknownVideo = errors.New("unknown video")
	errAccessDenied = errors.New("access denied")
)

// resultClaims are the claims of the result JWT of an IRMA disclosure session.
type resultClaims struct {
	jwt.RegisteredClaims
	ProofStatus string                 `json:"proofStatus"`
	Disclosed   [][]disclosedAttribute `json:"disclosed"`
}

type disclosedAttribute struct {
	ID       string  `json:"id"`
	RawValue *string `json:"rawvalue"`
}

// verifier decides whether a result JWT gives access to a video.
type verifier struct {
	publicKey    *rsa.PublicKey
	credentialID string
	videos       videoCatalog
	parser       *jwt.Parser
}

func newVerifier(cfg config, videos videoCatalog) *verifier {
	return &verifier{
		publicKey:    cfg.publicKey,
		credentialID: cfg.credentialID,
		videos:       videos,
		parser: jwt.NewParser(
			jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
			jwt.WithLeeway(jwtLeeway),
			jwt.WithExpirationRequired(),
			jwt.WithIssuedAt(),
		),
	}
}

// youtubeID returns the YouTube id of the video when the token proves
// membership and, if the video has an age limit, that the holder is old enough.
func (v *verifier) youtubeID(token, videoID string) (string, error) {
	video, ok := v.videos[videoID]
	if !ok {
		return "", errUnknownVideo
	}

	claims, err := v.parse(token)
	if err != nil {
		return "", fmt.Errorf("%w: %w", errAccessDenied, err)
	}

	if claims.ProofStatus != proofStatusValid {
		return "", fmt.Errorf("%w: proof status %q", errAccessDenied, claims.ProofStatus)
	}
	if !v.isMember(claims.Disclosed) {
		return "", fmt.Errorf("%w: not a member", errAccessDenied)
	}
	if !isOldEnough(claims.Disclosed, video.AgeRestriction) {
		return "", fmt.Errorf("%w: age %d not proven", errAccessDenied, video.AgeRestriction)
	}

	return video.YoutubeID, nil
}

// parse checks the signature against the configured key only; nothing
// in the token chooses the key or the algorithm.
func (v *verifier) parse(token string) (*resultClaims, error) {
	claims := &resultClaims{}
	_, err := v.parser.ParseWithClaims(token, claims, func(*jwt.Token) (any, error) {
		return v.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func (v *verifier) isMember(disclosed [][]disclosedAttribute) bool {
	memberType := disclosedValue(disclosed, []string{membershipAttributeID(v.credentialID)})

	return memberType == memberTypeRegular || memberType == memberTypePremium
}

func isOldEnough(disclosed [][]disclosedAttribute, age int) bool {
	if age == 0 {
		return true
	}

	answer := strings.ToLower(disclosedValue(disclosed, ageAttributeIDs(age)))

	return answer == ageAnswerYes || answer == ageAnswerJa
}

// disclosedValue returns the value of the first disclosed attribute with
// one of the given ids. It returns "" when there is none or its value is null.
func disclosedValue(disclosed [][]disclosedAttribute, ids []string) string {
	for _, conjunction := range disclosed {
		for _, attr := range conjunction {
			if !slices.Contains(ids, attr.ID) {
				continue
			}
			if attr.RawValue == nil {
				return ""
			}

			return *attr.RawValue
		}
	}

	return ""
}

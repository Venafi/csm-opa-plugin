package internal

import (
	"crypto/rand"
	"encoding/json"

	"github.com/open-policy-agent/opa/bundle"
	"github.com/venafi/csm-opa-plugin/internal/jwx/jwa"
	"github.com/venafi/csm-opa-plugin/internal/jwx/jws"
)

// CustomSigner demonstrates a custom bundle signing implementation.
type CustomSigner struct{}

// GenerateSignedToken demonstrates how to implement the bundle.Signer interface,
// for the purpose of creating custom bundle signing. Note: In this example,
// no actual signing is taking place, it simply demonstrates how one could begin
// a custom signing implementation.
func (s *CustomSigner) GenerateSignedToken(files []bundle.FileInfo, sc *bundle.SigningConfig, keyID string) (string, error) {
	payload, err := generatePayload(files, sc, keyID)
	if err != nil {
		return "", err
	}

	var headers jws.StandardHeaders

	if err := headers.Set(jws.AlgorithmKey, jwa.SignatureAlgorithm(sc.Algorithm)); err != nil {
		return "", err
	}

	if keyID != "" {
		if err := headers.Set(jws.KeyIDKey, keyID); err != nil {
			return "", err
		}
	}

	hdr, err := json.Marshal(headers)
	if err != nil {
		return "", err
	}

	token, err := jws.SignLiteral(payload,
		jwa.SignatureAlgorithm(sc.Algorithm),
		sc.Key,
		hdr,
		rand.Reader)
	if err != nil {
		return "", err
	}
	return string(token), nil
}

func generatePayload(files []bundle.FileInfo, sc *bundle.SigningConfig, keyID string) ([]byte, error) {
	payload := make(map[string]interface{})
	payload["files"] = files

	if sc.ClaimsPath != "" {
		claims, err := sc.GetClaims()
		if err != nil {
			return nil, err
		}

		for claim, value := range claims {
			payload[claim] = value
		}
	} else if keyID != "" {
		// keyid claim is deprecated but include it for backwards compatibility.
		payload["keyid"] = keyID
	}
	return json.Marshal(payload)
}

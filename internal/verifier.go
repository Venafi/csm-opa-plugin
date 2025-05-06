package internal

import (
	"context"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/open-policy-agent/opa/bundle"
	"github.com/venafi/csm-opa-plugin/internal/jwx/jwa"
	"github.com/venafi/csm-opa-plugin/internal/jwx/jws"
	c "github.com/venafi/vsign/pkg/crypto"
	"github.com/venafi/vsign/pkg/vsign"
)

// CustomVerifier demonstrates a custom bundle verification implementation.
type CustomVerifier struct{}

// VerifyBundleSignature demonstrates how to implement the bundle.Verifier interface,
// for the purpose of creating custom bundle verification. Note: In this example,
// no actual verification is taking place, it simply demonstrates how one could
// begin a custom verification implementation.
func (v *CustomVerifier) VerifyBundleSignature(sc bundle.SignaturesConfig, bvc *bundle.VerificationConfig) (map[string]bundle.FileInfo, error) {
	files := make(map[string]bundle.FileInfo)

	if len(sc.Signatures) == 0 {
		return files, errors.New(".signatures.json: missing JWT (expected exactly one)")
	}

	if len(sc.Signatures) > 1 {
		return files, errors.New(".signatures.json: multiple JWTs not supported (expected exactly one)")
	}

	for _, token := range sc.Signatures {
		payload, err := verifyJWTSignature(token, bvc)
		if err != nil {
			return files, err
		}

		for _, file := range payload.Files {
			files[file.Name] = file
		}
	}
	return files, nil
}

func verifyJWTSignature(token string, bvc *bundle.VerificationConfig) (*bundle.DecodedSignature, error) {
	// decode JWT to check if the header specifies the key to use and/or if claims have the scope.

	parts, err := jws.SplitCompact(token)
	if err != nil {
		return nil, err
	}

	var decodedHeader []byte
	if decodedHeader, err = base64.RawURLEncoding.DecodeString(parts[0]); err != nil {
		return nil, fmt.Errorf("failed to base64 decode JWT headers: %w", err)
	}

	var hdr jws.StandardHeaders
	if err := json.Unmarshal(decodedHeader, &hdr); err != nil {
		return nil, fmt.Errorf("failed to parse JWT headers: %w", err)
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var ds bundle.DecodedSignature
	if err := json.Unmarshal(payload, &ds); err != nil {
		return nil, err
	}

	// check for the id of the key to use for JWT signature verification
	// first in the OPA config. If not found, then check the JWT kid.
	keyID := bvc.KeyID
	if keyID == "" {
		keyID = hdr.KeyID
	}
	if keyID == "" {
		// If header has no key id, check the deprecated key claim.
		keyID = ds.KeyID
	}

	if keyID == "" {
		return nil, errors.New("verification key ID is empty")
	}

	// now that we have the keyID, fetch the actual key
	keyConfig, err := bvc.GetPublicKey(keyID)
	if err != nil {
		return nil, err
	}
	pubKey, err := loadPublicKey(keyID)
	if err != nil {
		return nil, err
	}

	// verify JWT signature
	alg := jwa.SignatureAlgorithm(keyConfig.Algorithm)

	_, err = jws.Verify([]byte(token), alg, pubKey)
	if err != nil {
		return nil, err
	}

	// verify the scope
	scope := bvc.Scope
	if scope == "" {
		scope = keyConfig.Scope
	}

	if ds.Scope != scope {
		return nil, errors.New("scope mismatch")
	}
	return &ds, nil
}

func loadPublicKey(keyResourceID string) (crypto.PublicKey, error) {
	os.Setenv("VSIGN_PROJECT", "placeholder")
	vSignCfg, err := vsign.BuildConfig(context.Background(), "")

	if err != nil {
		return nil, fmt.Errorf("error building config")
	}

	vSignCfg.Project = keyResourceID
	connector, err := vsign.NewClient(&vSignCfg)

	if err != nil {
		return nil, fmt.Errorf("unable to connect to %s: %s", vSignCfg.ConnectorType, err)
	}

	e, err := connector.GetEnvironment()
	if err != nil {
		return nil, fmt.Errorf("unable to get environment: %s", err)
	}

	certs, err := c.ParseCertificates(e.CertificateChainData)
	return certs[0].PublicKey, nil

}

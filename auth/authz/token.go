package authz

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"strings"
	"time"

	"github.com/docker/distribution/registry/auth/token"
	"github.com/docker/libtrust"
)

type Token struct {
	UserName string
	Service  string
	Access   []*token.ResourceActions

	Issuer     string
	Expiration int64

	PrivateKey libtrust.PrivateKey
	SigAlg     string
}

// MakeToken makes a valid jwt token based on parms.
func (t *Token) MakeToken() (string, error) {

	joseHeader := &token.Header{
		Type:       "JWT",
		SigningAlg: t.SigAlg,
		KeyID:      t.PrivateKey.KeyID(),
	}

	jwtID, err := randString(16)
	if err != nil {
		return "", fmt.Errorf("error to generate jwt id: %s", err)
	}

	now := time.Now().Unix()

	claimSet := &token.ClaimSet{
		Issuer:     t.Issuer,
		Subject:    t.UserName,
		Audience:   t.Service,
		Expiration: now + t.Expiration,
		NotBefore:  now,
		IssuedAt:   now,
		JWTID:      jwtID,
		Access:     t.Access,
	}

	var joseHeaderBytes, claimSetBytes []byte

	if joseHeaderBytes, err = json.Marshal(joseHeader); err != nil {
		return "", fmt.Errorf("unable to marshal jose header: %s", err)
	}
	if claimSetBytes, err = json.Marshal(claimSet); err != nil {
		return "", fmt.Errorf("unable to marshal claim set: %s", err)
	}

	encodedJoseHeader := base64UrlEncode(joseHeaderBytes)
	encodedClaimSet := base64UrlEncode(claimSetBytes)
	payload := fmt.Sprintf("%s.%s", encodedJoseHeader, encodedClaimSet)

	var signatureBytes []byte
	if signatureBytes, _, err = t.PrivateKey.Sign(strings.NewReader(payload), crypto.SHA256); err != nil {
		return "", fmt.Errorf("unable to sign jwt payload: %s", err)
	}

	signature := base64UrlEncode(signatureBytes)
	tokenString := fmt.Sprintf("%s.%s", payload, signature)
	return tokenString, nil
}

func randString(length int) (string, error) {
	const alphanum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rb := make([]byte, length)
	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}
	for i, b := range rb {
		rb[i] = alphanum[int(b)%len(alphanum)]
	}
	return string(rb), nil
}

func base64UrlEncode(b []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}

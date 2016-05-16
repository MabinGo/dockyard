package token

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/docker/libtrust"

	"github.com/containerops/dockyard/middleware/auth"
	"github.com/containerops/dockyard/utils/setting"
)

// accessController implements the auth.AccessController interface.
type accessController struct {
	realm       string
	issuer      string
	service     string
	rootCerts   *x509.CertPool
	trustedKeys map[string]libtrust.PublicKey
}

// accessSet maps a typed, named resource to
// a set of actions requested or authorized.
type accessSet map[auth.Resource]actionSet

// init handles registering the token auth backend.
func init() {
	auth.Register("token", auth.AccessController(&accessController{}))
}

// newAccessController creates an accessController using the given options.
func (ac *accessController) InitFunc() error {
	ac.realm = setting.Realm
	ac.issuer = setting.Issue
	ac.service = setting.Service

	fp, err := os.Open(setting.Rootcertbundle)
	if err != nil {
		return err
	}
	defer fp.Close()

	rawCertBundle, err := ioutil.ReadAll(fp)
	if err != nil {
		return err
	}

	var rootCerts []*x509.Certificate
	pemBlock, rawCertBundle := pem.Decode(rawCertBundle)
	for pemBlock != nil {
		cert, err := x509.ParseCertificate(pemBlock.Bytes)
		if err != nil {
			return err
		}

		rootCerts = append(rootCerts, cert)

		pemBlock, rawCertBundle = pem.Decode(rawCertBundle)
	}

	if len(rootCerts) == 0 {
		return err
	}

	rootPool := x509.NewCertPool()
	trustedKeys := make(map[string]libtrust.PublicKey, len(rootCerts))
	for _, rootCert := range rootCerts {
		rootPool.AddCert(rootCert)
		pubKey, err := libtrust.FromCryptoPublicKey(crypto.PublicKey(rootCert.PublicKey))
		if err != nil {
			return err
		}
		trustedKeys[pubKey.KeyID()] = pubKey
	}

	ac.rootCerts = rootPool
	ac.trustedKeys = trustedKeys

	return nil
}

// Authorized handles checking whether the given request is authorized
// for actions on resources described by the given access items.
func (ac *accessController) Authorized(authorization string, accessItems ...auth.Access) (string, error) {
	challenge := &authChallenge{
		realm:     ac.realm,
		service:   ac.service,
		accessSet: newAccessSet(accessItems...),
	}

	parts := strings.Split(authorization, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		challenge.err = ErrTokenRequired
		return "", challenge
	}

	rawToken := parts[1]

	token, err := NewToken(rawToken)
	if err != nil {
		challenge.err = err
		return "", challenge
	}

	verifyOpts := VerifyOptions{
		TrustedIssuers:    []string{ac.issuer},
		AcceptedAudiences: []string{ac.service},
		Roots:             ac.rootCerts,
		TrustedKeys:       ac.trustedKeys,
	}

	if err = token.Verify(verifyOpts); err != nil {
		challenge.err = err
		return "", challenge
	}

	accessSet := token.accessSet()
	for _, access := range accessItems {
		if !accessSet.contains(access) {
			challenge.err = ErrInsufficientScope
			return "", challenge
		}
	}

	return token.Claims.Subject, nil
}

// newAccessSet constructs an accessSet from
// a variable number of auth.Access items.
func newAccessSet(accessItems ...auth.Access) accessSet {
	accessSet := make(accessSet, len(accessItems))

	for _, access := range accessItems {
		resource := auth.Resource{
			Type: access.Type,
			Name: access.Name,
		}

		set, exists := accessSet[resource]
		if !exists {
			set = newActionSet()
			accessSet[resource] = set
		}

		set.add(access.Action)
	}

	return accessSet
}

// contains returns whether or not the given access is in this accessSet.
func (s accessSet) contains(access auth.Access) bool {
	actionSet, ok := s[access.Resource]
	if ok {
		return actionSet.contains(access.Action)
	}

	return false
}

// scopeParam returns a collection of scopes which can
// be used for a WWW-Authenticate challenge parameter.
// See https://tools.ietf.org/html/rfc6750#section-3
func (s accessSet) scopeParam() string {
	scopes := make([]string, 0, len(s))

	for resource, actionSet := range s {
		actions := strings.Join(actionSet.keys(), ",")
		scopes = append(scopes, fmt.Sprintf("%s:%s:%s", resource.Type, resource.Name, actions))
	}

	return strings.Join(scopes, " ")
}

// Errors used and exported by this package.
var (
	ErrInsufficientScope = errors.New("insufficient scope")
	ErrTokenRequired     = errors.New("authorization token required")
)

// authChallenge implements the auth.Challenge interface.
type authChallenge struct {
	err       error
	realm     string
	service   string
	accessSet accessSet
}

var _ auth.Challenge = authChallenge{}

// Error returns the internal error string for this authChallenge.
func (ac authChallenge) Error() string {
	return ac.err.Error()
}

// Status returns the HTTP Response Status Code for this authChallenge.
func (ac authChallenge) Status() int {
	return http.StatusUnauthorized
}

// challengeParams constructs the value to be used in
// the WWW-Authenticate response challenge header.
// See https://tools.ietf.org/html/rfc6750#section-3
func (ac authChallenge) challengeParams() string {

	str := fmt.Sprintf("Bearer realm=%q,service=%q", ac.realm, ac.service)

	if scope := ac.accessSet.scopeParam(); scope != "" {
		str = fmt.Sprintf("%s,scope=%q", str, scope)
	}

	if ac.err == ErrInvalidToken || ac.err == ErrMalformedToken {
		str = fmt.Sprintf("%s,error=%q", str, "invalid_token")
	} else if ac.err == ErrInsufficientScope {
		str = fmt.Sprintf("%s,error=%q", str, "insufficient_scope")
	}

	return str
}

// SetChallenge sets the WWW-Authenticate value for the response.
func (ac authChallenge) SetHeaders(w http.ResponseWriter) {
	w.Header().Add("WWW-Authenticate", ac.challengeParams())
}

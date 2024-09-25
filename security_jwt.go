package goapi

import (
	"github.com/golang-jwt/jwt/v5"
	json "github.com/json-iterator/go"
	"time"
)

type (
	SigningMethod        = jwt.SigningMethod
	SigningMethodRSA     = jwt.SigningMethodRSA
	SigningMethodHMAC    = jwt.SigningMethodHMAC
	SigningMethodECDSA   = jwt.SigningMethodECDSA
	SigningMethodEd25519 = jwt.SigningMethodEd25519
)

var (
	// hmac
	SigningMethodHS256 *SigningMethodHMAC = jwt.SigningMethodHS256
	SigningMethodHS384 *SigningMethodHMAC = jwt.SigningMethodHS384
	SigningMethodHS512 *SigningMethodHMAC = jwt.SigningMethodHS512
	// rsa
	SigningMethodRS256 *SigningMethodRSA = jwt.SigningMethodRS256
	SigningMethodRS384 *SigningMethodRSA = jwt.SigningMethodRS384
	SigningMethodRS512 *SigningMethodRSA = jwt.SigningMethodRS512
	// ecdsa
	SigningMethodES256 *SigningMethodECDSA = jwt.SigningMethodES256
	SigningMethodES384 *SigningMethodECDSA = jwt.SigningMethodES384
	SigningMethodES512 *SigningMethodECDSA = jwt.SigningMethodES512
	// ed25519
	SigningMethodEdDSA *SigningMethodEd25519 = jwt.SigningMethodEdDSA
)

// HTTPBearerJWT verification interface
type HTTPBearerJWT interface {
	// EncryptKey the encrypted key
	EncryptKey() (any, error)

	// DecryptKey the Decryption key
	DecryptKey() (any, error)

	// SigningMethod encryption and decryption methods
	SigningMethod() SigningMethod

	// HTTPBearerJWT jwt logical
	HTTPBearerJWT(jwt *JWT)
}

type JWT struct {
	// the `iss` (Issuer) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.1
	Issuer string `json:"iss"`

	// the `sub` (Subject) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.2
	Subject string `json:"sub"`

	// the `aud` (Audience) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.3
	Audience []string `json:"aud"`

	// the `exp` (Expiration Time) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.4
	ExpiresAt time.Time `json:"exp"`

	// the `nbf` (Not Before) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.5
	NotBefore time.Time `json:"nbf"`

	// the `iat` (Issued At) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.6
	IssuedAt time.Time `json:"iat"`

	// the `jti` (JWT ID) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.7
	ID string `json:"jti"`

	Extensions map[string]any `json:"ext"`
}

func (j *JWT) GetExpirationTime() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: j.ExpiresAt}, nil
}

func (j *JWT) GetIssuedAt() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: j.IssuedAt}, nil
}

func (j *JWT) GetNotBefore() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: j.NotBefore}, nil
}

func (j *JWT) GetIssuer() (string, error) {
	return j.Issuer, nil
}

func (j *JWT) GetSubject() (string, error) {
	return j.Subject, nil
}

func (j *JWT) GetAudience() (jwt.ClaimStrings, error) {
	return j.Audience, nil
}

// Encrypt Get Jwt encrypted string
func (j *JWT) Encrypt(bearerJWT HTTPBearerJWT) (string, error) {
	token := jwt.NewWithClaims(bearerJWT.SigningMethod(), j)
	encryptKey, err := bearerJWT.EncryptKey()
	if err != nil {
		return "", err
	}
	return token.SignedString(encryptKey)
}

func (j *JWT) MarshalJSON() ([]byte, error) {
	m := map[string]any{}
	if j.Issuer != "" {
		m["iss"] = j.Issuer
	}
	if j.Subject != "" {
		m["sub"] = j.Subject
	}
	if len(j.Audience) > 0 {
		m["aud"] = j.Audience
	}
	if !j.ExpiresAt.IsZero() {
		m["exp"] = j.ExpiresAt.Unix()
	}
	if !j.NotBefore.IsZero() {
		m["nbf"] = j.NotBefore.Unix()
	}
	if !j.IssuedAt.IsZero() {
		m["iat"] = j.IssuedAt.Unix()
	}
	if j.ID != "" {
		m["jti"] = j.ID
	}
	if len(j.Extensions) > 0 {
		m["ext"] = j.Extensions
	}
	return json.Marshal(m)
}

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
	// Issuer
	Iss string `json:"iss"`

	// Subject, usually a user ID
	Sub string `json:"sub"`

	// The audience is usually the server API
	Aud []string `json:"aud"`

	// Expiration time
	Exp time.Time `json:"exp"`

	// Effective time
	Nbf time.Time `json:"nbf"`

	// Issuance Times
	Iat time.Time `json:"iat"`

	Extensions map[string]any
}

func (j *JWT) GetExpirationTime() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: j.Exp}, nil
}

func (j *JWT) GetIssuedAt() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: j.Iat}, nil
}

func (j *JWT) GetNotBefore() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: j.Nbf}, nil
}

func (j *JWT) GetIssuer() (string, error) {
	return j.Iss, nil
}

func (j *JWT) GetSubject() (string, error) {
	return j.Sub, nil
}

func (j *JWT) GetAudience() (jwt.ClaimStrings, error) {
	return j.Aud, nil
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
	if j.Iss != "" {
		m["iss"] = j.Iss
	}
	if j.Sub != "" {
		m["sub"] = j.Sub
	}
	if len(j.Aud) > 0 {
		m["aud"] = j.Aud
	}
	if !j.Exp.IsZero() {
		m["exp"] = j.Exp.Unix()
	}
	if !j.Nbf.IsZero() {
		m["nbf"] = j.Nbf.Unix()
	}
	if !j.Iat.IsZero() {
		m["iat"] = j.Iat.Unix()
	}
	for k, v := range j.Extensions {
		if k != "iss" && k != "sub" && k != "aud" && k != "exp" && k != "nbf" && k != "iat" {
			m[k] = v
		}
	}
	return json.Marshal(m)
}

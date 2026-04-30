package goapi

// HTTPBearer verification interface
type HTTPBearer interface {
	HTTPBearer(token string) error
}

// HTTPBasic verification interface
type HTTPBasic interface {
	HTTPBasic(username, password string) error
}

// ApiKey verification interface
type ApiKey interface {
	ApiKey() error
}

// SecurityOmitempty Determine whether HTTPBearer, HTTPBasic, and HTTPBearerJWT can be empty
type SecurityOmitempty interface {
	Omitempty() bool
}

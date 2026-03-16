package goapi

// HTTPBearer verification interface
type HTTPBearer interface {
	HTTPBearer(token string)
}

// HTTPBasic verification interface
type HTTPBasic interface {
	HTTPBasic(username, password string)
}

// ApiKey verification interface
type ApiKey interface {
	ApiKey()
}

// SecurityOmitempty Determine whether HTTPBearer, HTTPBasic, and HTTPBearerJWT can be empty
type SecurityOmitempty interface {
	Omitempty() bool
}

package goapi

// HTTPBearer verification interface
type HTTPBearer interface {
	HTTPBearer(token string)
	Omitempty() bool
}

// HTTPBasic verification interface
type HTTPBasic interface {
	HTTPBasic(username, password string)
	Omitempty() bool
}

// ApiKey verification interface
type ApiKey interface {
	ApiKey()
}

package goapi

type SecurityDescription interface {
	Desc() string
}

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

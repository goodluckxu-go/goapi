package goapi

type HTTPBearer interface {
	HTTPBearer(token string)
}

type HTTPBasic interface {
	HTTPBasic(username, password string)
}

type ApiKey interface {
	ApiKey()
}

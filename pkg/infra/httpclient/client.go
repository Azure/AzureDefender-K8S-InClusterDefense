package httpclient

import (
	"net/http"
)

// IHttpClient interface for http native client for dependency injection purposes
type IHttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

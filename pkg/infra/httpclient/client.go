package httpclient

import (
	"net/http"
)

// IHttpClient interface for http native client for dependency injection purposes
type IHttpClient interface {
	// Do is interface that for http client do  - initate an http call and return the response.
	Do(req *http.Request) (*http.Response, error)
}

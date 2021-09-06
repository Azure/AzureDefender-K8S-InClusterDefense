package acrauth

import (
	"encoding/json"
	"fmt"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/httpclient"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	scheme = "https"
)

var (
	nilArgError = errors.New("NilArgError")
)

// IACRTokenExchanger responsible to exchange ARM token toACR refresh token
type IACRTokenExchanger interface {
	// ExchangeACRAccessToken receives registry endpoint and an armToken (token to azure mgmt.) and
	// exchanges it to an ACR refresh token and returns it
	ExchangeACRAccessToken(registry string, armToken string) (string, error)
}

// ACRTokenExchanger basic implementation for IACRTokenExchanger interface
type ACRTokenExchanger struct {
	tracerProvider trace.ITracerProvider
	httpClient     httpclient.IHttpClient
}

// NewACRTokenExchanger Ctor
func NewACRTokenExchanger(instrumentationProvider instrumentation.IInstrumentationProvider, httpClient httpclient.IHttpClient) *ACRTokenExchanger {
	return &ACRTokenExchanger{
		tracerProvider: instrumentationProvider.GetTracerProvider("ACRTokenProvider"),
		httpClient:     httpClient,
	}
}

// ExchangeACRAccessToken receives registry endpoint and an armToken (token to azure mgmt.) and
// exchanges it to an ACR refresh token and returns it
// Generates an HTTP call to registry/oauth2/exchange rest api to exchange the token
func (tokenExchanger *ACRTokenExchanger) ExchangeACRAccessToken(registry string, armToken string) (string, error) {
	tracer := tokenExchanger.tracerProvider.GetTracer("ExchangeACRAccessToken")
	tracer.Info("Received:", "registry", registry)

	// Argument validation
	if registry == "" || armToken == "" {
		err := errors.Wrap(nilArgError, "ACRTokenExchanger")
		tracer.Error(err,"")
		return "", err
	}

	// Build HTTP request
	exchangeURL := fmt.Sprintf("%s://%s/oauth2/exchange", scheme, registry)
	ul, err := url.Parse(exchangeURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse token exchange url: %w", err)
	}
	parameters := url.Values{}
	parameters.Add("grant_type", "access_token")
	parameters.Add("service", ul.Hostname())
	parameters.Add("access_token", armToken)
	// Seems like tenantId is not required - if ever needed it should be added via:	//parameters.Add("tenant", tenantID) - maybe it is needed on cross tenant
	// Not adding it for now...

	req, err := http.NewRequest("POST", exchangeURL, strings.NewReader(parameters.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to construct token exchange reqeust: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(parameters.Encode())))

	// Creates a defer to close request on panic
	var resp *http.Response
	defer closeResponse(resp)

	// Invokes call to registry
	resp, err = tokenExchanger.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send token exchange request: %w", err)
	}

	// If error
	if resp.StatusCode != 200 {
		responseBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("ACR token exchange endpoint returned error status: %d. body: %s", resp.StatusCode, string(responseBytes))
	}

	// Extract response
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read request body: %w", err)
	}

	// Get token from response
	var tokenResp tokenResponse
	err = json.Unmarshal(responseBytes, &tokenResp)
	if err != nil {
		return "", fmt.Errorf("failed to read token exchange response: %w. response: %s", err, string(responseBytes))
	}

	// Return the refresh token
	return tokenResp.RefreshToken, nil
}

// closeResponse defer function to close request upon http panic
func closeResponse(resp *http.Response) {
	if resp == nil {
		return
	}
	resp.Body.Close()
}

// tokenResponse represents the response object from exchange token rest api of the registry
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Resource     string `json:"resource"`
	TokenType    string `json:"token_type"`
}
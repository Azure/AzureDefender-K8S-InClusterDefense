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

type IACRTokenExchanger interface {
	ExchangeACRAccessToken(registry string, armToken string) (string, error)
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Resource     string `json:"resource"`
	TokenType    string `json:"token_type"`
}

type ACRTokenExchanger struct {
	tracerProvider trace.ITracerProvider
	httpClient     httpclient.IHttpClient
}

func NewACRTokenExchanger(instrumentationProvider instrumentation.IInstrumentationProvider, httpClient httpclient.IHttpClient) *ACRTokenExchanger {
	return &ACRTokenExchanger{
		tracerProvider: instrumentationProvider.GetTracerProvider("ACRTokenProvider"),
		httpClient:     httpClient,
	}
}

// ExchangeACRAccessToken exchanges an ARM access token to an ACR access token
func (tokenExchanger *ACRTokenExchanger) ExchangeACRAccessToken(registry string, armToken string) (string, error) {
	tracer := tokenExchanger.tracerProvider.GetTracer("ExchangeACRAccessToken")
	tracer.Info("Received:", "registry", registry)

	if registry == "" || armToken == "" {
		err := errors.Wrap(nilArgError, "ACRTokenExchanger")
		tracer.Error(err,"")
		return "", err
	}

	exchangeURL := fmt.Sprintf("%s://%s/oauth2/exchange", scheme, registry)
	ul, err := url.Parse(exchangeURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse token exchange url: %w", err)
	}
	parameters := url.Values{}
	parameters.Add("grant_type", "access_token")
	parameters.Add("service", ul.Hostname())
	// Seems like tenantId is not required - if ever needed it should be added via:	//parameters.Add("tenant", tenantID) - maybe it is needed on cross tenant
	// Not adding it for now...
	parameters.Add("access_token", armToken)

	req, err := http.NewRequest("POST", exchangeURL, strings.NewReader(parameters.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to construct token exchange reqeust: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(parameters.Encode())))

	var resp *http.Response
	defer closeResponse(resp)

	resp, err = tokenExchanger.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send token exchange request: %w", err)
	}

	if resp.StatusCode != 200 {
		responseBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("ACR token exchange endpoint returned error status: %d. body: %s", resp.StatusCode, string(responseBytes))
	}

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read request body: %w", err)
	}

	var tokenResp tokenResponse
	err = json.Unmarshal(responseBytes, &tokenResp)
	if err != nil {
		return "", fmt.Errorf("failed to read token exchange response: %w. response: %s", err, string(responseBytes))
	}

	return tokenResp.RefreshToken, nil
}

func closeResponse(resp *http.Response) {
	if resp == nil {
		return
	}
	resp.Body.Close()
}

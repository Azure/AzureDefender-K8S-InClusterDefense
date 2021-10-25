package acrauth

import (
	"encoding/json"
	"fmt"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/httpclient"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	registryerrors "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/registry/errors"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/retrypolicy"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/utils"
	"github.com/pkg/errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	// _scheme is the http calls schema to use
	_scheme = "https"
	// _accessTokenGrantType is the grant type used for exchange call under parameter name _granTypeParameterName in api call
	_accessTokenGrantType = "access_token"
	// _granTypeParameterName is the parameter name passed in exchange call to specify the access grant type to the api - in our case it's an access token kind (that this is an arm token)
	_granTypeParameterName = "grant_type"
	// _serviceParameterName is the parameter name passed in exchange call to service the access to - in our case the registry host name
	_serviceParameterName = "service"
	// _accessTokenParameter is the parameter name passed in exchange call for specifying the access token the arm token in our case (value is the actual arm token)
	_accessTokenParameter = "access_token"
	// _postHTTPRequestType is http post type
	_postHTTPRequestType = "POST"
	// _contentTypeHeaderName is the header name for content type
	_contentTypeHeaderName = "Content-Type"
	// _applicationUrlEncodedContentType is the value for the _contentTypeHeaderName to be an application url encoded
	_applicationUrlEncodedContentType = "application/x-www-form-urlencoded"
	// _contentTypeHeaderName is the header name for content length
	_contentLengthHeaderName = "Content-Length"
)

var (
	_refreshTokenEmptyError = errors.New("RefreshToken is empty")
)

// IACRTokenExchanger responsible to exchange ARM token to ACR refresh token
type IACRTokenExchanger interface {
	// ExchangeACRAccessToken receives registry endpoint and an armToken (token to azure mgmt.) and
	// exchanges it to an ACR refresh token and returns it
	ExchangeACRAccessToken(registry string, armToken string) (string, error)
}

// ACRTokenExchanger implements IACRTokenExchanger interface
var _ IACRTokenExchanger = (*ACRTokenExchanger)(nil)

// ACRTokenExchanger basic implementation for IACRTokenExchanger interface
type ACRTokenExchanger struct {
	// tracerProvider is class to provide tracers to functions
	tracerProvider trace.ITracerProvider
	// httpClient is the client to initiate http calls with
	httpClient httpclient.IHttpClient
	// retry policy
	retryPolicy retrypolicy.IRetryPolicy
}

// tokenResponse represents the response object from exchange token rest api of the registry
type tokenResponse struct {
	// AccessToken used to exchange
	AccessToken string `json:"access_token"`
	// RefreshToken is the refresh token exchanged to/generated
	RefreshToken string `json:"refresh_token"`
	// Resource is the resource received the token for
	Resource string `json:"resource"`
	// TokenType is type of token exchanged to - in this case refresh token
	TokenType string `json:"token_type"`
}

// NewACRTokenExchanger Ctor
func NewACRTokenExchanger(instrumentationProvider instrumentation.IInstrumentationProvider, httpClient httpclient.IHttpClient, retryPolicy retrypolicy.IRetryPolicy) *ACRTokenExchanger {
	return &ACRTokenExchanger{
		tracerProvider: instrumentationProvider.GetTracerProvider("ACRTokenProvider"),
		httpClient:     httpClient,
		retryPolicy:    retryPolicy,
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
		err := errors.Wrap(utils.NilArgumentError, "ACRTokenExchanger")
		tracer.Error(err, "")
		return "", err
	}

	// Build HTTP request
	exchangeURL := fmt.Sprintf("%s://%s/oauth2/exchange", _scheme, registry)
	exchangeUrl, err := url.Parse(exchangeURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse token exchange url: %w", err)
	}
	parameters := url.Values{}
	parameters.Add(_granTypeParameterName, _accessTokenGrantType)
	parameters.Add(_serviceParameterName, exchangeUrl.Hostname())
	parameters.Add(_accessTokenParameter, armToken)
	// Seems like tenantId is not required - if ever needed it should be added via:	//parameters.Add("tenant", tenantID) - maybe it is needed on cross tenant
	// Not adding it for now...

	req, err := http.NewRequest(_postHTTPRequestType, exchangeURL, strings.NewReader(parameters.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to construct token exchange reqeust: %w", err)
	}

	req.Header.Add(_contentTypeHeaderName, _applicationUrlEncodedContentType)
	req.Header.Add(_contentLengthHeaderName, strconv.Itoa(len(parameters.Encode())))

	// Creates a defer to close request on panic
	var resp *http.Response
	defer closeResponse(resp)

	// Invokes call to registry
	// TODO add retry policy

	err = tokenExchanger.retryPolicy.RetryAction(
		func() error {
			resp, err = tokenExchanger.httpClient.Do(req)
			if err == nil {
				return nil
			}
			// Err != nil so set response to nil and return error
			resp = nil
			return err
		},
		// Retry on all errors except not NoSuchError
		func(err error) bool { return !tokenExchanger.isNoSuchHostErr(err) },
	)

	if err != nil {
		// If registry is not found, convert to known err.
		if tokenExchanger.isNoSuchHostErr(err) {
			// If its this error - convert the error to known error and continue
			err = registryerrors.NewRegistryIsNotFoundErr(registry, err)
		}

		err = errors.Wrap(err, "failed to send token exchange request")
		tracer.Error(err, "")
		return "", err
	}

	// If error
	// TODO @tomerwinberger - do you think that we should add this also the retry policy?
	if resp.StatusCode != 200 {
		responseBytes, _ := ioutil.ReadAll(resp.Body)
		err = errors.Wrap(fmt.Errorf("ACR token exchange endpoint returned error status: %d. body: %s", resp.StatusCode, string(responseBytes)), "ACRTokenExchanger")
		tracer.Error(err, "")
		return "", err
	}

	// Extract response
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.Wrap(fmt.Errorf("failed to read request body: %w", err), "ACRTokenExchanger")
		tracer.Error(err, "")
		return "", err
	}

	// Get token from response
	var tokenResp tokenResponse
	err = json.Unmarshal(responseBytes, &tokenResp)
	if err != nil {
		err = errors.Wrap(fmt.Errorf("failed to read token exchange response: %w. response: %s", err, string(responseBytes)), "ACRTokenExchanger")
		tracer.Error(err, "")
		return "", err
	}

	if tokenResp.RefreshToken == "" {
		err = errors.Wrap(_refreshTokenEmptyError, "ACRTokenExchanger")
		tracer.Error(err, "")
		return "", err
	}

	// Return the refresh token
	return tokenResp.RefreshToken, nil
}

// closeResponse defer function to close request upon http panic
func closeResponse(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	resp.Body.Close()
}

// isNoSuchHostErr gets an error and returns true if the err is caused by DNSError - it means that the registry is not exist
// TODO currently we decided to start with this error as unscanned - we should see the metrics and decide if this error
// 	encountered sometimes when the registry is exists.
func (tokenExchanger *ACRTokenExchanger) isNoSuchHostErr(err error) bool {
	var dnsError *net.DNSError
	if errors.As(err, &dnsError) {
		return true
	}
	return false
}

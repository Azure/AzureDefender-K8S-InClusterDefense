package credscan

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// NewCredScanDataProvider Constructor
func NewCredScanDataProvider(instrumentationProvider instrumentation.IInstrumentationProvider) *CredScanDataProvider {
	return &CredScanDataProvider{
		tracerProvider:    instrumentationProvider.GetTracerProvider("NewCredScanDataProvider"),
	}
}



// convert request into json buffer
func (provider CredScanDataProvider) encodeResourceForCredScan(resourceMetadata admission.Request) ([]byte, error){
	tracer := provider.tracerProvider.GetTracer("encodeResourceForCredScan")
	body, err := json.Marshal(resourceMetadata)
	if err != nil {
		err = errors.Wrap(err, "CredScan.encodeResourceForCredScan failed on json.Marshal results")
		tracer.Error(err, "")
		return nil, err
	}
	body , err= json.Marshal(string(body))
	if err != nil {
		err = errors.Wrap(err, "CredScan.encodeResourceForCredScan failed on json.Marshal results")
		tracer.Error(err, "")
		return nil, err
	}
	return body, err
}

// postCredScanRequest http request to credscan server
// return json string represent cred scan results
func (provider CredScanDataProvider) postCredScanRequest(url string, jsonStr []byte) ([]byte, error) {
	tracer := provider.tracerProvider.GetTracer("postCredScanRequest")

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		err = errors.Wrap(err, "CredScan.postCredScanRequest failed on http.NewRequest")
		tracer.Error(err, "")
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		err = errors.Wrap(err, "CredScan.postCredScanRequest failed on posting http request")
		tracer.Error(err, "")
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return body, err
}

// convert the string response into array of scan results
func (provider CredScanDataProvider) parseCredScanResults(postRes []byte) ([]*CredScanInfo, error){
	tracer := provider.tracerProvider.GetTracer("parseCredScanResults")

	var scanResults []*CredScanInfo
	err := json.Unmarshal(postRes, &scanResults)
	if err != nil{
		err = errors.Wrap(err, "CredScan.parseCredScanResults failed on Unmarshal CredScanResults")
		tracer.Error(err, "")
		return nil, err
	}

	for _, match := range scanResults {
		if match.MatchingConfidence > _threshold {
			match.ScanStatus = _scanStatusUnhealthy
		}else{
			match.ScanStatus = _scanStatusHealthy
		}
	}
	return scanResults, err
}

// GetCredScanResults - get credential scan results of the resourceMetadata.
func (provider CredScanDataProvider) GetCredScanResults(resourceMetadata admission.Request) ([]*CredScanInfo, error){
	tracer := provider.tracerProvider.GetTracer("GetCredScanResults")
	body, err:= provider.encodeResourceForCredScan(resourceMetadata)
	if err != nil {
		err = errors.Wrap(err, "CredScan.GetCredScanResults failed on encodeResourceForCredScan")
		tracer.Error(err, "")
		return nil, err
	}

	postRes, err := provider.postCredScanRequest(_credScanServerUrl, body)
	if err != nil {
		err = errors.Wrap(err, "CredScan.GetCredScanResults failed on postCredScanRequest")
		tracer.Error(err, "")
		return nil, err
	}

	scanResults, err := provider.parseCredScanResults(postRes)
	if err != nil {
		err = errors.Wrap(err, "CredScan.GetCredScanResults failed on parseCredScanResults")
		tracer.Error(err, "")
		return nil, err
	}
	return scanResults, err
}
package runtime

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	"github.com/stretchr/testify/suite"
	"gomodules.xyz/jsonpatch/v2"
	"io/ioutil"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"strings"
	"testing"
	"time"
)

const (
	_expectedTestAddPatchOperation   = "add"
	_expectedTestAnnotationPatchPath = "/metadata/annotations"
	_url                             = "https://localhost:8000/mutate"
	_alpineDigest                    = "sha256:def822f9851ca422481ec6fee59a9966f12b351c62ccb9aca841526ffaa9f748"
	_imageScanUnhealthyDigest        = "sha256:f6a835256950f699175eecb9fd82e4a84684c9bab6ffb641b6fc23ff7b23e4b3"
	_runtimeTestEnabledEnvKey        = "RuntimeTestsEnabled"
)

var (
	_secondContainerVulnerabilityScanInfo = &contracts.ContainerVulnerabilityScanInfo{Name: "Or"}
)

// in order to run against a proxy in cluster: 'kubectl port-forward service/azure-defender-proxy-service 8000:443 -n kube-system'
type RuntimeTestSuite struct {
	suite.Suite
	req    *http.Request
	client *http.Client
}

func (suite *RuntimeTestSuite) SetupSuite() {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	suite.client = &http.Client{Transport: transCfg}
}

func (suite *RuntimeTestSuite) SetupTest() {
	var err error
	suite.req, err = http.NewRequest("Get", _url, nil)
	suite.req.Header.Set("Content-Type", "application/json")
	suite.NoError(err)
}

func (suite *RuntimeTestSuite) TestHappyPath() {
	containers := []corev1.Container{
		{
			Name:  "containerTest1",
			Image: "tomerwdevops.azurecr.io/alpine:latest",
		},
	}
	expected := []*contracts.ContainerVulnerabilityScanInfo{{Name: containers[0].Name,
		Image:      &contracts.Image{Name: containers[0].Image, Digest: _alpineDigest},
		ScanStatus: contracts.HealthyScan, ScanFindings: []*contracts.ScanFinding{}}}

	pod := createPodForTests(containers, nil)
	reqA := suite.createRequestForTests(pod)
	postBody, err := json.Marshal(reqA)
	suite.req.Body = ioutil.NopCloser(bytes.NewReader(postBody))
	resp, err := suite.client.Do(suite.req)
	admissionResponse := suite.assertCommonAndExtractAdmissionRes(resp, err)
	vulnInfo := suite.assertAndExtractVulnInfoList(admissionResponse)
	suite.Equal(expected, vulnInfo.Containers)
}

func (suite *RuntimeTestSuite) TestNonACR() {
	containers := []corev1.Container{
		{
			Name:  "containerTest1",
			Image: "tomerwdevops.nonacr.io/alpine:latest",
		},
	}

	expected := []*contracts.ContainerVulnerabilityScanInfo{{Name: containers[0].Name,
		Image:      &contracts.Image{Name: containers[0].Image, Digest: ""},
		ScanStatus: contracts.Unscanned, ScanFindings: nil, AdditionalData: map[string]string{"UnscannedReason": string(contracts.ImageIsNotInACRRegistryUnscannedReason)}}}

	pod := createPodForTests(containers, nil)
	reqA := suite.createRequestForTests(pod)
	postBody, err := json.Marshal(reqA)
	suite.req.Body = ioutil.NopCloser(bytes.NewReader(postBody))
	resp, err := suite.client.Do(suite.req)
	admissionResponse := suite.assertCommonAndExtractAdmissionRes(resp, err)
	suite.NotNil(admissionResponse)
	vulnInfo := suite.assertAndExtractVulnInfoList(admissionResponse)
	suite.Equal(expected, vulnInfo.Containers)
}

func (suite *RuntimeTestSuite) TestACRDoesntExists() {
	containers := []corev1.Container{
		{
			Name:  "containerTest1",
			Image: "doesntexistsaaa.azurecr.io/alpine:latest",
		},
	}

	expected := []*contracts.ContainerVulnerabilityScanInfo{{Name: containers[0].Name,
		Image:      &contracts.Image{Name: containers[0].Image, Digest: ""},
		ScanStatus: contracts.Unscanned, ScanFindings: nil, AdditionalData: map[string]string{"UnscannedReason": string(contracts.RegistryDoesNotExistUnscannedReason)}}}
	pod := createPodForTests(containers, nil)
	reqA := suite.createRequestForTests(pod)
	postBody, err := json.Marshal(reqA)
	suite.req.Body = ioutil.NopCloser(bytes.NewReader(postBody))
	resp, err := suite.client.Do(suite.req)
	admissionResponse := suite.assertCommonAndExtractAdmissionRes(resp, err)
	suite.NotNil(admissionResponse)
	vulnInfo := suite.assertAndExtractVulnInfoList(admissionResponse)
	suite.Equal(expected, vulnInfo.Containers)
}

func (suite *RuntimeTestSuite) TestImageDoesntExists() {
	containers := []corev1.Container{
		{
			Name:  "containerTest1",
			Image: "tomerwdevops.azurecr.io/doesntexists:latest",
		},
	}

	expected := []*contracts.ContainerVulnerabilityScanInfo{{Name: containers[0].Name,
		Image:      &contracts.Image{Name: containers[0].Image, Digest: ""},
		ScanStatus: contracts.Unscanned, ScanFindings: nil, AdditionalData: map[string]string{"UnscannedReason": string(contracts.ImageDoesNotExistUnscannedReason)}}}

	pod := createPodForTests(containers, nil)
	reqA := suite.createRequestForTests(pod)
	postBody, err := json.Marshal(reqA)
	suite.req.Body = ioutil.NopCloser(bytes.NewReader(postBody))
	resp, err := suite.client.Do(suite.req)
	admissionResponse := suite.assertCommonAndExtractAdmissionRes(resp, err)
	suite.NotNil(admissionResponse)
	vulnInfo := suite.assertAndExtractVulnInfoList(admissionResponse)
	suite.Equal(expected, vulnInfo.Containers)
}

func (suite *RuntimeTestSuite) TestImageBadFormat_Allowed() {
	containers := []corev1.Container{
		{
			Name:  "containerTest1",
			Image: "tomerwdevops.azurecr.io / doesntexists:latest",
		},
	}

	pod := createPodForTests(containers, nil)
	reqA := suite.createRequestForTests(pod)
	postBody, err := json.Marshal(reqA)
	suite.req.Body = ioutil.NopCloser(bytes.NewReader(postBody))
	resp, err := suite.client.Do(suite.req)
	admissionResponse := suite.assertCommonAndExtractAdmissionResError(resp, err)
	suite.NotNil(admissionResponse)
	// Important to not fail allowed on error
	suite.True(admissionResponse.Allowed)
	suite.NotNil(admissionResponse.Result)
	suite.Nil(admissionResponse.Patches)
}

func (suite *RuntimeTestSuite) TestUnhealthy() {
	containers := []corev1.Container{
		{
			Name:  "containerTest1",
			Image: "tomerwdevops.azurecr.io/imagescan:62",
		},
	}

	expected := []*contracts.ContainerVulnerabilityScanInfo{{Name: containers[0].Name,
		Image:      &contracts.Image{Name: containers[0].Image, Digest: _imageScanUnhealthyDigest},
		ScanStatus: contracts.UnhealthyScan, ScanFindings: []*contracts.ScanFinding{}}}

	pod := createPodForTests(containers, nil)
	reqA := suite.createRequestForTests(pod)
	postBody, err := json.Marshal(reqA)
	suite.req.Body = ioutil.NopCloser(bytes.NewReader(postBody))
	resp, err := suite.client.Do(suite.req)
	admissionResponse := suite.assertCommonAndExtractAdmissionRes(resp, err)
	suite.NotNil(admissionResponse)
	vulnInfo := suite.assertAndExtractVulnInfoList(admissionResponse)
	suite.Equal(1, len(vulnInfo.Containers))
	suite.Equal(expected[0].Image, vulnInfo.Containers[0].Image)
	suite.Equal(expected[0].ScanStatus, vulnInfo.Containers[0].ScanStatus)
	suite.True(100 < len(vulnInfo.Containers[0].ScanFindings))
}

func (suite *RuntimeTestSuite) TestMultipleImages() {
	containers := []corev1.Container{
		{
			Name:  "containerTest1",
			Image: "tomerwdevops.azurecr.io/alpine:latest",
		},
		{
			Name:  "containerTest2",
			Image: "tomerwdevops.azurecr.io/imagescan:62",
		},
	}

	expected := []*contracts.ContainerVulnerabilityScanInfo{
		{Name: containers[0].Name, Image: &contracts.Image{Name: containers[0].Image, Digest: _alpineDigest},
			ScanStatus: contracts.HealthyScan, ScanFindings: []*contracts.ScanFinding{}},
		{Name: containers[1].Name, Image: &contracts.Image{Name: containers[1].Image, Digest: _imageScanUnhealthyDigest},
			ScanStatus: contracts.UnhealthyScan, ScanFindings: nil}}

	pod := createPodForTests(containers, nil)
	reqA := suite.createRequestForTests(pod)
	postBody, err := json.Marshal(reqA)
	suite.req.Body = ioutil.NopCloser(bytes.NewReader(postBody))
	resp, err := suite.client.Do(suite.req)
	admissionResponse := suite.assertCommonAndExtractAdmissionRes(resp, err)
	suite.NotNil(admissionResponse)
	vulnInfo := suite.assertAndExtractVulnInfoList(admissionResponse)
	suite.Equal(2, len(vulnInfo.Containers))
	firstIndex := 0
	secIndex := 1
	if expected[0].Name == vulnInfo.Containers[1].Name {
		firstIndex = 1
		secIndex = 0
	}

	suite.Equal(expected[firstIndex], vulnInfo.Containers[firstIndex])
	suite.Equal(expected[secIndex].Name, vulnInfo.Containers[secIndex].Name)
	suite.Equal(expected[secIndex].Image, vulnInfo.Containers[secIndex].Image)
	suite.Equal(expected[secIndex].ScanStatus, vulnInfo.Containers[secIndex].ScanStatus)
	suite.True(100 < len(vulnInfo.Containers[secIndex].ScanFindings))
}

func (suite *RuntimeTestSuite) assertCommonAndExtractAdmissionRes(resp *http.Response, err error) *admission.Response {
	suite.NoError(err)
	suite.NotNil(resp)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.NotNil(resp.Body)
	payload, err := ioutil.ReadAll(resp.Body)
	suite.NoError(err)
	var admissionReview = &v1.AdmissionReview{}
	err = json.Unmarshal(payload, admissionReview)
	suite.NotNil(admissionReview.Response)
	var patches []jsonpatch.JsonPatchOperation
	err = json.Unmarshal(admissionReview.Response.Patch, &patches)
	suite.NoError(err)
	suite.True(admissionReview.Response.Allowed)
	suite.Equal(int32(200), admissionReview.Response.Result.Code)
	return &admission.Response{AdmissionResponse: *admissionReview.Response, Patches: patches}
}

func (suite *RuntimeTestSuite) assertCommonAndExtractAdmissionResError(resp *http.Response, err error) *admission.Response {
	suite.NoError(err)
	suite.NotNil(resp)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.NotNil(resp.Body)
	payload, err := ioutil.ReadAll(resp.Body)
	suite.NoError(err)
	var admissionReview = &v1.AdmissionReview{}
	err = json.Unmarshal(payload, admissionReview)
	suite.NotNil(admissionReview.Response)
	suite.Nil(admissionReview.Response.Patch)
	suite.NoError(err)
	suite.True(admissionReview.Response.Allowed)
	suite.Equal(int32(500), admissionReview.Response.Result.Code)
	return &admission.Response{AdmissionResponse: *admissionReview.Response}
}

func (suite *RuntimeTestSuite) assertAndExtractVulnInfoList(res *admission.Response) *contracts.ContainerVulnerabilityScanInfoList {
	suite.NotNil(res)
	var list = &contracts.ContainerVulnerabilityScanInfoList{}
	suite.Equal(1, len(res.Patches))
	suite.Equal("add", res.Patches[0].Operation)
	suite.Equal("/metadata/annotations", res.Patches[0].Path)
	annotations := res.Patches[0].Value.(map[string]interface{})
	suite.Equal(1, len(annotations))
	err := json.Unmarshal([]byte(annotations["azuredefender.io/containers.vulnerability.scan.info"].(string)), list)
	suite.NoError(err)
	suite.True((time.Now().UTC().Sub(list.GeneratedTimestamp)) <= time.Second*2)
	suite.NotNil(list.Containers)
	return list
}

func (suite *RuntimeTestSuite) createRequestForTests(pod *corev1.Pod) *v1.AdmissionReview {
	raw, err := json.Marshal(pod)
	suite.NoError(err)

	return &v1.AdmissionReview{
		Request: &v1.AdmissionRequest{
			Name:      "podTest",
			Namespace: "default",
			Operation: v1.Create,
			Kind: metav1.GroupVersionKind{
				Kind:    "Pod",
				Group:   "",
				Version: "v1",
			},
			Object: runtime.RawExtension{
				Raw: raw,
			},
		},
	}
}

func createPodForTests(containers []corev1.Container, initContainers []corev1.Container) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "podTest",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{},
		Spec: corev1.PodSpec{
			Containers:     containers,
			InitContainers: initContainers,
		},
	}
}

func (suite *RuntimeTestSuite) checkPatch(expected []*contracts.ContainerVulnerabilityScanInfo, patch jsonpatch.JsonPatchOperation) {
	// Verify the operation and the patch
	suite.Equal(_expectedTestAddPatchOperation, patch.Operation)
	suite.Equal(_expectedTestAnnotationPatchPath, patch.Path)

	// Get data string
	mapAnnotations, ok := patch.Value.(map[string]string)
	suite.True(ok)

	suite.Equal(1, len(mapAnnotations))
	strValue, ok := mapAnnotations[contracts.ContainersVulnerabilityScanInfoAnnotationName]
	suite.True(ok)

	// Unmarshal
	scanInfoList := new(contracts.ContainerVulnerabilityScanInfoList)
	err := json.Unmarshal([]byte(strValue), scanInfoList)
	suite.Nil(err)

	// Verify timestamp
	diff := time.Now().UTC().Sub(scanInfoList.GeneratedTimestamp)
	suite.True((diff >= 0 && diff < time.Second))
	suite.Equal(time.UTC, scanInfoList.GeneratedTimestamp.Location())

	// Verify containers slice deep equal verification
	suite.Equal(expected, scanInfoList.Containers)
}

func TestRuntime(t *testing.T) {
	e, ok := os.LookupEnv(_runtimeTestEnabledEnvKey)
	if !ok || strings.ToLower(e) != "true" {
		t.SkipNow() // set test env var _runtimeTestEnabledEnvKey to true
	}

	suite.Run(t, new(RuntimeTestSuite))
}

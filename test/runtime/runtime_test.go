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
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"testing"
	"time"
)

const (
	_expectedTestAddPatchOperation   = "add"
	_expectedTestAnnotationPatchPath = "/metadata/annotations"
)

var (
	_containers = []corev1.Container{
		{
			Name:  "containerTest1",
			Image: "image1.com",
		},
		{
			Name:  "containerTest2",
			Image: "image2.com",
		},
		{
			Name:  "containerTest3",
			Image: "image3.com",
		},
	}

	_firstContainerVulnerabilityScanInfo  = &contracts.ContainerVulnerabilityScanInfo{Name: "Lior"}
	_secondContainerVulnerabilityScanInfo = &contracts.ContainerVulnerabilityScanInfo{Name: "Or"}
)

type TomerTestSuite struct {
	suite.Suite
	req *http.Request
	client *http.Client
}

func (suite *TomerTestSuite) SetupSuite() {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	suite.client = &http.Client{Transport: transCfg}
}

func (suite *TomerTestSuite) SetupTest() {
	var err error
	suite.req, err = http.NewRequest("Get", "https://localhost:8000/mutate", nil)
	suite.req.Header.Set("Content-Type",  "application/json")
	suite.NoError(err)
}


func (suite *TomerTestSuite) TestA() {
	pod := createPodForTests(_containers, nil)
	reqA := suite.createRequestForTests(pod)
	postBody, err := json.Marshal(reqA)
	suite.req.Body = ioutil.NopCloser(bytes.NewReader(postBody))
	resp, err := suite.client.Do(suite.req)
	admissionResponse := suite.assertCommonAndExtractAdmissionReq(resp, err)
	suite.NotNil(admissionResponse)
}

func (suite *TomerTestSuite) assertCommonAndExtractAdmissionReq(resp *http.Response, err error) *admission.Response {
	suite.NoError(err)
	suite.NotNil(resp)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.NotNil(resp.Body)
	payload, err := ioutil.ReadAll(resp.Body)
	suite.NoError(err)
	var admissionReview =  &v1.AdmissionReview{}
	err = json.Unmarshal(payload, admissionReview)
	suite.NotNil(admissionReview.Response)
	var patches []jsonpatch.JsonPatchOperation
	err = json.Unmarshal(admissionReview.Response.Patch, &patches)
	suite.NoError(err)
	return &admission.Response{AdmissionResponse: *admissionReview.Response, Patches: patches}
}

func (suite *TomerTestSuite) createRequestForTests(pod *corev1.Pod) *v1.AdmissionReview {
	raw, err := json.Marshal(pod)
	suite.NoError(err)

	return	&v1.AdmissionReview{
		Request: 	&v1.AdmissionRequest{
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

func (suite *TomerTestSuite) checkPatch(expected []*contracts.ContainerVulnerabilityScanInfo, patch jsonpatch.JsonPatchOperation) {
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

func TestTomer(t *testing.T){
	suite.Run(t, new(TomerTestSuite))
}



package admisionrequest

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"testing"
)

type TestSuite struct {
	suite.Suite
	req *admission.Request
	pod *corev1.Pod
}

func (suite *TestSuite) SetupTest() {

	suite.pod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "podTest",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "containerTest",
					Image: "image.com",
				},
			},
		},
	}

	raw, err := json.Marshal(suite.pod)
	if err != nil {
		log.Fatal(err)
	}

	suite.req = &admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Name: "podTest",
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

func (suite *TestSuite) Test_UnmarshalPod_PodAdmissionReqWithMatchingObject_AsExpected() {

	pod, err := UnmarshalPod(suite.req)
	suite.Nil(err)
	reflect.DeepEqual(suite.pod, pod)
}

func (suite *TestSuite) Test_UnmarshalPod_BadFormat_Error() {
	suite.req.Object.Raw = []byte("{ \"a\" : \"badFormat\"")

	pod, err := UnmarshalPod(suite.req)
	suite.Nil(pod)
	suite.NotEqual(nil, err)
}

func (suite *TestSuite) Test_UnmarshalPod_RequestNull_Error() {
	pod, err := UnmarshalPod(nil)
	suite.Nil(pod)
	suite.Equal(_errInvalidAdmission, err)
}

func (suite *TestSuite) Test_UnmarshalPod_EmptyRawObject_Error() {
	suite.req.Object.Raw = []byte{}

	pod, err := UnmarshalPod(suite.req)
	suite.Nil(pod)
	suite.Equal(_errObjectNotFound, err)
}

func (suite *TestSuite) Test_UnmarshalPod_FNotPodKindRequest_Error() {
	suite.req.Kind.Kind = "NotPod"

	pod, err := UnmarshalPod(suite.req)
	suite.Nil(pod)
	suite.Equal(_errUnexpectedResource, err)
}

func (suite *TestSuite) Test_UnmarshalPod_PodKindConstVerification() {
	suite.Equal("Pod", PodKind)
}

func TestUnmarshalPod(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

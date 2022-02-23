package admisionrequest

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	admissionv1 "k8s.io/api/admission/v1"
	apps1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"testing"
)


//mata :=metav1.ObjectMeta{
//	Name: "podTest",
//}
//
//podSpec = Spec: corev1.PodSpec{
//	Containers: []corev1.Container{
//		{
//			Name:  "containerTest",
//			Image: "image.com",
//		},
//	},
//}

type TestSuite struct {
	suite.Suite
	req *admission.Request
	pod *corev1.Pod
	deployment *apps1.Deployment
}



func (suite *TestSuite) SetupTest() {
	meta:=metav1.ObjectMeta{
		Name: "podTest",
	}
	spec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "containerTest",
				Image: "image.com",
			},
		},
	}
	suite.pod = &corev1.Pod{
		ObjectMeta: meta,
		Spec: spec,
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

	suite.deployment = &apps1.Deployment{
		ObjectMeta: meta,
		Spec: apps1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec:spec,
			} ,
		},
	}
}

func (suite *TestSuite) Test_GetWorkloadResourceFromAdmissionRequest_PodAdmissionReqWithMatchingObject_AsExpected() {

	workLoadResource, err := GetWorkloadResourceFromAdmissionRequest(suite.req)
	podResource :=WorkloadResource{}
	podResource.PodSpec = suite.pod.Spec
	podResource.ResourceMetadata = suite.pod.ObjectMeta
	suite.Nil(err)
	reflect.DeepEqual(podResource, workLoadResource)
}

func (suite *TestSuite) Test_GetWorkloadResourceFromAdmissionRequest_BadFormat_Error() {
	suite.req.Object.Raw = []byte("{ \"a\" : \"badFormat\"")

	workLoadResource, err := GetWorkloadResourceFromAdmissionRequest(suite.req)
	suite.Nil(workLoadResource)
	suite.NotEqual(nil, err)
}

func (suite *TestSuite) Test_GetWorkloadResourceFromAdmissionRequest_RequestNull_Error() {
	workLoadResource, err := GetWorkloadResourceFromAdmissionRequest(nil)
	suite.Nil(workLoadResource)
	suite.Equal(_errInvalidAdmission, err)
}

func (suite *TestSuite) Test_GetWorkloadResourceFromAdmissionRequest_EmptyRawObject_Error() {
	suite.req.Object.Raw = []byte{}
	workLoadResource, err := GetWorkloadResourceFromAdmissionRequest(suite.req)
	suite.Nil(workLoadResource)
	suite.Equal(_errObjectNotFound, err)
}

func (suite *TestSuite) Test_GetWorkloadResourceFromAdmissionRequest_NotWorkloadResourceKindRequest_Error() {
	suite.req.Kind.Kind = "NotWorkloadResource"
	workLoadResource, err := GetWorkloadResourceFromAdmissionRequest(suite.req)
	suite.Nil(workLoadResource)
	suite.Equal(_errUnexpectedResource, err)
}

func (suite *TestSuite) Test_UnmarshalPod_PodKindConstVerification() {
	suite.Equal("Pod", PodKind)
}

func TestUnmarshalPod(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

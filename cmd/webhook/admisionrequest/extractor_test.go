package admisionrequest


import (
	"encoding/json"
	"github.com/pkg/errors"
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


func PodToWorkloadResourceForTest(pod *corev1.Pod)(workloadResource * WorkloadResource){
	listContainers :=make([]Container,len(pod.Spec.Containers))
	for i:=0;i<len(pod.Spec.Containers);i++{
		listContainers[i] = Container{Name: pod.Spec.Containers[i].Name,Image: pod.Spec.Containers[i].Image}
	}
	listInitContainers :=make([]Container,len(pod.Spec.InitContainers))
	for i:=0;i<len(pod.Spec.Containers);i++{
		listInitContainers[i] = Container{Name: pod.Spec.InitContainers[i].Name,Image: pod.Spec.InitContainers[i].Image}
	}
	spec := PodSpec{Containers: listContainers, InitContainers: listInitContainers, ImagePullSecrets: pod.Spec.ImagePullSecrets, ServiceAccountName: pod.Spec.ServiceAccountName}
	metadata := ObjectMetadata{Namespace: pod.Namespace, Annotations: pod.Annotations,OwnerReferences: pod.OwnerReferences}
	return &WorkloadResource{Metadata: metadata,Spec: spec}
}

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
	suite.Equal(errors.New(_errMsgInvalidAdmission), err)
}

func (suite *TestSuite) Test_GetWorkloadResourceFromAdmissionRequest_EmptyRawObject_Error() {
	suite.req.Object.Raw = []byte{}
	workLoadResource, err := GetWorkloadResourceFromAdmissionRequest(suite.req)
	suite.Nil(workLoadResource)
	suite.Equal(errors.New(_errMsgObjectNotFound), err)
}

func (suite *TestSuite) Test_GetWorkloadResourceFromAdmissionRequest_NotWorkloadResourceKindRequest_Error() {
	suite.req.Kind.Kind = "NotWorkloadResource"
	workLoadResource, err := GetWorkloadResourceFromAdmissionRequest(suite.req)
	suite.Nil(workLoadResource)
	suite.Equal(errors.New(_errMsgUnexpectedResource), err)
}


func TestUnmarshalPod(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

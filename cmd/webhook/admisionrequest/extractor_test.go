package admisionrequest

import (
	"encoding/json"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	admissionv1 "k8s.io/api/admission/v1"
	apps1 "k8s.io/api/apps/v1"
	batch1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"strings"
	"testing"
)

var (
	_actualContainers = []corev1.Container{
		{
			Name:  "containerTest",
			Image: "image.com",
		},
	}
	_expectedContainers = []*Container{
		{
			Name:  "containerTest",
			Image: "image.com",
		},
	}
	_actualInitContainers = []corev1.Container{
		{
			Name:  "initContainerTest",
			Image: "image.com",
		},
	}
	_expectedInitContainers = []*Container{
		{
			Name:  "initContainerTest",
			Image: "image.com",
		},
	}
	_name                     = "podTest"
	_expectedImagePullSecrets = []*corev1.LocalObjectReference{
		{
			Name: "secret",
		},
	}
	_actualImagePullSecrets = []corev1.LocalObjectReference{
		{
			Name: "secret",
		},
	}
	_serviceAccountName = "podServiceAccount"
	_namespace          = "podNameSpace"
	_annotation         = map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	_expectedOwnerReferences = []*OwnerReference{
		{
			APIVersion: "v1",
			Kind:       "pod",
			Name:       "podName",
		},
	}
	_actualOwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: "v1",
			Kind:       "pod",
			Name:       "podName",
		},
	}
)

type TestSuite struct {
	suite.Suite
	pod                       *corev1.Pod
	workloadResource          *WorkloadResource
	emptyWorkloadResource     *WorkloadResource
	podReq                    *admission.Request
	emptyPodReq               *admission.Request
	daemonSetReq              *admission.Request
	podWithEmptyPropertiesReq *admission.Request
	extractor                 Extractor
	config ExtractorConfiguration
}

func (suite *TestSuite) SetupTest() {
	suite.pod = createFullPodForTests()
	suite.workloadResource = createFullWorkloadResourceForTests()
	suite.emptyWorkloadResource = createEmptyWorkloadResourceForTests()
	suite.podReq = createReq(suite.pod, "Pod")
	suite.config = ExtractorConfiguration{SupportedKubernetesWorkloadResources: []string{"Pod", "Deployment",
		"ReplicaSet", "StatefulSet", "DaemonSet", "Job", "CronJob", "ReplicationController"}}
	suite.extractor = *NewExtractor(instrumentation.NewNoOpInstrumentationProvider(),&suite.config)
}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_PodAdmissionReqWithMatchingObject_AsExpected() {
	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(suite.podReq)
	suite.Nil(err)
	suite.True(reflect.DeepEqual(suite.workloadResource, workLoadResource))
}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_DeploymentAdmissionReqWithMatchingObject_AsExpected() {
	deployment := createFullDeploymentForTests()
	req := createReq(deployment, "Deployment")
	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(req)
	suite.Nil(err)
	suite.True(reflect.DeepEqual(suite.workloadResource, workLoadResource))
}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_ReplicaSetAdmissionReqWithMatchingObject_AsExpected() {
	replicaSet := createFullReplicaSetForTests()
	req := createReq(replicaSet, "ReplicaSet")
	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(req)
	suite.Nil(err)
	suite.True(reflect.DeepEqual(suite.workloadResource, workLoadResource))
}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_ReplicationControllerAdmissionReqWithMatchingObject_AsExpected() {
	replicationController := createFullReplicationControllerForTests()
	req := createReq(replicationController, "ReplicationController")
	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(req)
	suite.Nil(err)
	suite.True(reflect.DeepEqual(suite.workloadResource, workLoadResource))
}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_StatefulSetAdmissionReqWithMatchingObject_AsExpected() {
	statefulSet := createFullStatefulSetForTests()
	req := createReq(statefulSet, "StatefulSet")
	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(req)
	suite.Nil(err)
	suite.True(reflect.DeepEqual(suite.workloadResource, workLoadResource))
}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_DaemonSetAdmissionReqWithMatchingObject_AsExpected() {
	daemonSet := createFullDaemonSetForTests()
	req := createReq(daemonSet, "DaemonSet")
	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(req)
	suite.Nil(err)
	suite.True(reflect.DeepEqual(suite.workloadResource, workLoadResource))
}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_JobAdmissionReqWithMatchingObject_AsExpected() {
	job := createFullJobForTests()
	req := createReq(job, "Job")
	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(req)
	suite.Nil(err)
	suite.True(reflect.DeepEqual(suite.workloadResource, workLoadResource))
}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_CronJobAdmissionReqWithMatchingObject_AsExpected() {
	cronJob := createFullCronJobForTests()
	req := createReq(cronJob, "CronJob")
	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(req)
	suite.Nil(err)
	suite.True(reflect.DeepEqual(suite.workloadResource, workLoadResource))
}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_EmptyPodAdmissionReqWithMatchingObject_AsExpected() {
	emptyPod := createEmptyPodForTests()
	req := createReq(emptyPod, "Pod")
	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(req)
	suite.Nil(err)
	suite.True(reflect.DeepEqual(suite.emptyWorkloadResource, workLoadResource))
}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_PodWithEmptyPropertiesAdmissionReqWithMatchingObject_AsExpected() {
	emptyPropertiesPod := createPodWithEmptyPropertiesForTests()
	req := createReq(emptyPropertiesPod, "Pod")
	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(req)
	suite.Nil(err)
	suite.True(reflect.DeepEqual(suite.emptyWorkloadResource, workLoadResource))
}

func (suite *TestSuite) Test_GetWorkloadResourceFromAdmissionRequest_BadFormat_Error() {
	suite.podReq.Object.Raw = []byte("{ \"a\" : \"badFormat\"")
	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(suite.podReq)
	suite.Nil(workLoadResource)
	suite.NotEqual(nil, err)
}

func (suite *TestSuite) Test_GetWorkloadResourceFromAdmissionRequest_RequestNull_Error() {
	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(nil)
	suite.Nil(workLoadResource)
	suite.True(errors.Is(err, _errInvalidAdmission))
}

func (suite *TestSuite) Test_GetWorkloadResourceFromAdmissionRequest_EmptyRawObject_Error() {
	suite.podReq.Object.Raw = []byte{}
	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(suite.podReq)
	suite.Nil(workLoadResource)
	suite.True(errors.Is(err, _errWorkloadResourceEmpty))
}

func (suite *TestSuite) Test_GetWorkloadResourceFromAdmissionRequest_NilRawObject_Error() {
	suite.podReq.Object.Raw = nil
	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(suite.podReq)
	suite.Nil(workLoadResource)
	suite.True(errors.Is(err,_errWorkloadResourceEmpty))
}

func (suite *TestSuite) Test_GetWorkloadResourceFromAdmissionRequest_NotWorkloadResourceKindRequest_Error() {
	suite.podReq.Kind.Kind = "NotWorkloadResource"
	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(suite.podReq)
	suite.Nil(workLoadResource)
	suite.True(strings.Contains(err.Error(), "NotWorkloadResource is unsupported kind of workload resource"))
}

func TestExtractWorkloadResourceFromAdmissionRequest(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func createPodSpec() corev1.PodSpec {
	return corev1.PodSpec{
		Containers:         _actualContainers,
		InitContainers:     _actualInitContainers,
		ImagePullSecrets:   _actualImagePullSecrets,
		ServiceAccountName: _serviceAccountName,
	}
}

func createMetadata() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:            _name,
		Namespace:       _namespace,
		Annotations:     _annotation,
		OwnerReferences: _actualOwnerReferences,
	}
}

func createFullPodForTests() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: createMetadata(),
		Spec:       createPodSpec(),
	}
}

func createFullDeploymentForTests() *apps1.Deployment {
	return &apps1.Deployment{
		ObjectMeta: createMetadata(),
		Spec: apps1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: createPodSpec(),
			},
		},
	}
}

func createFullReplicaSetForTests() *apps1.ReplicaSet {
	return &apps1.ReplicaSet{
		ObjectMeta: createMetadata(),
		Spec: apps1.ReplicaSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: createPodSpec(),
			},
		},
	}
}

func createFullStatefulSetForTests() *apps1.StatefulSet {
	return &apps1.StatefulSet{
		ObjectMeta: createMetadata(),
		Spec: apps1.StatefulSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: createPodSpec(),
			},
		},
	}
}

func createFullJobForTests() *batch1.Job {
	return &batch1.Job{
		ObjectMeta: createMetadata(),
		Spec: batch1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: createPodSpec(),
			},
		},
	}
}

func createFullCronJobForTests() *batch1.CronJob {
	return &batch1.CronJob{
		ObjectMeta: createMetadata(),
		Spec: batch1.CronJobSpec{
			JobTemplate: batch1.JobTemplateSpec{
				Spec: batch1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: createPodSpec(),
					},
				},
			},
		},
	}
}

func createFullDaemonSetForTests() *apps1.DaemonSet {
	return &apps1.DaemonSet{
		ObjectMeta: createMetadata(),
		Spec: apps1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: createPodSpec(),
			},
		},
	}
}

func createFullReplicationControllerForTests() *corev1.ReplicationController {
	return &corev1.ReplicationController{
		ObjectMeta: createMetadata(),
		Spec: corev1.ReplicationControllerSpec{
			Template: &corev1.PodTemplateSpec{
				Spec: createPodSpec(),
			},
		},
	}
}

func createFullWorkloadResourceForTests() *WorkloadResource {
	return newWorkLoadResource(newObjectMetadata(_name, _namespace, _annotation, _expectedOwnerReferences),
		newSpec(_expectedContainers, _expectedInitContainers, _expectedImagePullSecrets, _serviceAccountName))
}
func createEmptyPodForTests() *corev1.Pod {
	return &corev1.Pod{}
}

func createPodWithEmptyPropertiesForTests() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       corev1.PodSpec{},
	}
}

func createEmptyWorkloadResourceForTests() *WorkloadResource {
	return newWorkLoadResource(newObjectMetadata("", "", nil, nil),
		newEmptySpec())
}

func createReq(resource interface{}, kind string) *admission.Request {
	bytes, err := json.Marshal(resource)
	if err != nil {
		log.Fatal(err)
	}
	return &admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Name: "",
			Kind: metav1.GroupVersionKind{
				Kind:    kind,
				Group:   "v1",
				Version: "",
			},
			Object: runtime.RawExtension{
				Raw: bytes,
			},
		},
	}
}

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
	"testing"
)

var (
	containersOriginal = []corev1.Container{
		{
			Name:  "containerTest",
			Image: "image.com",
		},
	}
	containersNew = []Container{
		{
			Name:  "containerTest",
			Image: "image.com",
		},
	}
	initContainersOriginal = []corev1.Container{
		{
			Name:  "initContainerTest",
			Image: "image.com",
		},
	}
	initContainersNew = []Container{
		{
			Name:  "initContainerTest",
			Image: "image.com",
		},
	}
	name             = "podTest"
	imagePullSecrets = []corev1.LocalObjectReference{
		{
			Name: "secret",
		},
	}
	serviceAccountName = "podServiceAccount"
	namespace          = "podNameSpace"
	annotation         = map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	ownerReferences = []OwnerReference{
		{
			APIVersion: "",
			Kind:       "",
			Name:       "",
		},
	}
)


type TestSuite struct {
	suite.Suite
	pod        *corev1.Pod
	replicationController *corev1.ReplicationController
	deployment *apps1.Deployment
	replicaSet *apps1.ReplicaSet
	statefulSet *apps1.StatefulSet
	daemonSet *apps1.DaemonSet
	job *batch1.Job
	cronJob *batch1.CronJob
	workloadResource *WorkloadResource
	emptyPod *corev1.Pod
	emptyWorkloadResource *WorkloadResource

	podReq *admission.Request
	emptyPodReq *admission.Request
	deploymentReq *admission.Request
	replicaSetReq *admission.Request
	statefulSetReq *admission.Request
	replicationControllerReq *admission.Request
	daemonSetReq *admission.Request
	jobReq *admission.Request
	cronJobReq *admission.Request

	extractor Extractor




}

func (suite *TestSuite) SetupTest() {
	suite.pod = createFullPodForTests()
	suite.emptyPod = createEmptyPodForTests()
	suite.workloadResource = createFullWorkloadResourceForTests()
	suite.emptyWorkloadResource = createEmptyWorkloadResourceForTests()
	suite.deployment = createFullDeploymentForTests()
	suite.daemonSet = createFullDaemonSetForTests()
	suite.replicaSet = createFullReplicaSetForTests()
	suite.statefulSet = createFullStatefulSetForTests()
	suite.replicationController = createFullReplicationControllerForTests()
	suite.job = createFullJobForTests()
	suite.cronJob = createFullCronJobForTests()

	suite.podReq = createReq(suite.pod,"Pod")
	suite.emptyPodReq = createReq(suite.emptyPodReq,"Pod")
	suite.deploymentReq = createReq(suite.deployment,"Deployment")
	suite.replicaSetReq = createReq(suite.replicaSet,"ReplicaSet")
	suite.statefulSetReq = createReq(suite.statefulSet,"StatefulSet")
	suite.replicationControllerReq = createReq(suite.replicationController,"ReplicationController")
	suite.daemonSetReq = createReq(suite.daemonSet,"DaemonSet")
	suite.jobReq = createReq(suite.jobReq,"Job")
	suite.cronJobReq = createReq(suite.cronJobReq,"CronJob")

	suite.extractor = *NewExtractor(instrumentation.NewNoOpInstrumentationProvider())



}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_PodAdmissionReqWithMatchingObject_AsExpected() {

	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(suite.podReq)
	suite.Nil(err)
	reflect.DeepEqual(suite.workloadResource, workLoadResource)
}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_DeploymentAdmissionReqWithMatchingObject_AsExpected() {

	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(suite.deploymentReq)
	suite.Nil(err)
	reflect.DeepEqual(suite.workloadResource, workLoadResource)
}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_ReplicaSetAdmissionReqWithMatchingObject_AsExpected() {

	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(suite.replicaSetReq)
	suite.Nil(err)
	reflect.DeepEqual(suite.workloadResource, workLoadResource)
}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_ReplicationControllerAdmissionReqWithMatchingObject_AsExpected() {

	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(suite.replicationControllerReq)
	suite.Nil(err)
	reflect.DeepEqual(suite.workloadResource, workLoadResource)
}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_StatefulSetAdmissionReqWithMatchingObject_AsExpected() {

	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(suite.statefulSetReq)
	suite.Nil(err)
	reflect.DeepEqual(suite.workloadResource, workLoadResource)
}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_DaemonSetAdmissionReqWithMatchingObject_AsExpected() {

	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(suite.daemonSetReq)
	suite.Nil(err)
	reflect.DeepEqual(suite.workloadResource, workLoadResource)
}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_JobAdmissionReqWithMatchingObject_AsExpected() {

	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(suite.jobReq)
	suite.Nil(err)
	reflect.DeepEqual(suite.workloadResource, workLoadResource)
}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_CronJobAdmissionReqWithMatchingObject_AsExpected() {

	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(suite.jobReq)
	suite.Nil(err)
	reflect.DeepEqual(suite.workloadResource, workLoadResource)
}

func (suite *TestSuite) Test_ExtractWorkloadResourceFromAdmissionRequest_EmptyPodAdmissionReqWithMatchingObject_AsExpected() {

	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(suite.emptyPodReq)
	suite.Nil(err)
	reflect.DeepEqual(suite.emptyWorkloadResource, workLoadResource)
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
	suite.Equal(errors.New(_errMsgInvalidAdmission), err)
}

func (suite *TestSuite) Test_GetWorkloadResourceFromAdmissionRequest_EmptyRawObject_Error() {
	suite.podReq.Object.Raw = []byte{}
	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(suite.podReq)
	suite.Nil(workLoadResource)
	suite.Equal(errors.New(_errMsgObjectNotFound), err)
}

func (suite *TestSuite) Test_GetWorkloadResourceFromAdmissionRequest_NotWorkloadResourceKindRequest_Error() {
	suite.podReq.Kind.Kind = "NotWorkloadResource"
	workLoadResource, err := suite.extractor.ExtractWorkloadResourceFromAdmissionRequest(suite.podReq)
	suite.Nil(workLoadResource)
	suite.Equal(errors.New(_errMsgUnexpectedResource), err)
}

func TestUnmarshalPod(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func createPodSpec() corev1.PodSpec {
	return corev1.PodSpec{
		Containers:         containersOriginal,
		InitContainers:     initContainersOriginal,
		ImagePullSecrets:   imagePullSecrets,
		ServiceAccountName: serviceAccountName,
	}
}

func createMetadata() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        name,
		Namespace:   namespace,
		Annotations: annotation,
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
	return newWorkLoadResource(newObjectMetadata(name, namespace, annotation, ownerReferences),
		newSpec(containersNew, initContainersNew, imagePullSecrets, serviceAccountName))
}
func createEmptyPodForTests() *corev1.Pod {
	return &corev1.Pod{}
}


func createEmptyWorkloadResourceForTests() *WorkloadResource {
	return newWorkLoadResource(newObjectMetadata("", "", nil, nil),
		newSpec(nil, nil, nil, ""))
}

func createReq(resource interface{},kind string) *admission.Request{
	bytes, err := json.Marshal(resource)
	if err != nil {
		log.Fatal(err)
	}
	return &admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Name: "",
			Kind: metav1.GroupVersionKind{
				Kind:    kind,
				Group:   "",
				Version: "",
			},
			Object: runtime.RawExtension{
				Raw: bytes,
			},
		},
	}
}


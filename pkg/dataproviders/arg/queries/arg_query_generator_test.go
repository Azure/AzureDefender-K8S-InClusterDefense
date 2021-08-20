package queries

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const _expectedContainerVulnerabilityScanResultsQuery = `
securityresources
 | where type == 'microsoft.security/assessments/subassessments'
 | where id matches regex  '(.+?)/providers/Microsoft.Security/assessments/dbd0cb49-b563-45e7-9724-889e799fa648/'
 | extend digest = tostring(properties.additionalData.imageDigest)
 | extend repository = tostring(properties.additionalData.repositoryName)
 | extend registry = tostring(properties.additionalData.registryHost)
 | where   registry =~ "tomer.azurecr.io" and repository =~ "test-image" and digest == "sha256:763bdd5314d126766d54cec7585f361c8c1429a2c51c818f0e7d0cab21a1481e"
 | parse id with  registryResourceId '/providers/Microsoft.Security/assessments/' *
 | extend scanFindingSeverity = tostring(properties.status.severity), scanStatus = tostring(properties.status.code)
 | extend id = tostring(properties.id), patchable = tobool(properties.additionalData.patchable)
 | project registry, repository, digest, scanStatus, scanFindingSeverity, id, patchable
`


// Tests query temaplte it self and it's generation
func Test_QueryGenerator_GenerateImageVulnerabilityScanQuery(t *testing.T) {
	generator, err := CreateARGQueryGenerator()
	parameters := &ContainerVulnerabilityScanResultsQueryParameters{
		Registry: "tomer.azurecr.io",
		Repository: "test-image",
		Digest: "sha256:763bdd5314d126766d54cec7585f361c8c1429a2c51c818f0e7d0cab21a1481e",
	}
	query, err := generator.GenerateImageVulnerabilityScanQuery(parameters)
	assert.Nil(t, err)
	assert.Equal(t, _expectedContainerVulnerabilityScanResultsQuery, query)
}

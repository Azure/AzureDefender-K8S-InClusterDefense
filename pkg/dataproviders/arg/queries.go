package arg

const _containerVulnerabilityScanResultsQuery = `
securityresources
 | where type == 'microsoft.security/assessments/subassessments'
 | where id matches regex  '(.+?)/providers/Microsoft.Security/assessments/dbd0cb49-b563-45e7-9724-889e799fa648/'
 | extend imageDigest = tostring(properties.additionalData.imageDigest)
 | extend repository = tostring(properties.additionalData.repositoryName)

 | parse id with  registryResourceId '/providers/Microsoft.Security/assessments/' *
 | parse registryResourceId with  * "/providers/Microsoft.ContainerRegistry/registries/" registryName
 | extend imageDigest = tostring(properties.additionalData.imageDigest)
 | extend repository = tostring(properties.additionalData.repositoryName)
 | extend scanFindingSeverity = tostring(properties.status.severity), scanStatus = tostring(properties.status.code)
 | summarize scanFindingSeverityCount = count() by scanFindingSeverity, scanStatus, registryResourceId, registryName, repository, imageDigest
 | summarize severitySummary = make_bag(pack(scanFindingSeverity, scanFindingSeverityCount)) by registryResourceId, registryName, repository, imageDigest, scanStatus
`

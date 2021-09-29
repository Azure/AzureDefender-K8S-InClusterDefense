package k8sazuredefenderblockresourceswithsecrets

# This violation checks if the resource contain secrets.
violation[{"msg": msg, "details": details}] {
    weaknesses := getResourceweaknesses(input.review)
    weakness := weaknesses.CredScanInfo[_]
    weakness.MatchingConfidence > input.parameters.matchingConfidenceThresholdForExcludingResourceWithSecrets
    msg := sprintf("secret found in the resource. The secret type is: <%v>", [weakness.credentialInfo.name])
    details := weakness
    }

# Gets review object and returns unnmarshelled scan resulsts (i.e. as array of scan results).
getResourceweaknesses(review) = weaknesses{
    scanResults := review.object.metadata.annotations["credScan.scan.info"]
    weaknesses := json.unmarshal(scanResults)
  }
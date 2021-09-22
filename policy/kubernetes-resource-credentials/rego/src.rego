# code for rego playground
package k8sazuredefenderblockresourceswithsecrets

default threshold = 0

violation[{"msg": msg, "details": details}] {
    weaknesses := getResourceweaknesses(input)
    weakness := weaknesses.CredScanInfo[_]
    weakness.MatchingConfidence > threshold
    msg := sprintf("secret found in the resource. The secret type is: <%v>", [weakness.credentialInfo.name])
    details := weakness
}

getResourceweaknesses(review) = weaknesses{
    scanResults := review.metadata.annotations["credScan.scan.info"]
    weaknesses := json.unmarshal(scanResults)
}
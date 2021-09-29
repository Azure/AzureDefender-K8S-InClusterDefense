package k8sazuredefenderblockresourceswithsecrets

# Check that if the annotations are empty then there is no violation.
test_input_no_annotations {
    input := { "review": input_no_annotations}
    results := violation with input as input
    count(results) == 0
}
# Check that if there is one secret then there is one violation.
test_input_one_secret {
    input := { "review": input_one_secret}
    results := violation with input as input
    count(results) == 1
}
# Check that if there is two secrets then there is two violation.
test_input_two_secrets {
    input := { "review": input_two_secrets}
    results := violation with input as input
    count(results) == 2
}
# Check that if there is one secret below threshold then there is no violation.
test_input_below_threshold {
    input := { "review": input_below_threshold}
    results := violation with input as input
    count(results) == 0
}
# Check that if there is one secret below and one secret above threshold then there is one violation.
test_input_threshold {
    input := { "review": input_threshold}
    results := violation with input as input
    count(results) == 1
}





input_no_annotations =
{
    "parameters": {
        "matchingConfidenceThresholdForExcludingResourceWithSecrets": 60,
        "sevirity": {
            "High": 0
        }
    },
    "review": {
        "object": {
            "metadata": {
                "annotations": {
                    "azuredefender.io/containers.vulnerability.scan.info": "{\"generatedTimestamp\":\"2021-05-04T23:53:20Z\",\"containers\":[{\"name\":\"testContainer\",\"image\":{\"name\":\"tomer.azurecr.io/core/app:4.6\",\"digest\":\"sha256:4a\"},\"scanStatus\":\"unhealthyScan\",\"scanFindings\":[{\"patchable\":true,\"id\":\"123\",\"severity\":\"High\"},{\"patchable\":true,\"id\":\"124\",\"severity\":\"High\"},{\"patchable\":true,\"id\":\"125\",\"severity\":\"High\"}]}]}",
                	"credScan.scan.info": "{\"generatedTimestamp\":\"2021-09-21T17:36:39.7945965Z\",\"CredScanInfo\":[]}"
                },
                "creationTimestamp": "2021-05-04T23:53:20Z"
            }
        }
    }
}

input_one_secret =
{
    "parameters": {
        "matchingConfidenceThresholdForExcludingResourceWithSecrets": 60,
        "sevirity": {
            "High": 0
        }
    },
    "review": {
        "object": {
            "metadata": {
                "annotations": {
                    "azuredefender.io/containers.vulnerability.scan.info": "{\"generatedTimestamp\":\"2021-05-04T23:53:20Z\",\"containers\":[{\"name\":\"testContainer\",\"image\":{\"name\":\"tomer.azurecr.io/core/app:4.6\",\"digest\":\"sha256:4a\"},\"scanStatus\":\"unhealthyScan\",\"scanFindings\":[{\"patchable\":true,\"id\":\"123\",\"severity\":\"High\"},{\"patchable\":true,\"id\":\"124\",\"severity\":\"High\"},{\"patchable\":true,\"id\":\"125\",\"severity\":\"High\"}]}]}",
                	"credScan.scan.info": "{\"generatedTimestamp\":\"2021-09-21T17:36:39.7945965Z\",\"CredScanInfo\":[{\"credentialInfo\":{\"name\":\"Azure Storage Account Access Key\"},\"MatchingConfidence\":74.96325,\"ScanStatus\":\"healthy\"}]}"
                },
                "creationTimestamp": "2021-05-04T23:53:20Z"
            }
        }
    }
}

input_two_secrets =
{
    "parameters": {
        "matchingConfidenceThresholdForExcludingResourceWithSecrets": 60,
        "sevirity": {
            "High": 0
        }
    },
    "review": {
        "object": {
            "metadata": {
                "annotations": {
                    "azuredefender.io/containers.vulnerability.scan.info": "{\"generatedTimestamp\":\"2021-05-04T23:53:20Z\",\"containers\":[{\"name\":\"testContainer\",\"image\":{\"name\":\"tomer.azurecr.io/core/app:4.6\",\"digest\":\"sha256:4a\"},\"scanStatus\":\"unhealthyScan\",\"scanFindings\":[{\"patchable\":true,\"id\":\"123\",\"severity\":\"High\"},{\"patchable\":true,\"id\":\"124\",\"severity\":\"High\"},{\"patchable\":true,\"id\":\"125\",\"severity\":\"High\"}]}]}",
                	"credScan.scan.info": "{\"generatedTimestamp\":\"2021-09-21T17:36:39.7945965Z\",\"CredScanInfo\":[{\"credentialInfo\":{\"name\":\"Azure Storage Account Access Key\"},\"MatchingConfidence\":74.96325,\"ScanStatus\":\"healthy\"},{\"credentialInfo\":{\"name\":\"General Password\"},\"MatchingConfidence\":99.9,\"ScanStatus\":\"unhealthy\"}]}"
                },
                "creationTimestamp": "2021-05-04T23:53:20Z"
            }
        }
    }
}

input_below_threshold =
{
    "parameters": {
        "matchingConfidenceThresholdForExcludingResourceWithSecrets": 75,
        "sevirity": {
            "High": 0
        }
    },
    "review": {
        "object": {
            "metadata": {
                "annotations": {
                    "azuredefender.io/containers.vulnerability.scan.info": "{\"generatedTimestamp\":\"2021-05-04T23:53:20Z\",\"containers\":[{\"name\":\"testContainer\",\"image\":{\"name\":\"tomer.azurecr.io/core/app:4.6\",\"digest\":\"sha256:4a\"},\"scanStatus\":\"unhealthyScan\",\"scanFindings\":[{\"patchable\":true,\"id\":\"123\",\"severity\":\"High\"},{\"patchable\":true,\"id\":\"124\",\"severity\":\"High\"},{\"patchable\":true,\"id\":\"125\",\"severity\":\"High\"}]}]}",
                	"credScan.scan.info": "{\"generatedTimestamp\":\"2021-09-21T17:36:39.7945965Z\",\"CredScanInfo\":[{\"credentialInfo\":{\"name\":\"Azure Storage Account Access Key\"},\"MatchingConfidence\":74.96325,\"ScanStatus\":\"healthy\"}]}"
                },
                "creationTimestamp": "2021-05-04T23:53:20Z"
            }
        }
    }
}

input_threshold =
{
    "parameters": {
        "matchingConfidenceThresholdForExcludingResourceWithSecrets": 75,
        "sevirity": {
            "High": 0
        }
    },
    "review": {
        "object": {
            "metadata": {
                "annotations": {
                    "azuredefender.io/containers.vulnerability.scan.info": "{\"generatedTimestamp\":\"2021-05-04T23:53:20Z\",\"containers\":[{\"name\":\"testContainer\",\"image\":{\"name\":\"tomer.azurecr.io/core/app:4.6\",\"digest\":\"sha256:4a\"},\"scanStatus\":\"unhealthyScan\",\"scanFindings\":[{\"patchable\":true,\"id\":\"123\",\"severity\":\"High\"},{\"patchable\":true,\"id\":\"124\",\"severity\":\"High\"},{\"patchable\":true,\"id\":\"125\",\"severity\":\"High\"}]}]}",
                	"credScan.scan.info": "{\"generatedTimestamp\":\"2021-09-21T17:36:39.7945965Z\",\"CredScanInfo\":[{\"credentialInfo\":{\"name\":\"Azure Storage Account Access Key\"},\"MatchingConfidence\":74.96325,\"ScanStatus\":\"healthy\"},{\"credentialInfo\":{\"name\":\"General Password\"},\"MatchingConfidence\":99.9,\"ScanStatus\":\"unhealthy\"}]}"
                },
                "creationTimestamp": "2021-05-04T23:53:20Z"
            }
        }
    }
}
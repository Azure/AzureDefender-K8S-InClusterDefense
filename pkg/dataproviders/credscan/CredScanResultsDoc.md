# Azure Defender K8S In Cluster Defense Credential Scan Results Documentation

## Results Format From CredScan Server

---
- scanProfileName
- credentialInfo
  - id
  - name
  - description
  - helpLink
  - tags
- matchingConfidence
- match
  - matchPrefix
  - matchContext
  - matchLongPrefix
  - matchValue
  - matchPrefixInline
  - innerMatchShortContext
  - matchShortContext
  - matchPostfix
  - matchLines
  - matchPrevLine
  - contextStartIndex
  - prefixStartIndex
  - longProximity
  - shortProximity
  - components
  - matchName
  - startIndex
  - length
  - matchLongPostfix
  - lineStartIndex
  - lastLineStartIndex
  - location
  - matchedBy
  - isFiltered
  - filteredBy
  - isTriaged
  - triagedBy
- patternName
- rankerExpressionGroupName
- components
- location
---


## Fields Used in CredScan Demo
- credentialInfo
  - name
- matchingConfidence

**In example**
- credentialInfo
  - name: "General Password"
- matchingConfidence: 99.9
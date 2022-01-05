{{/*
Generate certificates for redis cache
Get 2 arguments - service name and expire duration as strings
*/}}
{{- define "self.gen-certs" -}}
{{- $serviceName := index . 0 }}
{{- $expireDuration := index . 1 }}
{{- $certsDuration := ($expireDuration | int) -}}
{{- $ca := genCA "custom-ca" $certsDuration -}}
{{- $altNames := list ( $serviceName ) -}}
{{- $cert := genSignedCert $serviceName nil $altNames $certsDuration $ca -}}
ca.cert: {{ $ca.Cert | b64enc}}
tls.crt: {{ $cert.Cert | b64enc }}
tls.key: {{ $cert.Key | b64enc }}
{{- end -}}


{{/*
Generate secret data containing certs files and password.
No indentation is needed - just place it under the data key in your secret.
please note that there is no watch or verification of the certs.
Get 4 arguments - secret name, release namespace, service name and expire duration as strings
*/}}
{{- define "secret-data.deploy-certs" -}}
{{- $secret_name := index . 0 }}
{{- $ReleaseNamespace := index . 1 }}
{{- $serviceName := index . 2 }}
{{- $expireDuration := index . 3 }}
  # try to get the old secret
  {{- $old_sec := lookup "v1" "Secret" $ReleaseNamespace $secret_name -}}
  # check, if a secret is already set
  {{- if or (not $old_sec) (not $old_sec.data) -}}
  # if not set, then generate new certs and a new password
{{ ( include "self.gen-certs" (list $serviceName $expireDuration)) | indent 2}}
  {{- else -}}
  # if set, then use the old value
  ca.cert: {{ index $old_sec.data "ca.cert" }}
  tls.crt: {{ index $old_sec.data "tls.crt" }}
  tls.key: {{ index $old_sec.data "tls.key" }}
  {{- end -}}
{{- end -}}


{{/*
Generate secret data containing password.
No indentation is needed - just place it under the data key in your secret.
Get 2 arguments - secret name and release namespace as strings
*/}}
{{- define "secret-data.deploy-pass" -}}
{{- $secret_name := index . 0 }}
{{- $ReleaseNamespace := index . 1 }}
  # try to get the old secret
  {{- $old_sec := lookup "v1" "Secret" $ReleaseNamespace $secret_name -}}
  # check, if a secret is already set
  {{- if or (not $old_sec) (not $old_sec.data) -}}
  # if not set, then generate a new password
  REDIS_PASS: {{ randAlphaNum 32 | b64enc | quote }}
  {{- else -}}
  # if set, then use the old value
  REDIS_PASS: {{ index $old_sec.data "REDIS_PASS" }}
  {{- end -}}
{{- end -}}
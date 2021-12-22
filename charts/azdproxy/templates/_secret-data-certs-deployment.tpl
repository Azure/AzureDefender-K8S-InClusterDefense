{{/*
Generate secret data containing certs files and passwprd.
No indentation is needed - just place it under the data key in your secret.
*/}}
{{- define "secret-data.deploy-certs" -}}
{{- $secret_name := printf "%s-redis-tls-certs" .Values.AzDProxy.prefixResourceDeployment -}}
  # try to get the old secret
  {{- $old_sec := lookup "v1" "Secret" .Release.Namespace $secret_name -}}
  # check, if a secret is already set
  {{- if or (not $old_sec) (not $old_sec.data) -}}
  # if not set, then generate new certs and a new password
{{ ( include "custom.gen-certs" . ) | indent 2}}
  {{- else -}}
  # if set, then use the old value
  ca.cert: {{ index $old_sec.data "ca.cert" }}
  tls.crt: {{ index $old_sec.data "tls.crt" }}
  tls.key: {{ index $old_sec.data "tls.key" }}
  REDIS_PASS: {{ index $old_sec.data "REDIS_PASS" }}
  {{- end -}}
{{- end -}}
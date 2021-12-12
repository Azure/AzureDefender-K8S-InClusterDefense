{{/*
Generate certificates for api server
*/}}
{{- define "custom-metrics.gen-certs" -}}
{{- $ca := genCA "custom-metrics-ca" 365 -}}
{{- $altNames := list ( "azure-defender-proxy-redis-service" ) -}}
{{- $cert := genSignedCert "azure-defender-proxy-redis-service" nil $altNames 365 $ca -}}
ca.cert: {{ $ca.Cert | b64enc}}
tls.crt: {{ $cert.Cert | b64enc }}
tls.key: {{ $cert.Key | b64enc }}
REDIS_PASS: {{ randAlphaNum 32 | b64enc | quote }}
{{- end -}}
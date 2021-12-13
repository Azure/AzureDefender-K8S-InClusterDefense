{{/*
Generate certificates for api server
*/}}
{{- define "custom.gen-certs" -}}
{{- $ca := genCA "custom-ca" 365 -}}
{{- $serviceName := tpl .Values.AzDProxy.cache.argDataProviderCacheConfiguration.address . -}}
{{- $altNames := list ( $serviceName ) -}}
{{- $cert := genSignedCert $serviceName nil $altNames 365 $ca -}}
ca.cert: {{ $ca.Cert | b64enc}}
tls.crt: {{ $cert.Cert | b64enc }}
tls.key: {{ $cert.Key | b64enc }}
REDIS_PASS: {{ randAlphaNum 32 | b64enc | quote }}
{{- end -}}
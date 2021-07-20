{{/*
Common labels
*/}}

{{- define "common.labels" -}}
# This should be the app name, reflecting the entire app.
app.kubernetes.io/name: {{ .Chart.Name }}
# This should be the chart name and version
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
# It is for finding all things managed by Tiller
app.kubernetes.io/managed-by: {{ .Release.Service }}
# It aids in differentiating between different instances of the same application
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
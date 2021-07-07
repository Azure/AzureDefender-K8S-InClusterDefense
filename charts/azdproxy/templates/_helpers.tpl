{{/*
Expand the name of the chart.
*/}}
{{- define "azdproxy.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}
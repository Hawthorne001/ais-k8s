{{/*
Chart label value.
*/}}
{{- define "aisauth.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Labels for the resources this chart owns.
*/}}
{{- define "aisauth.labels" -}}
helm.sh/chart: {{ include "aisauth.chart" . }}
app.kubernetes.io/name: authn
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}


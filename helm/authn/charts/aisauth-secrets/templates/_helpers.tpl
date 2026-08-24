{{/*
Chart label value.
*/}}
{{- define "aisauth-secrets.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Labels for the resources this chart owns.
*/}}
{{- define "aisauth-secrets.labels" -}}
helm.sh/chart: {{ include "aisauth-secrets.chart" . }}
app.kubernetes.io/name: authn
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "aisauth-secrets.adminSecretName" -}}
{{- .Values.adminSecretName | default (printf "%s-su-creds" .Release.Name) -}}
{{- end -}}

{{- define "aisauth-secrets.hmacSecretName" -}}
{{- .Values.hmacSecretName | default (printf "%s-jwt-signing-key" .Release.Name) -}}
{{- end -}}

{{- define "aisauth-secrets.rsaPassphraseSecretName" -}}
{{- .Values.rsaPassphraseSecretName | default (printf "%s-rsa-passphrase" .Release.Name) -}}
{{- end -}}

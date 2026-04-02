{{- define "webhook.fullname" -}}
{{ .Release.Name }}
{{- end -}}

{{- define "webhook.labels" -}}
app.kubernetes.io/name: {{ include "webhook.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "webhook.selfSignedIssuer" -}}
{{ include "webhook.fullname" . }}-selfsign
{{- end -}}

{{- define "webhook.rootCAIssuer" -}}
{{ include "webhook.fullname" . }}-ca
{{- end -}}

{{- define "webhook.rootCACertificate" -}}
{{ include "webhook.fullname" . }}-ca
{{- end -}}

{{- define "webhook.servingCertificate" -}}
{{ include "webhook.fullname" . }}-tls
{{- end -}}

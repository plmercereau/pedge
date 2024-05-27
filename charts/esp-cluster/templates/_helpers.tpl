{{- define "userPassword" -}}
{{- $context := . -}}
{{- $secretName := printf "%s" $context.name -}}

{{- $existingSecret := lookup "v1" "Secret" $context.namespace $secretName -}}
{{- if $existingSecret }}
  {{- $existingSecret.data.password | b64dec -}}
{{- else -}}
  {{- $randSecret := randAlphaNum 32 -}}
  {{- $randSecret }}
{{- end -}}
{{- end -}}

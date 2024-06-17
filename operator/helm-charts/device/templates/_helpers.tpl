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


{{- define "checksum" -}}
{{- $context := . -}}
{{- $out := "" -}}
{{- range $key, $value := $context -}}
{{- $out = printf "%s%s" $out $value | sha256sum -}}
{{- end -}}
{{- $out | sha256sum -}}
{{- end -}}


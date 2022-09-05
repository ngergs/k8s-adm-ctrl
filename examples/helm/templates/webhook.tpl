{{- define "webhookConfig" -}}
{{- if .webhooks -}}
apiVersion: admissionregistration.k8s.io/v1
kind: {{ .webhookType }}WebhookConfiguration
metadata:
  name: {{ .ctrl.name }}
  annotations:
    cert-manager.io/inject-ca-from: {{ .namespace }}/selfsigned-ca
    app.kubernetes.io/component: webhook
webhooks:
{{- $base := . -}}
{{- range $webhook := .webhooks }}
  - name: {{ $webhook.name }}.{{ $base.ctrl.name }}.{{ lower $base.webhookType }}
    matchPolicy: Equivalent
    admissionReviewVersions: {{ $base.ctrl.admissionReviewVersions }}
    rules: {{ $webhook.rules | toYaml | nindent 4 }}
    clientConfig:
      service:
        name: webhook-{{ $base.ctrl.name }}
        namespace: {{ $base.namespace }}
        {{- if $webhook.path}}
        path: {{ $webhook.path }}
        {{- end }}
    sideEffects: None
    timeoutSeconds: 10
{{- end }}
{{- end }}
{{- end }}

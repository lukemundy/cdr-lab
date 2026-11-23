{{- /* Helm's required accepts an empty list, so this check uses empty. */ -}}
{{- define "cdrlab.required" -}}
  {{- if empty (index . 1) -}}
    {{- fail (printf "set %s: %s" (index . 0) (index . 2)) -}}
  {{- end -}}
  {{- index . 1 -}}
{{- end -}}

{{- define "cdrlab.secretKey" -}}
valueFrom:
  secretKeyRef:
    name: {{ index . 0 }}
    key: {{ index . 1 }}
{{- end -}}

{{- define "cdrlab.containerSecurityContext" -}}
allowPrivilegeEscalation: false
readOnlyRootFilesystem: true
capabilities:
  drop: [ALL]
{{- end -}}

{{- define "cdrlab.image" -}}
{{- $image := index . 0 -}}
repository: {{ $image.repository }}
tag: {{ include "cdrlab.required" (list (printf "images.%s.tag" (index . 1)) ($image.tag | default (index . 2)) "the image tag, or a chart with an appVersion") | quote }}
{{- with $image.digest }}
digest: {{ . }}
{{- end }}
{{- end -}}

{{- /* With no host, a listener has its Service name and port; with a host and no port, it has port 443. */ -}}
{{- define "cdrlab.listener" -}}
{{- $v := index . 2 | default dict -}}
{{- $l := dict "host" (index . 0) "port" (index . 1) "external" false "path" (index . 3) -}}
{{- with $v.host -}}
  {{- $_ := set $l "host" . -}}
  {{- $_ := set $l "port" ($v.port | default 443) -}}
  {{- $_ := set $l "external" true -}}
{{- end -}}
{{- $hostHeader := $l.host -}}
{{- if ne (int $l.port) 443 -}}
  {{- $hostHeader = printf "%s:%d" $l.host (int $l.port) -}}
{{- end -}}
{{- $_ := set $l "hostHeader" $hostHeader -}}
{{- $_ := set $l "uri" (printf "https://%s" $hostHeader) -}}
{{- toYaml $l -}}
{{- end -}}

{{- /* The Register record of each preset brand, by the name of the brand, from the copy of the example seed. */ -}}
{{- define "cdrlab.presetRecords" -}}
{{- $records := dict -}}
{{- range (.Files.Get "files/preset-brands.json" | fromJson).dataHolders -}}
  {{- $brand := first .dataHolderBrands -}}
  {{- $_ := set $records ($brand.brandName | lower | replace " " "-") . -}}
{{- end -}}
{{- toYaml $records -}}
{{- end -}}

{{- define "cdrlab.secretMount" -}}
type: secret
name: {{ index . 0 }}
advancedMounts:
  {{- range $controller, $containers := index . 2 }}
  {{ $controller }}:
    {{- range $containers }}
    {{ . }}:
      - path: {{ index $ 1 }}
        readOnly: true
    {{- end }}
  {{- end }}
{{- end -}}

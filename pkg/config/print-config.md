<!-- PLEASE COPY THE FOLLOWING OUTPUT AS IS INTO THE GITHUB ISSUE (Don't forget to mask your usernames, passwords, IPs and other sensitive information when using this in an issue ) -->

### Runtime

AdguardHome-Sync Version: {{ .Version }}
Build: {{ .Build }}
OperatingSystem: {{ .OperatingSystem }}
Architecture: {{ .Architecture }}

### AdGuardHome sync aggregated config

```yaml
{{ .AggregatedConfig }}
```
{{- if .ConfigFilePath }}
### AdGuardHome sync unmodified config file

Config file path: {{ .ConfigFilePath }}

```yaml
{{ .ConfigFileContent }}
```
{{- end }}

### Environment Variables

```ini
{{ .EnvironmentVariables }}
```

<!-- END OF GITHUB ISSUE CONTENT -->
# SSM template

Utility for fetching SSM parameters using [templates](https://golang.org/pkg/text/template/)

## Examples

### Passing the template as argument
```
ssm-template -region us-west-2 '{{.Parameter "/ssm/parameter/key"}}'
```

### Piping the template
```
echo '
yaml:
  foo: {{.Parameter "/ssm/parameter/key/foo"}}
  bar: {{.Parameter "/ssm/parameter/bar"}}
' | ssm-template
```

### Exporting env variables
```
$(echo '
export FOO={{.Parameter "/ssm/parameter/key/foo"}}
export BAR={{.Parameter "/ssm/parameter/key/bar"}}
' | ssm-template)
```

### Getting values by path
```
echo '
{{ range $key, $value := .ParametersByPath "/ssm/parameter/path" }}
  {{ $key }}: {{ $value }}
{{ end }}
' | ssm-template
```


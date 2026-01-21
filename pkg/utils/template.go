package utils //nolint:revive // utils is a common package name

import (
	"bytes"
	"text/template"
)

// ExecuteTemplate parses and executes a template string with the given data.
func ExecuteTemplate(tmplStr string, data any) (string, error) {
	tmpl, err := template.New("tmpl").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

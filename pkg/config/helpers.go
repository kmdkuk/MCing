package config

import (
	"bytes"
	"maps"
	"text/template"
)

// MergeMap merges two maps.
func MergeMap(m1, m2 map[string]string) map[string]string {
	m := make(map[string]string)
	maps.Copy(m, m1)
	maps.Copy(m, m2)
	if len(m) == 0 {
		return nil
	}
	return m
}

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

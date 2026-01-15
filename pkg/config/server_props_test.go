package config

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGenServerProps(t *testing.T) {
	t.Run("equality", testEquality)
	t.Run("overwritten", testOverwritten)
	t.Run("nil", testNil)
}

func testEquality(t *testing.T) {
	userProps := map[string]string{
		"pvp":        "false",
		"difficulty": "hard",
	}
	expect, err := GenServerProps(userProps)
	if err != nil {
		t.Error("error occurred", err)
	}
	actual, err := GenServerProps(userProps)
	if err != nil {
		t.Error("error occurred", err)
	}
	if !cmp.Equal(expect, actual) {
		t.Error("The same output is not being produced for the same input.", cmp.Diff(expect, actual))
	}
}

func testOverwritten(t *testing.T) {
	userProps := map[string]string{
		"server-port": "25575",
	}
	notInclude := "server-port=25575"
	actual, err := GenServerProps(userProps)
	if err != nil {
		t.Error("error occurred", err)
	}
	if strings.Contains(actual, notInclude) {
		t.Error("Properties that should not have been updated have been overwritten.")
	}
}

//go:embed testdata/default.server.properties
var defaultProps string

func testNil(t *testing.T) {
	actual, err := GenServerProps(nil)
	if err != nil {
		t.Error("error occurred", err)
	}

	if !cmp.Equal(defaultProps, actual) {
		t.Error("The output is different when nil is given.", cmp.Diff(defaultProps, actual))
	}
}

func TestParseServerProps(t *testing.T) {
	t.Run("success", testSuccess)
}

func testSuccess(t *testing.T) {
	actual, err := ParseServerPropsFromPath("testdata/parse-test.server.properties")
	expect := map[string]string{
		"null":          "",
		"test":          "test",
		"contain-blank": "contain blank",
		"contain-equal": "contain=equal",
	}
	if err != nil {
		t.Error("error occurred", err)
	}
	if !cmp.Equal(actual, expect) {
		t.Error("The output is different when nil is given.", cmp.Diff(actual, expect))
	}
}

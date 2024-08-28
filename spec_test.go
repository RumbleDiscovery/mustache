package mustache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

var disabledTests = map[string]map[string]struct{}{
	"interpolation.json": {
		// disabled b/c Go uses "&#34;" in place of "&quot;"
		// both are valid escapings, and we validate the behavior in mustache_test.go
		"HTML Escaping":                      struct{}{},
		"Implicit Iterators - HTML Escaping": struct{}{},
		// Not currently compliant with null interpolation tests added in v1.2.1
		"Basic Null Interpolation":           struct{}{},
		"Triple Mustache Null Interpolation": struct{}{},
		"Ampersand Null Interpolation":       struct{}{},
	},
	"~lambdas.json": {
		"Interpolation":                        struct{}{},
		"Interpolation - Expansion":            struct{}{},
		"Interpolation - Alternate Delimiters": struct{}{},
		"Interpolation - Multiple Calls":       struct{}{},
		"Escaping":                             struct{}{},
		"Section - Alternate Delimiters":       struct{}{},
		"Inverted Section":                     struct{}{},
	},
	"~inheritance.json": {}, // not implemented
}

type specTest struct {
	Name        string            `json:"name"`
	Data        any               `json:"data"`
	Expected    string            `json:"expected"`
	Template    string            `json:"template"`
	Description string            `json:"desc"`
	Partials    map[string]string `json:"partials"`
}

type specTestSuite struct {
	Tests []specTest `json:"tests"`
}

func TestSpec(t *testing.T) {
	root := filepath.Join(os.Getenv("PWD"), "spec", "specs")
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("Could not find the specs folder at %s, ensure the submodule exists by running 'git submodule update --init'", root)
		}
		t.Fatal(err)
	}

	paths, err := filepath.Glob(root + "/*.json")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(paths)

	for _, path := range paths {
		_, file := filepath.Split(path)
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var suite specTestSuite
		err = json.Unmarshal(b, &suite)
		if err != nil {
			t.Fatal(err)
		}
		for _, test := range suite.Tests {
			runTest(t, file, &test)
		}
	}
}

type LambdaFn func(text string, render RenderFn) (string, error)

var lambdas = map[string]LambdaFn{
	"Section": func(text string, render RenderFn) (string, error) {
		if text == "{{x}}" {
			return "yes", nil
		}
		return "no", nil
	},
	"Section - Expansion": func(text string, render RenderFn) (string, error) {
		return render(fmt.Sprintf("%s{{planet}}%s", text, text))
	},
	"Section - Multiple Calls": func(text string, render RenderFn) (string, error) {
		return render(fmt.Sprintf("__%s__", text))
	},
}

func runTest(t *testing.T, file string, test *specTest) {
	disabled, ok := disabledTests[file]
	if ok {
		// Can disable a single test or the entire file.
		if _, ok := disabled[test.Name]; ok || len(disabled) == 0 {
			t.Logf("[%s %s]: Skipped", file, test.Name)
			return
		}
	}

	if file == "~lambdas.json" {
		lambda := lambdas[test.Name]
		(test.Data.(map[string]any))["lambda"] = lambda
	}
	var out string
	var oerr error
	if len(test.Partials) > 0 {
		tmpl, err := New().WithPartials(&StaticProvider{test.Partials}).CompileString(test.Template)
		if err != nil {
			t.Error(err)
		}
		out, oerr = tmpl.Render(test.Data)
	} else {
		t.Logf("test.Template = %s", test.Template)
		tmpl, err := New().CompileString(test.Template)
		if err != nil {
			t.Error(err)
		} else {
			out, oerr = tmpl.Render(test.Data)
		}
	}
	if oerr != nil {
		t.Errorf("[%s %s]: %s", file, test.Name, oerr.Error())
		return
	}
	if out != test.Expected {
		t.Errorf("[%s %s]: Expected %q, got %q", file, test.Name, test.Expected, out)
		return
	}

	t.Logf("[%s %s]: Passed", file, test.Name)
}

// Package v1api is a best-effort attempt to implement the old v1 API using the
// new v2 API.
package v1api

import (
	"fmt"
	"os"
	"path"

	"github.com/runZeroInc/mustache/v2"
)

// ParseString compiles a mustache template string. The resulting output can
// be used to efficiently render the template multiple times with different data
// sources.
func ParseString(data string) (*mustache.Template, error) {
	return ParseStringRaw(data, false)
}

// ParseStringRaw compiles a mustache template string. The resulting output can
// be used to efficiently render the template multiple times with different data
// sources.
func ParseStringRaw(data string, forceRaw bool) (*mustache.Template, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("couldn't parse string: %w", err)
	}
	partials := &mustache.FileProvider{
		Paths: []string{cwd},
	}

	return ParseStringPartialsRaw(data, partials, forceRaw)
}

// ParseStringPartials compiles a mustache template string, retrieving any
// required partials from the given provider. The resulting output can be used
// to efficiently render the template multiple times with different data
// sources.
func ParseStringPartials(data string, partials mustache.PartialProvider) (*mustache.Template, error) {
	return ParseStringPartialsRaw(data, partials, false)
}

// ParseStringPartialsRaw compiles a mustache template string, retrieving any
// required partials from the given provider. The resulting output can be used
// to efficiently render the template multiple times with different data
// sources.
func ParseStringPartialsRaw(data string, partials mustache.PartialProvider, forceRaw bool) (*mustache.Template, error) {
	escapeMode := mustache.EscapeHTML
	if forceRaw {
		escapeMode = mustache.Raw
	}
	return mustache.New().WithPartials(partials).WithEscapeMode(escapeMode).CompileString(data)
}

// ParseFile loads a mustache template string from a file and compiles it. The
// resulting output can be used to efficiently render the template multiple
// times with different data sources.
func ParseFile(filename string) (*mustache.Template, error) {
	dirname, _ := path.Split(filename)
	partials := &mustache.FileProvider{
		Paths: []string{dirname},
	}

	return ParseFilePartials(filename, partials)
}

// ParseFilePartials loads a mustache template string from a file, retrieving any
// required partials from the given provider, and compiles it. The resulting
// output can be used to efficiently render the template multiple times with
// different data sources.
func ParseFilePartials(filename string, partials mustache.PartialProvider) (*mustache.Template, error) {
	return ParseFilePartialsRaw(filename, false, partials)
}

// ParseFilePartialsRaw loads a mustache template string from a file, retrieving
// any required partials from the given provider, and compiles it. The resulting
// output can be used to efficiently render the template multiple times with
// different data sources.
func ParseFilePartialsRaw(filename string, forceRaw bool, partials mustache.PartialProvider) (*mustache.Template, error) {
	escapeMode := mustache.EscapeHTML
	if forceRaw {
		escapeMode = mustache.Raw
	}
	return mustache.New().WithPartials(partials).WithEscapeMode(escapeMode).CompileFile(filename)
}

// Render compiles a mustache template string and uses the the given data source
// - generally a map or struct - to render the template and return the output.
func Render(data string, context ...interface{}) (string, error) {
	return RenderRaw(data, false, context...)
}

// RenderRaw compiles a mustache template string and uses the the given data
// source - generally a map or struct - to render the template and return the
// output.
func RenderRaw(data string, forceRaw bool, context ...interface{}) (string, error) {
	return RenderPartialsRaw(data, nil, forceRaw, context...)
}

// RenderPartials compiles a mustache template string and uses the the given partial
// provider and data source - generally a map or struct - to render the template
// and return the output.
func RenderPartials(data string, partials mustache.PartialProvider, context ...interface{}) (string, error) {
	return RenderPartialsRaw(data, partials, false, context...)
}

// RenderPartialsRaw compiles a mustache template string and uses the the given
// partial provider and data source - generally a map or struct - to render the
// template and return the output.
func RenderPartialsRaw(data string, partials mustache.PartialProvider, forceRaw bool, context ...interface{}) (string, error) {
	tmpl := mustache.New()
	if forceRaw {
		tmpl = tmpl.WithEscapeMode(mustache.Raw)
	}
	if partials != nil {
		tmpl = tmpl.WithPartials(partials)
	}
	renderer, err := tmpl.CompileString(data)
	if err != nil {
		return "", err
	}
	return renderer.Render(context...)
}

// RenderInLayout compiles a mustache template string and layout "wrapper" and
// uses the given data source - generally a map or struct - to render the
// compiled templates and return the output.
func RenderInLayout(data string, layoutData string, context ...interface{}) (string, error) {
	return RenderInLayoutPartials(data, layoutData, nil, context...)
}

// RenderInLayoutPartials compiles a mustache template string and layout
// "wrapper" and uses the given data source - generally a map or struct - to
// render the compiled templates and return the output.
func RenderInLayoutPartials(data string, layoutData string, partials mustache.PartialProvider, context ...interface{}) (string, error) {
	layoutCmpl := mustache.New()
	if partials != nil {
		layoutCmpl.WithPartials(partials)
	}
	layoutTmpl, err := layoutCmpl.CompileString(layoutData)
	if err != nil {
		return "", err
	}
	cmpl := mustache.New()
	if partials != nil {
		cmpl.WithPartials(partials)
	}
	tmpl, err := cmpl.CompileString(data)
	if err != nil {
		return "", err
	}
	return tmpl.RenderInLayout(layoutTmpl, context...)
}

// RenderFile loads a mustache template string from a file and compiles it, and
// then uses the the given data source - generally a map or struct - to render
// the template and return the output.
func RenderFile(filename string, context ...interface{}) (string, error) {
	tmpl, err := mustache.New().CompileFile(filename)
	if err != nil {
		return "", err
	}
	return tmpl.Render(context...)
}

// RenderFileInLayout loads a mustache template string and layout "wrapper"
// template string from files and compiles them, and  then uses the the given
// data source - generally a map or struct - to render the compiled templates
// and return the output.
func RenderFileInLayout(filename string, layoutFile string, context ...interface{}) (string, error) {
	layoutTmpl, err := mustache.New().CompileFile(layoutFile)
	if err != nil {
		return "", err
	}

	tmpl, err := mustache.New().CompileFile(filename)
	if err != nil {
		return "", err
	}
	return tmpl.RenderInLayout(layoutTmpl, context...)
}

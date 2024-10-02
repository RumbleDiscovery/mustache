package mustache

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// PartialProvider comprises the behaviors required of a struct to be able to provide partials to the mustache rendering
// engine.
type PartialProvider interface {
	// Get accepts the name of a partial and returns the parsed partial, if it could be found; a valid but empty
	// template, if it could not be found; or nil and error if an error occurred (other than an inability to find
	// the partial).
	Get(name string) (string, error)
}

// FileProvider implements the PartialProvider interface by providing partials drawn from a filesystem. When a partial
// named `NAME`  is requested, FileProvider searches each listed path for a file named as `NAME` followed by any of the
// listed extensions. The default for `Paths` is to search the current working directory. The default for `Extensions`
// is to examine, in order, no extension; then ".mustache"; then ".stache". If Unsafe is set, partial names are allowed
// to begin with '.' or '..' after cleaning, meaning they can potentially refer to files outside any of the listed
// directory paths.
type FileProvider struct {
	Paths      []string
	Extensions []string
	Unsafe     bool
}

// Get accepts the name of a partial and returns the parsed partial.
func (fp *FileProvider) Get(name string) (string, error) {
	clean := name
	if !fp.Unsafe {
		// Use a '/' prefix so filepath.Clean can prevent a directory traversal
		cname := "/" + strings.Trim(name, "/\\")
		cname = strings.ReplaceAll(filepath.Clean(cname), "\\", "/")
		cname = strings.TrimLeft(cname, "/")
		if cname != name || cname == "" {
			return "", fmt.Errorf("can't use %s: %w", name, ErrUnsafePartialName) //nolint:all
		}
		clean = cname
	}

	var paths []string
	if fp.Paths != nil {
		paths = fp.Paths
	} else {
		paths = []string{""}
	}

	var exts []string
	if fp.Extensions != nil {
		exts = fp.Extensions
	} else {
		exts = []string{"", ".mustache", ".stache"}
	}

	var f *os.File
	var err error
	for _, p := range paths {
		for _, e := range exts {
			pname := filepath.Join(p, clean+e)
			f, err = os.Open(pname)
			if err == nil {
				break
			}
		}
	}

	if f == nil {
		return "", fmt.Errorf("%s: %w", name, ErrPartialNotFound)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("error reading partial %s: %w", name, err)
	}

	return string(data), nil
}

var _ PartialProvider = (*FileProvider)(nil)

// StaticProvider implements the PartialProvider interface by providing partials drawn from a map, which maps partial
// name to template contents.
type StaticProvider struct {
	Partials map[string]string
}

// Get accepts the name of a partial and returns the parsed partial.
func (sp *StaticProvider) Get(name string) (string, error) {
	if sp.Partials != nil {
		if data, ok := sp.Partials[name]; ok {
			return data, nil
		}
	}

	return "", nil
}

var _ PartialProvider = (*StaticProvider)(nil)

func (tmpl *Template) getPartials(partials PartialProvider, name, indent string) (*Template, error) {
	if partials == nil {
		return nil, ErrNoPartialProvider
	}
	data, err := partials.Get(name)
	if err != nil {
		return nil, err
	}

	// indent non empty lines
	r := regexp.MustCompile(`(?m:^(.+)$)`)
	data = r.ReplaceAllString(data, indent+"$1")

	return tmpl.parent.CompileString(data)
}

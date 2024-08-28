# Mustache template engine for Go

[![Go Doc](https://pkg.go.dev/badge/github.com/runZeroInc/mustache)](https://pkg.go.dev/github.com/runZeroInc/mustache)
[![Go Report Card](https://goreportcard.com/badge/github.com/runZeroInc/mustache)](https://goreportcard.com/report/github.com/runZeroInc/mustache)
[![codecov](https://codecov.io/gh/runZeroInc/mustache/branch/main/graph/badge.svg?token=S4xOabgqR8)](https://codecov.io/gh/runZeroInc/mustache)

<img src="images/logo.jpeg" alt="logo" width="100"/>

----

## Why a fork?

I forked [cbroglie/mustache](https://github.com/cbroglie/mustache) because it does not appear to be maintained, and I
wanted to add the following functionality:

- Add support for JSON and plain text escaping modes (for example, for templating e-mail, or JSON messages for Slack notifications).
- Add [a previously submitted patch for lambda support](https://github.com/cbroglie/mustache/pull/28).
- Add a fluent API.

I also wanted to clear up some security holes, including two found by fuzzing.

The goal is for this to be a robust, performant, standards-compliant Mustache template engine for Go, and for it to be
safe to allow end users to supply templates. Extensions to the templating language are generally not desired. If you
want more than Mustache offers, consider a [Handlebars](https://handlebarsjs.com/) implementation such
as [Mario](https://github.com/imantung/mario).

----

## CLI overview

```bash
% go get github.com/runZeroInc/mustache/...
% mustache
Usage:
  mustache [data] template [flags]

Examples:
  $ mustache data.yml template.mustache
  $ cat data.yml | mustache template.mustache
  $ mustache --layout wrapper.mustache data template.mustache
  $ mustache --overide over.yml data.yml template.mustache

Flags:
  -h, --help   help for mustache
  --layout     a file to use as the layout template
  --override   a data.yml file whose definitions supercede data.yml
%
```

----

## Package overview

This library is an implementation of the Mustache template language in Go.

### Mustache spec compliance

[mustache/spec](https://github.com/mustache/spec) contains the formal standard for Mustache, and it is included as a submodule (using v1.2.1) for testing compliance. All of the tests pass (big thanks to [kei10in](https://github.com/kei10in)), with the exception of the null interpolation tests added in v1.2.1. The optional inheritance and lambda support has not been fully implemented.

----

## Documentation

For more information about mustache, check out the [mustache project page](https://mustache.github.io/) or
the [mustache manual](http://mustache.github.com/mustache.5.html).

Also check out some [example mustache files](http://github.com/defunkt/mustache/tree/master/examples/).

----

## Installation

To install mustache.go, simply run `go get github.com/runZeroInc/mustache/...`. To use it in a program, use `import "github.com/runZeroInc/mustache"`

----

## Usage

Starting with version 2, a fluent API is provided, and compilation and rendering of templates is performed as separate
steps, with separate error returns. This makes it easier to distinguish between syntactically invalid templates, and
errors at render time.

First, use `mustache.New()` to obtain a Compiler. You can then set options on the compiler:

```go
cmpl := mustache.New()
cmpl.WithErrors(true)
cmpl.WithPartials(&FileProvider{
	Paths: []string{"/app/templates"},
	Extensions: []string{".html", ".mustache"}
})
cmpl.WithEscapeMode(mustache.EscapeHTML)
```

Then you can use the compiler you've configured to compile your template(s):

```go
tmpl1, err := cmpl.CompileString("This is {{mustache}}")
tmpl2, err := cmpl.CompileFile("main.mustache")
```

Finally, you can render the compiled templates using any number of contextual data objects, generally expected to be `map[string]any` or a `struct`:

```go
output, err := tmpl1.Render(map[string]string{"mustache":"awesome!"})
```

The compiler options can be chained together:

```go
tmpl, err := mustache.New().WithErrors(true).CompileString("This is {{mustache}}")
```

There are also two additional methods for using layouts (explained below); as well as several more that can provide a
custom Partial retrieval.

Unlike in the v1 API, the defaults for the compiler are intended to be safe, with no partial support -- you have to
provide a PartialProvider explicitly if you want to use partials. So by default you get:

 - No partials
 - No errors when data is missing from the context
 - HTML escaping

There are no longer functions to render a template without compiling to a `*Template` object. The engine always compiles
even if you throw the template away when you're done with it, so there's no speed benefit to having a non-compiling
option.

For more example usage, please see `mustache_test.go`

----

## Escaping

By default, mustache.go follows the official mustache HTML escaping rules. That is, if you enclose a variable with two
curly brackets, `{{var}}`, the contents are HTML-escaped. For instance, strings like `5 > 2` are converted to `5 &gt; 2`.
To use raw characters, use three curly brackets `{{{var}}}`.

This implementation of Mustache also allows you to run the engine in JSON mode, in which case the standard JSON quoting
rules are used. To do this, use `.WithEscapeMode(mustache.JSON)` to set the escape mode on the compiler. Note that the
JSON escaping rules are different from the rules used by Go's text/template.JSEscape, and do not guarantee that the JSON
will be safe to include as part of an HTML page. In JSON mode, references to objects and slices in the template will be rendered to JSON objects and arrays.

A third mode of `mustache.Raw` allows the use of Mustache templates to generate plain text, such as e-mail messages and
console application help text.

----

## Layouts

It is a common pattern to include a template file as a "wrapper" for other templates. The wrapper may include a header
and a footer, for instance. Mustache.go supports this pattern with the following method:

```go
(contentTemplate *Template) RenderInLayout(layoutTemplate *Template, context ...any) (string, error)
```

The layout must have a variable called `{{content}}`. For example, given the following files:

layout.html.mustache:

```html
<html>
<head><title>Hi</title></head>
<body>
{{{content}}}
</body>
</html>
```

template.html.mustache:

```html
<h1>Hello World!</h1>
```

...and suitable code to load and compile them:

```go
template, _ := mustache.New().CompileFile("template.html.mustache")
layout, _ := mustache.New().CompileFile("layout.html.mustache")
```

A call to `template.RenderInLayout(layout)` will produce:

```html
<html>
<head><title>Hi</title></head>
<body>
<h1>Hello World!</h1>
</body>
</html>
```

----

## Custom PartialProvider

Mustache supports user-defined repositories for mustache partials.

A `PartialProvider` is any object that responds to `Get(string) (*Template,error)`, and two examples are provided -- 
a `FileProvider` that loads files from disk, and a `StaticProvider` alias for a `map[string]string`. Using either 
of these is simple:

```go

fp := &FileProvider{
  Paths: []string{ "", "/opt/mustache", "templates/" },
  Extensions: []string{ "", ".stache", ".mustache" },
}

tmpl, err := mustache.New().WithPartials(fp).CompileString("This partial is loaded from a file: {{>foo}}")

sp := StaticProvider(map[string]string{
  "foo": "{{>bar}}",
  "bar": "some data",
})

tmpl, err := mustache.New().WithPartials(sp).CompileString("This partial is loaded from a map: {{>foo}}", sp)
```

----

## A note about method receivers

Mustache.go supports calling methods on objects, but you have to be aware of Go's limitations. For example, lets's say
you have the following type:

```go
type Person struct {
    FirstName string
    LastName string
}

func (p *Person) Name1() string {
    return p.FirstName + " " + p.LastName
}

func (p Person) Name2() string {
    return p.FirstName + " " + p.LastName
}
```

While they appear to be identical methods, `Name1` has a pointer receiver, and `Name2` has a value receiver. Objects of
type `Person`(non-pointer) can only access `Name2`, while objects of type `*Person`(person) can access both. This is by
design in the Go language.

So if you write the following:

```go
tmpl.Render("{{Name1}}", Person{"John", "Smith"})
```

It'll be blank. You either have to use `&Person{"John", "Smith"}`, or call `Name2`

## Supported features

- Variables
- Comments
- Change delimiter
- Sections (boolean, enumerable, and inverted)
- Partials
- Lambdas
- HTML, JSON or plain text output

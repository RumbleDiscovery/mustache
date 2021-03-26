# Mustache template engine for Go

[![Go Doc](https://pkg.go.dev/badge/github.com/RumbleDiscovery/mustache)](https://pkg.go.dev/github.com/RumbleDiscovery/mustache)
[![Go Report Card](https://goreportcard.com/badge/github.com/RumbleDiscovery/mustache)](https://goreportcard.com/report/github.com/RumbleDiscovery/mustache)
[![Build Status](https://img.shields.io/travis/RumbleDiscovery/mustache.svg)](https://travis-ci.com/github/RumbleDiscovery/mustache/settings)
[![codecov](https://codecov.io/gh/RumbleDiscovery/mustache/branch/main/graph/badge.svg?token=S4xOabgqR8)](https://codecov.io/gh/RumbleDiscovery/mustache)

<img src="images/logo.jpeg" alt="logo" width="100"/>

----

## Why a fork?

I forked [cbroglie/mustache](https://github.com/cbroglie/mustache) because it does not appear to be maintained, and I
wanted to add the following functionality:

- Add support for JSON and plain text escaping modes (for example, for templating e-mail, or JSON messages for Slack notifications).
- Add [a previously submitted patch for lambda support](https://github.com/cbroglie/mustache/pull/28).

I also wanted to clear up some security holes, including two found by fuzzing.

The goal is for this to be a robust, performant, standards-compliant Mustache template engine for Go, and for it to be
safe to allow end users to supply templates. Extensions to the templating language are generally not desired. If you
want more than Mustache offers, consider a [Handlebars](https://handlebarsjs.com/) implementation such
as [Mario](https://github.com/imantung/mario).

----

## CLI overview

```bash
➜  ~ go get github.com/RumbleDiscovery/mustache/...
➜  ~ mustache
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
➜  ~
```

----

## Package overview

This library is an implementation of the Mustache template language in Go.

### Mustache spec compliance

[mustache/spec](https://github.com/mustache/spec) contains the formal standard for Mustache, and it is included as a
submodule (using v1.1.3) for testing compliance. All of the tests pass (big thanks
to [kei10in](https://github.com/kei10in)), though the optional lambda support has not been fully implemented.

----

## Documentation

For more information about mustache, check out the [mustache project page](https://mustache.github.io/) or
the [mustache manual](http://mustache.github.com/mustache.5.html).

Also check out some [example mustache files](http://github.com/defunkt/mustache/tree/master/examples/).

----

## Installation

To install mustache.go, simply run `go get github.com/RumbleDiscovery/mustache/...`. To use it in a program, use `import "github.com/RumbleDiscovery/mustache"`

----

## Usage

There are four main methods in this package:

```go
Render(data string, context ...interface{}) (string, error)

RenderFile(filename string, context ...interface{}) (string, error)

ParseString(data string) (*Template, error)

ParseFile(filename string) (*Template, error)
```

There are also two additional methods for using layouts (explained below); as well as several more that can provide a
custom Partial retrieval.

The `Render` method takes a string and a data source, which is generally a map or struct, and returns the output string.
If the template file contains an error, the return value is a description of the error. There's a similar
method, `RenderFile`, which takes a filename as an argument and uses that for the template contents.

```go
data, err := mustache.Render("hello {{c}}", map[string]string{"c": "world"})
```

If you're planning to render the same template multiple times, you do it efficiently by compiling the template first:

```go
tmpl, _ := mustache.ParseString("hello {{c}}")
var buf bytes.Buffer
for i := 0; i < 10; i++ {
    tmpl.FRender(&buf, map[string]string{"c": "world"})
}
```

For more example usage, please see `mustache_test.go`

----

## Escaping

By default, mustache.go follows the official mustache HTML escaping rules. That is, if you enclose a variable with two
curly brackets, `{{var}}`, the contents are HTML-escaped. For instance, strings like `5 > 2` are converted to `5 &gt; 2`.
To use raw characters, use three curly brackets `{{{var}}}`.

This implementation of Mustache also allows you to run the engine in JSON mode, in which case the standard JSON quoting
rules are used. To do this, compile the template then set `tmpl.OutputMode = mustache.EscapeJSON`. Note that the JSON
escaping rules are different from the rules used by Go's text/template.JSEscape, and do not guarantee that the JSON will
be safe to include as part of an HTML page.

A third mode of `mustache.Raw` allows the use of Mustache templates to generate plain text, such as e-mail messages and
console application help text.

----

## Layouts

It is a common pattern to include a template file as a "wrapper" for other templates. The wrapper may include a header
and a footer, for instance. Mustache.go supports this pattern with the following two methods:

```go
RenderInLayout(data string, layout string, context ...interface{}) (string, error)

RenderFileInLayout(filename string, layoutFile string, context ...interface{}) (string, error)
```

The layout file must have a variable called `{{content}}`. For example, given the following files:

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

A call to `RenderFileInLayout("template.html.mustache", "layout.html.mustache", nil)` will produce:

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

Mustache.go has been extended to support a user-defined repository for mustache partials, instead of the default of
requiring file-based templates.

Several new top-level functions have been introduced to take advantage of this:

```go

func RenderPartials(data string, partials PartialProvider, context ...interface{}) (string, error)

func RenderInLayoutPartials(data string, layoutData string, partials PartialProvider, context ...interface{}) (string, error)

func ParseStringPartials(data string, partials PartialProvider) (*Template, error)

func ParseFilePartials(filename string, partials PartialProvider) (*Template, error)

```

A `PartialProvider` is any object that responds to `Get(string)
(*Template,error)`, and two examples are provided- a `FileProvider` that
recreates the old behavior (and is indeed used internally for backwards
compatibility); and a `StaticProvider` alias for a `map[string]string`. Using
either of these is simple:

```go

fp := &FileProvider{
  Paths: []string{ "", "/opt/mustache", "templates/" },
  Extensions: []string{ "", ".stache", ".mustache" },
}

tmpl, err := ParseStringPartials("This partial is loaded from a file: {{>foo}}", fp)

sp := StaticProvider(map[string]string{
  "foo": "{{>bar}}",
  "bar": "some data",
})

tmpl, err := ParseStringPartials("This partial is loaded from a map: {{>foo}}", sp)
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
mustache.Render("{{Name1}}", Person{"John", "Smith"})
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

package schema

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/jsonnet-libs/docsonnet/pkg/docsonnet"
)

func compareSchemas(a, b *jsonschema.Schema) error {
	if (a == nil) != (b == nil) {
		return fmt.Errorf("nil check failed; %v != %v", a, b)
	}
	if a.Schema != b.Schema {
		return fmt.Errorf("schemas do not match; expected %q, got %q", a.Schema, b.Schema)
	}
	if a.Title != b.Title {
		return fmt.Errorf("titles do not match; expected %q, got %q", a.Title, b.Title)
	}
	if a.Type != b.Type {
		return fmt.Errorf("types do not match; expected %q, got %q", a.Type, b.Type)
	}
	if len(a.Properties) != len(b.Properties) {
		return fmt.Errorf("properties do not have matching lengths; expected %d, got %d", len(a.Properties), len(b.Properties))
	}
	for p := range a.Properties {
		if err := compareSchemas(a.Properties[p], b.Properties[p]); err != nil {
			return fmt.Errorf("schemas for property %q do not match: %w", p, err)
		}
	}
	if len(a.PrefixItems) != len(b.PrefixItems) {
		return fmt.Errorf("prefix items do not have matching lengths; expected %d, got %d", len(a.PrefixItems), len(b.PrefixItems))
	}
	for i := range a.PrefixItems {
		if err := compareSchemas(a.PrefixItems[i], b.PrefixItems[i]); err != nil {
			return fmt.Errorf("schemas for prefix item %d do not match: %w", i, err)
		}
	}
	return nil
}

func TestConvert(t *testing.T) {
	for _, tc := range []struct {
		name    string
		jsonnet []byte
		schema  *jsonschema.Schema
	}{
		{
			name: "empty package",
			jsonnet: []byte([]byte(
				`local d = import "doc-util/main.libsonnet";
{
    '#': d.pkg(
      name='Empty',
      url='github.com/squat/schemasonnet/test/empty/main.libsonet',
      help='empty is a test packag.',
    ),
}
`),
			),
			schema: &jsonschema.Schema{
				Schema:     Draft2020_12,
				Title:      "Empty",
				Type:       schemaTypeObject,
				Properties: make(map[string]*jsonschema.Schema),
			},
		},
		{
			name: "package with function",
			jsonnet: []byte([]byte(
				`local d = import "doc-util/main.libsonnet";
{
    '#': d.pkg(
      name='Package with Function',
      url='github.com/squat/schemasonnet/test/package-with-function/main.libsonet',
      help='package-with-function is a test package.',
    ),
    "#greet": d.fn("greet greets you", [d.arg("who", d.T.string)]),
    greet(who):: "Hello, %s!" % who,
}
`),
			),
			schema: &jsonschema.Schema{
				Schema: Draft2020_12,
				Title:  "Package with Function",
				Type:   schemaTypeObject,
				Properties: map[string]*jsonschema.Schema{
					"greet": {
						Schema: Draft2020_12,
						Title:  "greet",
						Type:   schemaTypeObject,
						Properties: map[string]*jsonschema.Schema{
							propertyArguments: {
								Schema: Draft2020_12,
								Title:  titleArguments,
								Type:   schemaTypeArray,
								PrefixItems: []*jsonschema.Schema{
									{
										Schema: Draft2020_12,
										Title:  "who",
										Type:   schemaTypeString,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "package with nested function",
			jsonnet: []byte([]byte(
				`local d = import "doc-util/main.libsonnet";
{
    '#': d.pkg(
      name='Package with Nested Function',
      url='github.com/squat/schemasonnet/test/package-with-nested-function/main.libsonet',
      help='package-with-nested-function is a test package.',
    ),
    "#greet": d.fn("greet greets you", [d.arg("greeter", d.T.func)]),
    greet(greeter):: "Hello, %s!" % greeter(),
}
`),
			),
			schema: &jsonschema.Schema{
				Schema: Draft2020_12,
				Title:  "Package with Nested Function",
				Type:   schemaTypeObject,
				Properties: map[string]*jsonschema.Schema{
					"greet": {
						Schema: Draft2020_12,
						Title:  "greet",
						Type:   schemaTypeObject,
						Properties: map[string]*jsonschema.Schema{
							propertyArguments: {
								Schema: Draft2020_12,
								Title:  titleArguments,
								Type:   schemaTypeArray,
								PrefixItems: []*jsonschema.Schema{
									{
										Schema: Draft2020_12,
										Title:  "greeter",
										Type:   schemaTypeObject,
										Properties: map[string]*jsonschema.Schema{
											propertyArguments: {
												Schema: Draft2020_12,
												Title:  titleArguments,
												Type:   schemaTypeArray,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "package with object",
			jsonnet: []byte([]byte(
				`local d = import "doc-util/main.libsonnet";
{
    '#': d.pkg(
      name='Package with Object',
      url='github.com/squat/schemasonnet/test/package-with-object/main.libsonet',
      help='package-with-object is a test package.',
    ),
    "#foo": d.obj(""),
    foo:: {
	"#bar": d.val("string", "bar is a string"),
        bar:: "bar",
	"#baz": d.val("number", "baz is a number"),
        baz:: "baz",
    },
}
`),
			),
			schema: &jsonschema.Schema{
				Schema: Draft2020_12,
				Title:  "Package with Object",
				Type:   schemaTypeObject,
				Properties: map[string]*jsonschema.Schema{
					"foo": {
						Schema: Draft2020_12,
						Title:  "foo",
						Type:   schemaTypeObject,
						Properties: map[string]*jsonschema.Schema{
							"bar": {
								Schema: Draft2020_12,
								Title:  "bar",
								Type:   schemaTypeString,
							},
							"baz": {
								Schema: Draft2020_12,
								Title:  "baz",
								Type:   schemaTypeNumber,
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, err := os.CreateTemp(t.TempDir(), "main.jsonnet")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := f.Write(tc.jsonnet); err != nil {
				t.Fatal(err)
			}
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
			pkg, err := docsonnet.Load(f.Name(), &docsonnet.Opts{})
			if err != nil {
				t.Fatal(err)
			}
			schema := Convert(pkg)
			if err := compareSchemas(tc.schema, schema); err != nil {
				t.Fatal(err)
			}
		})
	}
}

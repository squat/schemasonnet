package schema

import (
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/jsonnet-libs/docsonnet/pkg/docsonnet"
)

const (
	Draft2020_12      = "https://json-schema.org/draft/2020-12/schema"
	titleArguments    = "Arguments"
	propertyArguments = "Arguments"
	schemaTypeObject  = "object"
	schemaTypeArray   = "array"
	schemaTypeString  = "string"
	schemaTypeNumber  = "number"
	schemaTypeBool    = "boolean"
)

func Convert(pkg *docsonnet.Package) *jsonschema.Schema {
	return fieldsToSchema(pkg.Name, pkg.Help, pkg.API)
}

func fieldsToSchema(name, help string, fields docsonnet.Fields) *jsonschema.Schema {
	s := &jsonschema.Schema{
		Schema:      Draft2020_12,
		Title:       name,
		Description: help,
		Type:        schemaTypeObject,
		Properties:  make(map[string]*jsonschema.Schema),
	}
	for _, field := range fields {
		switch {
		case field.Function != nil:
			s.Properties[field.Function.Name] = &jsonschema.Schema{
				Schema:      Draft2020_12,
				Title:       field.Function.Name,
				Description: field.Function.Help,
				Type:        schemaTypeObject,
				Properties: map[string]*jsonschema.Schema{
					propertyArguments: argumentsToSchema(field.Function.Args),
				},
			}
		case field.Object != nil:
			s.Properties[field.Object.Name] = fieldsToSchema(field.Object.Name, field.Object.Help, field.Object.Fields)

		case field.Value != nil:
			s.Properties[field.Value.Name] = valueToSchema(field.Value.Name, field.Value.Help, field.Value.Type)
		}

	}

	return s
}

func typeToType(t docsonnet.Type) string {
	switch t {
	case docsonnet.TypeString:
		return schemaTypeString
	case docsonnet.TypeNumber:
		return schemaTypeNumber
	case docsonnet.TypeBool:
		return schemaTypeBool
	case docsonnet.TypeObject:
		return schemaTypeObject
	case docsonnet.TypeArray:
		return schemaTypeArray
	case docsonnet.TypeFunc:
		return schemaTypeObject
	case docsonnet.TypeAny:
		fallthrough
	default:
		return ""
	}
}

func argumentsToSchema(arguments []docsonnet.Argument) *jsonschema.Schema {
	s := &jsonschema.Schema{
		Schema: Draft2020_12,
		Title:  titleArguments,
		Type:   schemaTypeArray,
	}

	for _, argument := range arguments {
		s.PrefixItems = append(s.PrefixItems, valueToSchema(argument.Name, "", argument.Type))
	}

	return s
}

func valueToSchema(name, help string, t docsonnet.Type) *jsonschema.Schema {
	s := &jsonschema.Schema{
		Schema:      Draft2020_12,
		Title:       name,
		Description: help,
		Type:        typeToType(t),
	}
	if t == docsonnet.TypeFunc {
		s.Properties = map[string]*jsonschema.Schema{
			// Pass nil because docsonnet does not allow describing
			// the arguments of functions that are themselves arguments.
			propertyArguments: argumentsToSchema(nil),
		}
	}
	return s
}

package openapi2_mcp

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"maps"
	"net/url"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	// "github.com/mark3labs/mcp-go/mcp"
)

func coalesce[T comparable](vals ...T) T {
	var zero T
	for _, val := range vals {
		if val != zero {
			return val
		}
	}
	return zero
}

var pathParamRE = regexp.MustCompile(`\{[^}]+\}`)

func resolvePath(
	path string,
	params []*openapi2.Parameter,
	args map[string]any) string {
	vals := make(map[string]any, 0)
	for _, param := range params {
		if param.In != "path" {
			continue
		}

		vals[param.Name] = args[param.Name]
	}

	result := pathParamRE.ReplaceAllStringFunc(path, func(m string) string {
		name := m[1 : len(m)-1]
		if val, ok := vals[name]; ok {
			return val.(string)
		}
		return m
	})
	return result
}

func resolveQuery(
	params []*openapi2.Parameter,
	args map[string]any,
) string {
	qv := url.Values{}

	for _, param := range params {
		if param.In != "query" {
			continue
		}

		if val, ok := args[param.Name]; ok {
			qv.Set(param.Name, val.(string))
		}
	}

	return qv.Encode()
}

func resolveBody(params []*openapi2.Parameter, args map[string]any) (io.Reader, error) {
	m := make(map[string]any, 0)
	for _, param := range params {
		if param.In != "body" {
			continue
		}

		values, ok := args["body"].(map[string]any)
		if !ok {
			continue
		}

		maps.Copy(m, values)
	}

	if len(m) == 0 {
		return nil, nil
	}
	payload, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(payload), nil
}

// buildFullURL constructs the full URL for an API request based on the OpenAPI specification, path parameters, and query parameters.
func (o *Spec) buildFullURL(path string, op *openapi2.Operation, method string, args map[string]any) string {
	p := resolvePath(path, op.Parameters, args)
	q := resolveQuery(op.Parameters, args)

	fullURL := o.scheme + "://" + o.doc.Host + o.doc.BasePath + p
	if q != "" {
		fullURL += "?" + q
	}

	if o.urlBuilder != nil {
		fullURL = o.urlBuilder(op, method, o.scheme, o.doc.Host, o.doc.BasePath, p, q)
	}

	return fullURL
}

func schemaToJsonSchema(doc *openapi2.T, schema *openapi2.Schema) map[string]any {
	m := make(map[string]any, 0)

	if !schema.Type.IsEmpty() {
		m["type"] = schema.Type
	}

	if schema.Description != "" {
		m["description"] = schema.Description
	}

	if len(schema.Required) > 0 {
		m["required"] = schema.Required
	}

	if items := schema.Items; items != nil {
		sc := schemaToJsonSchema(doc, items.Value)
		if items.Ref != "" {
			sc = schemaToJsonSchema(doc, getSchemaFromRef(doc, items.Ref))
		}
		m["items"] = sc
	}

	if schema.Enum != nil {
		m["enum"] = schema.Enum
	}

	if schema.Format != "" {
		m["format"] = schema.Format
	}

	if schema.Pattern != "" {
		m["pattern"] = schema.Pattern
	}

	if schema.Default != nil {
		m["default"] = schema.Default
	}

	if len(schema.Properties) > 0 {
		props := make(map[string]any)
		for propName, propSchema := range schema.Properties {
			if val := propSchema.Value; val != nil && val.Items != nil && val.Items.Ref != "" {
				props[propName] = schemaToJsonSchema(doc, getSchemaFromRef(doc, val.Items.Ref))
			} else if propSchema.Ref != "" {
				props[propName] = schemaToJsonSchema(doc, getSchemaFromRef(doc, propSchema.Ref))
			} else {
				props[propName] = schemaToJsonSchema(doc, propSchema.Value)
			}

		}
		m["properties"] = props
	}
	return m
}

func getSchemaFromRef(doc *openapi2.T, ref string) *openapi2.Schema {
	s := strings.TrimPrefix(ref, "#/definitions/")
	r, ok := doc.Definitions[s]
	if !ok {
		log.Printf("Definition not found for ref: %s", ref)
		return nil
	}
	return r.Value
}

package openapi3

import (
	"bytes"
	"encoding/json"
	"io"
	"maps"
	"net/url"
	"regexp"

	"github.com/getkin/kin-openapi/openapi3"
)

var pathParamRE = regexp.MustCompile(`\{[^}]+\}`)

func coalesce[T comparable](vals ...T) T {
	var zero T
	for _, val := range vals {
		if val != zero {
			return val
		}
	}
	return zero
}

func resolvePath(
	path string,
	params openapi3.Parameters,
	args map[string]any) string {
	vals := make(map[string]any, 0)
	for _, param := range params {
		if param.Value.In != "path" {
			continue
		}

		vals[param.Value.Name] = args[param.Value.Name]
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
	params openapi3.Parameters,
	args map[string]any,
) string {
	qv := url.Values{}

	for _, param := range params {
		if param.Value.In != "query" {
			continue
		}

		if val, ok := args[param.Value.Name]; ok {
			qv.Set(param.Value.Name, val.(string))
		}
	}

	return qv.Encode()
}

func resolveBody(params openapi3.Parameters, args map[string]any) (io.Reader, error) {
	m := make(map[string]any, 0)
	for _, param := range params {
		if param.Value.In != "body" {
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
func (o *Spec) buildFullURL(path string, op *openapi3.Operation, method string, args map[string]any) string {
	p := resolvePath(path, op.Parameters, args)
	q := resolveQuery(op.Parameters, args)

	fullURL := o.server + p
	if q != "" {
		fullURL += "?" + q
	}

	if o.urlBuilder != nil {
		fullURL = o.urlBuilder(op, method, p, q)
	}

	return fullURL
}

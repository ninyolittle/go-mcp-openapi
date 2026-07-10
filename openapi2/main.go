package openapi2_mcp

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/mark3labs/mcp-go/server"
)

// Spec represents an OpenAPI 2.0 specification instance with customizable options.
type Spec struct {
	doc *openapi2.T

	urlBuilder func(operation *openapi2.Operation, method string, scheme, host, basePath, path, query string) string
	headers    func(ctx context.Context) (map[string]string, error)
	scheme     string
}

// New creates a new OpenAPI 2.0 specification instance with the provided options.
func New(opts ...option) *Spec {
	p := &Spec{
		scheme: "https",
	}

	for _, opt := range opts {
		opt(p)
	}

	if len(p.doc.Schemes) > 0 {
		p.scheme = p.doc.Schemes[0]
	}
	return p
}

type option func(*Spec)

// WithHeaders allows you to set custom headers for the OpenApiTools instance.
// This can be useful for adding authentication tokens or other necessary headers when making requests to the API.
// Example usage:
//
//	o := NewOpenApiSpec(
//		LoadFromURL("https://example.com/api-docs.json"),
//		WithHeaders(func(ctx context.Context) map[string]string {
//	    	return map[string]string{
//	    		"Authorization": "Bearer <token>",
//	    		"Content-Type":  "application/json",
//	    	}
//		}),
//	)
func WithHeaders(provider func(ctx context.Context) (map[string]string, error)) option {
	return func(o *Spec) {
		o.headers = provider
	}
}

// WithUrlBuilder allows you to set a custom URL builder function for the OpenApiTools instance.
// This can be useful for customizing how URLs are constructed when making requests to the API.
// Example usage:
//
//	o := NewOpenApiSpec(
//		LoadFromURL("https://example.com/api-docs.json"),
//		WithUrlBuilder(func(scheme, host, basePath, path, query string) string {
//			base := scheme + "://" + host + basePath + "/custom" + path
//			if query != "" {
//				base += "?" + query
//			}
//			return base
//		}),
//
// )
func WithUrlBuilder(urlBuilder func(operation *openapi2.Operation, method string, scheme, host, basePath, path, query string) string) option {
	return func(o *Spec) {
		o.urlBuilder = urlBuilder
	}
}

// FilterByOperationIds filters the OpenAPI specification to only include operations with the specified operation IDs.
func FilterByOperationIds(operationIds []string) option {
	allowed := make(map[string]struct{}, len(operationIds))
	for _, operationId := range operationIds {
		allowed[operationId] = struct{}{}
	}

	return func(o *Spec) {
		for _, path := range o.doc.Paths {
			for method, operation := range path.Operations() {
				if _, ok := allowed[operation.OperationID]; !ok {
					path.SetOperation(method, nil)
				}
			}

		}
	}
}

// FilterByPaths filters the OpenAPI specification to only include paths that match the specified paths.
func FilterByPaths(paths []string) option {
	allowed := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		allowed[path] = struct{}{}
	}

	return func(o *Spec) {
		for pathStr := range o.doc.Paths {
			if _, ok := allowed[pathStr]; !ok {
				delete(o.doc.Paths, pathStr)
			}
		}
	}
}

// FilterByTags filters the OpenAPI specification to only include operations that have at least one of the specified tags.
func FilterByTags(tags []string) option {
	allowed := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		allowed[tag] = struct{}{}
	}

	hasAllowedTag := func(tags []string) bool {
		for _, tag := range tags {
			if _, ok := allowed[tag]; ok {
				return true
			}
		}
		return false
	}

	return func(o *Spec) {
		for _, path := range o.doc.Paths {
			for method, operation := range path.Operations() {
				if !hasAllowedTag(operation.Tags) {
					path.SetOperation(method, nil)
				}
			}
		}
	}
}

// LoadFromUrl loads an OpenAPI 2.0 specification from a URL.
func LoadFromUrl(url string) option {
	return func(d *Spec) {
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		if err = json.NewDecoder(resp.Body).Decode(&d.doc); err != nil {
			log.Fatal(err)
		}
	}
}

// LoadFromFile loads an OpenAPI 2.0 specification from a JSON file.
func LoadFromFile(path string) option {
	return func(d *Spec) {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}

		if err = json.Unmarshal(data, &d.doc); err != nil {
			log.Fatal(err)
		}
	}
}

// RegisterToMCPGoServer registers the OpenAPI specification as tools in the MCP-Go server.
// It builds the tools based on the OpenAPI operations and adds them to the server.
func (o *Spec) RegisterToMCPGoServer(server *server.MCPServer) error {
	mgo := &mcpGoTools{o}
	tools, err := mgo.buildTools()
	if err != nil {
		return err
	}

	log.Printf("%d tools registered in %s", len(tools), o.doc.Info.Title)
	server.AddTools(tools...)
	return nil
}

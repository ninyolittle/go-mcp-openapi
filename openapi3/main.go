package openapi3

import (
	"context"
	"log"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
)

type Spec struct {
	doc *openapi3.T

	server     string
	headers    func(ctx context.Context) (map[string]string, error)
	urlBuilder func(p *openapi3.Operation, method, path, query string) string
}

type option func(*Spec)

func FilterByTags(tags []string) option {
	allowed := make(map[string]struct{})
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
		for _, path := range o.doc.Paths.Map() {
			for method, operation := range path.Operations() {
				if !hasAllowedTag(operation.Tags) {
					path.SetOperation(method, nil)
				}
			}
		}
	}

}

func WithHeaders(provider func(ctx context.Context) (map[string]string, error)) option {
	return func(o *Spec) {
		o.headers = provider
	}
}

func LoadWithUrl(u string) option {
	return func(o *Spec) {
		loader := openapi3.NewLoader()
		parsedURL, err := url.Parse(u)
		if err != nil {
			log.Fatal(err)
		}
		doc, err := loader.LoadFromURI(parsedURL)
		if err != nil {
			log.Fatal(err)
		}
		o.doc = doc
	}
}

func LoadFromFile(path string) option {
	return func(o *Spec) {
		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromFile(path)
		if err != nil {
			log.Fatal(err)
		}
		o.doc = doc
	}
}

func WithURLBuilder(builder func(p *openapi3.Operation, method, path, query string) string) option {
	return func(o *Spec) {
		o.urlBuilder = builder
	}
}

// WithServer allows you to specify a function that will select the
// server URL from the list of servers defined in the OpenAPI spec.
// If no servers are defined, it will default to an empty string.
func WithServer(servers func([]string) string) option {
	return func(o *Spec) {
		var s []string
		for _, server := range o.doc.Servers {
			s = append(s, server.URL)
		}
		o.server = servers(s)
	}
}

func New(opts ...option) *Spec {
	p := &Spec{}

	if len(p.doc.Servers) > 0 {
		p.server = p.doc.Servers[0].URL
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

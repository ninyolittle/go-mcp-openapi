package openapi2_mcp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type mcpGoTools struct {
	*Spec
}

func (o *mcpGoTools) buildHandler(
	path string,
	operation *openapi2.Operation,
	method string,
	ctx context.Context,
	req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	fullURL := o.buildFullURL(path, operation, method, args)

	payload, err := resolveBody(operation.Parameters, args)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, fullURL, payload)
	if err != nil {
		return nil, err
	}
	headers, err := o.headers(ctx)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(body)), nil
}

// preprocessToolOpts processes the parameters of an OpenAPI operation and
// generates corresponding MCP tool options based on their types and properties.
func (o *mcpGoTools) preprocessToolOpts(
	toolOpts *[]mcp.ToolOption, operation *openapi2.Operation) {
	for _, param := range operation.Parameters {
		propOpts := []mcp.PropertyOption{
			mcp.Description(param.Description),
		}

		if param.Required {
			propOpts = append(propOpts, mcp.Required())
		}

		switch {
		case param.Type.Is("string"):
			*toolOpts = append(*toolOpts, mcp.WithString(param.Name, propOpts...))
		case param.Type.Is("number"):
			*toolOpts = append(*toolOpts, mcp.WithNumber(param.Name, propOpts...))
		case param.Type.Is("integer"):
			*toolOpts = append(*toolOpts, mcp.WithInteger(param.Name, propOpts...))
		case param.Type.Is("boolean"):
			*toolOpts = append(*toolOpts, mcp.WithBoolean(param.Name, propOpts...))
		case param.Type.Is("array"):
			*toolOpts = append(*toolOpts, mcp.WithArray(param.Name, propOpts...))
		default:
			propOpts = append(propOpts, mcp.Properties(schemaToJsonSchema(o.doc, getSchemaFromRef(o.doc, param.Schema.Ref))["properties"].(map[string]any)))
			*toolOpts = append(*toolOpts, mcp.WithObject(param.Name, propOpts...))
		}
	}
}

var nonAlnum = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func (o *mcpGoTools) buildTools() ([]server.ServerTool, error) {
	tools := make([]server.ServerTool, 0)

	for path, pathItem := range o.doc.Paths {
		for method, operation := range pathItem.Operations() {
			if operation == nil {
				continue
			}

			readOnly := method == "GET"
			destructive := method == "DELETE"

			toolOpts := []mcp.ToolOption{
				mcp.WithDescription(coalesce(operation.Summary, operation.Description)),
				mcp.WithReadOnlyHintAnnotation(readOnly),
				mcp.WithDestructiveHintAnnotation(destructive),
			}

			o.preprocessToolOpts(&toolOpts, operation)

			toolName := operation.OperationID
			if toolName == "" {
				a := fmt.Sprintf("%s_%s", strings.ToLower(method), strings.ReplaceAll(strings.Trim(path, "/"), "/", "_"))
				toolName = strings.Trim(nonAlnum.ReplaceAllString(a, "_"), "_")
			}

			tools = append(tools, server.ServerTool{
				Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
					return o.buildHandler(
						path, operation, method, ctx, request)
				},
				Tool: mcp.NewTool(toolName, toolOpts...),
			})

		}
	}

	return tools, nil
}

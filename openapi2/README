# openapi2_mcp

openapi2_mcp is a Go package that provides a set of tools for working with OpenAPI 2.0 specifications. It allows you to load, filter, and modify OpenAPI specifications, as well as generate server code and client code based on the specification.

## Create a new OpenAPI specification

```go
import (
    omcpgo "github.com/ninyolittle/go-mcp-openapi/openapi2_mcp"
)

func main () {
    spec, err := omcpgo.NewOpenApiSpec()
}
```

## Options

### Load options

This option specifies how to load the OpenAPI specification. You must provide either a file path or a URL; otherwise, an error is returned. Currently, only OpenAPI specifications in JSON format are supported. The following load options are available:

#### Load from file

```go
spec, err := omcpgo.NewOpenApiSpec(
    omcpgo.LoadFromFile("openapi.json"),
)
```

#### Load from URL

```go
spec, err := omcpgo.NewOpenApiSpec(
    omcpgo.LoadFromURL("https://example.com/openapi.json"),
)
```

### Filter options

Filters is used if you dont want to expose all of your API endpoints.

#### By operation ID

This is the most common filter, and it allows you to specify which operations should be included in the OpenAPI specification based on their operation ID.

Suppose you have the following operations in your OpenAPI specification:

```json
...
"paths": {
"/user": {
    "get": {
    "operationId": "getUser"
    },
    "post": {
    "operationId": "createUser"
    }
}
}
```

In go, you can filter the OpenAPI specification to include only the `getUser` and `createUser` operations by using the `FilterByOperationID` option:

```go
spec, err := omcpgo.NewOpenApiSpec(
    omcpgo.LoadFromFile("openapi.json"),
    omcpgo.FilterByOperationID([]string{"getUser", "createUser"}),
)
```

#### By HTTP method

This filter allows you to specify which operations should be included in the OpenAPI specification based on their HTTP method. For example, if you want to include only the `GET` and `POST` operations, you can use the `FilterByHTTPMethod` option:

NOTE: The HTTP method filter is case-insensitive, so you can use either uppercase or lowercase letters.

```go
spec, err := omcpgo.NewOpenApiSpec(
    omcpgo.LoadFromFile("openapi.json"),
    omcpgo.FilterByHTTPMethod([]string{"GET", "POST"}),
)
```

#### By Paths

This filter allows you to specify which operations should be included in the OpenAPI specification based on their path. For example, if you want to include only the operations with the `/user` and `/admin` paths, you can use the `FilterByPaths` option:

```go
spec, err := omcpgo.NewOpenApiSpec(
    omcpgo.LoadFromFile("openapi.json"),
    omcpgo.FilterByPaths([]string{"/user", "/admin"}),
)
```

#### By Tags

This filter allows you to specify which operations should be included in the OpenAPI specification based on their tags. For example, if you want to include only the operations with the `user` and `admin` tags, you can use the `FilterByTags` option:

NOTE: The tags filter is in the operation level, so you can use it to filter operations that have specific tags.

Example:

```json
...
"paths": {
    "/access_token": {
        "post": {
            "consumes": [
                "application/x-www-form-urlencoded"
            ],
            "produces": [
                "application/json"
            ],
            "tags": [
                "user"
            ],
        }
    }
}
```

```go
spec, err := omcpgo.NewOpenApiSpec(
    omcpgo.LoadFromFile("openapi.json"),
    omcpgo.FilterByTags([]string{"user", "admin"}),
)
```

### Modifier options

Modifiers are used to modify the OpenAPI specification after it has been loaded. This can be useful if you want to add additional information to the documentation. These are the available modifiers:

#### With Host

This modifier allows you to specify the host for the OpenAPI specification. This is useful as this value will be used as host in the URL when calling the API. For example, if you want to set the host to `api.example.com`, you can use the `WithHost` option:

```go
spec, err := omcpgo.NewOpenApiSpec(
    omcpgo.LoadFromFile("openapi.json"),
    omcpgo.WithHost("api.example.com"),
)
```

#### With URL Builder

This modifier lets you specify a custom URL builder for the OpenAPI specification. The custom URL builder is used to construct request URLs when calling the API, replacing the default URL generation logic. This is useful when the default behavior does not match your API’s URL structure or when additional customization is required. To use a custom URL builder, pass the `WithURLBuilder` option:

```go
spec, err := omcpgo.NewOpenApiSpec(
    omcpgo.LoadFromFile("openapi.json"),
    omcpgo.WithURLBuilder(func(op *openapi3.Operation, method, scheme, host, basePath string, path string, query string) string {
        // Custom URL building logic here
        base := scheme + "://" + host + basePath + path
            if query != "" {
                base += "?" + query
            }
        return base
    }),
)
```

#### With Headers

This modifier allows you to specify custom headers for the OpenAPI specification. This is useful when you want to add additional information to the documentation, such as authentication tokens or other metadata. These values are used every time a request is made. To use custom headers, pass the `WithHeaders` option:

```go
spec, err := omcpgo.NewOpenApiSpec(
    omcpgo.LoadFromFile("openapi.json"),
    omcpgo.WithHeaders(func (ctx context.Context) map[string]string {
        // Custom header logic here
        return map[string]string{
            "Authorization": "Bearer <token>",
            "X-Custom-Header": "CustomValue",
        }
    }),
)
```

The `ctx` parameter allows you to access the request context, which can be useful for retrieving information about the current request or user session when generating headers.

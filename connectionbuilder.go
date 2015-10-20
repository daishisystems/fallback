package fallback

import (
	"encoding/json"
)

type connectionBuilder interface {
	createConnection()
	addHTTPPOSTBody() error
	addHTTPHeaders()
	addPayloads()
	addFallback()
}

// ConnectionBuilder represents a Builder-pattern based means of constructing
// Connection instances.
type ConnectionBuilder struct {
	name, method, path        string
	returnsJSON               bool
	body, output, customError interface{}
	headers                   map[string]string
	fallback                  Connecter

	Connection *Connection
}

// NewConnectionBuilder returns a new ConnectionBuilder instance based on the
// specified metadata pertaining to ConnectionBuilder.
func NewConnectionBuilder(name, method, path string, returnsJSON bool,
	body interface{}, headers map[string]string, output,
	customError interface{}, fallback Connecter) *ConnectionBuilder {

	return &ConnectionBuilder{
		name:        name,
		method:      method,
		path:        path,
		returnsJSON: returnsJSON,
		body:        body,
		output:      output,
		customError: customError,
		headers:     headers,
		fallback:    fallback,
	}
}

func (builder *ConnectionBuilder) createConnection() {

	builder.Connection = &Connection{
		Name:   builder.name,
		Method: builder.method,
		Path:   builder.path,
	}

	if builder.returnsJSON {
		builder.Connection.Headers = map[string]string{
			"Content-Type": "application/json",
		}
	}
}

func (builder *ConnectionBuilder) addHTTPPOSTBody() error {

	body, err := json.Marshal(builder.body)
	if err != nil {
		return err
	}

	builder.Connection.Body = body
	return nil
}

func (builder *ConnectionBuilder) addHTTPHeaders() {

	if builder.Connection.Headers != nil {
		for header, value := range builder.headers {
			builder.Connection.Headers[header] = value
		}
	} else {
		builder.Connection.Headers = builder.headers
	}
}

func (builder *ConnectionBuilder) addPayloads() {

	builder.Connection.Output = builder.output
	builder.Connection.CustomError = builder.customError
}

func (builder *ConnectionBuilder) addFallback() {

	builder.Connection.Fallback = builder.fallback
}

// ConnectionManager represents the Director structure that applies to
// ConnectionBuilder when creating Connection instances.
type ConnectionManager struct{}

// CreateConnection represents the constructor method that applies to
// ConnectionBuilder when creating Connection instances. It invokes each
// relevant construction method on builder in order to yield a complete
// Connection instance.
func (manager *ConnectionManager) CreateConnection(builder *ConnectionBuilder) {

	builder.createConnection()

	if builder.body != nil {
		builder.addHTTPPOSTBody()
	}

	builder.addHTTPHeaders()
	builder.addPayloads()

	if builder.fallback != nil {
		builder.addFallback()
	}
}

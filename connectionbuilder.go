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

type ConnectionBuilder struct {
	name, method, path        string
	returnsJSON               bool
	body, output, customError interface{}
	headers                   map[string]string
	fallback                  Connecter

	Connection *Connection
}

func NewConnectionBuilder(name, method, path string, returnsJSON bool, body interface{},
	headers map[string]string, output interface{}, customError interface{},
	fallback Connecter) *ConnectionBuilder {

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

	marshaled, err := json.Marshal(builder.body)
	if err != nil {
		return err
	}

	builder.Connection.Body = marshaled
	return nil
}

func (builder *ConnectionBuilder) addHTTPHeaders() {

	builder.Connection.Headers = builder.headers
}

func (builder *ConnectionBuilder) addPayloads() {

	builder.Connection.Output = builder.output
	builder.Connection.CustomError = builder.customError
}

func (builder *ConnectionBuilder) addFallback() {

	builder.Connection.Fallback = builder.fallback
}

type ConnectionManager struct{}

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

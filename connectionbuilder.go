package fallback

import (
	"encoding/json"
)

type connectionBuilder interface {
	createConnection(name, method, path string, returnsJSON bool) *Connection
	addHTTPPOSTBody(body interface{}) error
	addHTTPHeaders(headers map[string]string)
	addPayloads(output interface{}, customError interface{})
	addFallback(fallback Connecter)
}

type ConnectionBuilder struct {
	Connection *Connection
}

func (builder *ConnectionBuilder) createConnection(name, method,
	path string, returnsJSON bool) {

	builder.Connection = &Connection{
		Name:   name,
		Method: method,
		Path:   path,
	}

	if returnsJSON {
		builder.Connection.Headers = map[string]string{
			"Content-Type": "application/json",
		}
	}
}

func (builder *ConnectionBuilder) addHTTPPOSTBody(body interface{}) error {

	marshaled, err := json.Marshal(body)
	if err != nil {
		return err
	}

	builder.Connection.Body = marshaled
	return nil
}

func (builder *ConnectionBuilder) addHTTPHeaders(headers map[string]string) {

	builder.Connection.Headers = headers
}

func (builder *ConnectionBuilder) addPayloads(output interface{},
	customError interface{}) {

	builder.Connection.Output = output
	builder.Connection.CustomError = customError
}

func (builder *ConnectionBuilder) addFallback(fallback Connecter) {

	builder.Connection.Fallback = fallback
}

type ConnectionManager struct {
	builder *ConnectionBuilder
}

func (manager *ConnectionManager) CreateConnection(
	name, method, path string, returnsJSON bool, body interface{},
	headers map[string]string, output interface{}, customError interface{},
	fallback Connecter) {

	manager.builder.createConnection(name, method, path, returnsJSON)

	if body != nil {
		manager.builder.addHTTPPOSTBody(body)
	}

	manager.builder.addHTTPHeaders(headers)
	manager.builder.addPayloads(output, customError)

	if fallback != nil {
		manager.builder.addFallback(fallback)
	}
}

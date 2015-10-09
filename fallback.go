// Package fallback provides an enhanced degree of redundancy to HTTP requests
// by introducing a Chain of Responsibility, consisting of a series of fallback
// HTTP requests, to augment an initial HTTP request. Should the initial HTTP
// request fail, the next fallback HTTP request in the chain will execute. Any
// number of fallback HTTP requests can be chained sequentially. Redundancy is
// achieved by executing each fallback HTTP request in a recursive manner until
// one of the requests succeeds, or all requests fail.
package fallback

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

// Connecter represents the Handler abstraction in the Chain of Responsibility
// that consists of a series of fallback HTTP requests. It provides a contract
// which defines the prerequisites that implementations must adhere to in order
// to take part in the chain.
type Connecter interface {
	GetName() string

	CreateHTTPRequest(method, path string, body []byte,
		headers map[string]string) (*http.Request, error)

	ExecuteHTTPRequest(method, path string, body []byte,
		headers map[string]string) (int, error)
}

// Connection represents a Concrete Handler implementation in the Chain of
// Responsibility that consists of a series of fallback HTTP requests.
// Consuming clients can utilise this class directly, as per the examples
// provided, or provide custom implementations derived from Connecter.
type Connection struct {
	Name   string      // The name used to describe the Connection.
	Host   string      // The Host URI segment excluding other segments such as query string.
	Output interface{} // A custom struct that represents a deserialised object returned as result
	// of a successful HTTP request.
	CustomError interface{}
	Fallback    Connecter
}

// NewConnection returns a new Connection instance based on the supplied
// metadata pertaining to Connection.
func NewConnection(name, host string, output interface{},
	customError interface{}, fallback Connecter) *Connection {

	return &Connection{name, host, output, customError, fallback}
}

// GetName returns the Connection name.
func (connection Connection) GetName() string {

	return connection.Name
}

func (connection Connection) CreateHTTPRequest(method, path string,
	body []byte, headers map[string]string) (*http.Request, error) {

	var request *http.Request
	var err error

	if body == nil {
		request, err = http.NewRequest(method, connection.Host+"/"+path, nil)
	} else {
		request, err = http.NewRequest(method, connection.Host+"/"+path, bytes.NewBuffer(body))
	}

	if err != nil {
		return nil, err
	}

	for key := range headers {
		request.Header.Add(key, headers[key])
	}
	return request, nil
}

func (connection Connection) ExecuteHTTPRequest(method, path string,
	body []byte, headers map[string]string) (int, error) {

	client := &http.Client{}

	request, err := connection.CreateHTTPRequest(method, path, body, headers)
	if err != nil {
		if connection.Fallback != nil {
			statusCode, err :=
				connection.Fallback.ExecuteHTTPRequest(method, path, body, headers)
			if statusCode < 200 || statusCode > 299 {
				return statusCode, err
			}
			return statusCode, nil
		}
		return 400, err // This error will occur if the URI is malformed or otherwise invalid.
	}

	resp, err := client.Do(request)
	if err != nil {
		if connection.Fallback != nil {
			statusCode, err :=
				connection.Fallback.ExecuteHTTPRequest(method, path, body, headers)
			if statusCode < 200 || statusCode > 299 {
				return statusCode, err
			}
			return statusCode, nil
		}
		return 503, err // This error will occur if the URI is unreachable.
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		if connection.Fallback != nil {
			statusCode, err :=
				connection.Fallback.ExecuteHTTPRequest(method, path, body, headers)
			if statusCode < 200 || statusCode > 299 {
				return statusCode, err
			}
			return statusCode, nil
		}
		return 404, nil
	} else if resp.StatusCode < 200 || resp.StatusCode > 299 {
		if connection.Fallback != nil {
			statusCode, err :=
				connection.Fallback.ExecuteHTTPRequest(method, path, body, headers)
			if statusCode < 200 || statusCode > 299 {
				return statusCode, err
			}
			return statusCode, nil
		}

		dec := json.NewDecoder(resp.Body)

		err := dec.Decode(connection.CustomError)
		if err != nil {
			return resp.StatusCode, errors.New("Unable to parse custom error.")
		}
		return resp.StatusCode, nil
	} else {
		dec := json.NewDecoder(resp.Body)

		err := dec.Decode(connection.Output)
		if err != nil {
			return resp.StatusCode, errors.New("Unable to parse custom error.")
		}
		return resp.StatusCode, nil
	}
}

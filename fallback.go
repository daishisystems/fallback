// Package fallback provides an enhanced degree of redundancy to HTTP requests
// by introducing a Chain of Responsibility, consisting of a series of fallback
// HTTP requests, to augment an initial HTTP request. Should the initial HTTP
// request fail, the next fallback HTTP request in the chain will execute.
//
// Any number of fallback HTTP requests can be chained sequentially. Redundancy
// is achieved by executing each fallback HTTP request in a recursive manner
// until one of the requests succeeds, or all requests fail.
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
// Consuming clients can utilise this class directly, or provide custom
// implementations derived from Connecter.
//
// Name: The name used to describe the Connection.
//
// Host: he Host URI segment excluding other segments such as query string.
//
// Output: A custom struct that represents a deserialised object returned as
// result of a successful HTTP request.
//
// CustomError: A custom struct that represents a deserialised object returned
// as result of an unsuccessful HTTP request.
//
// Fallback: The next link in the Chain of Responsibility. Fallback represents
// an underlying HTTP request that will be invoked in the event failure during
// execution of this HTTP request.
type Connection struct {
	Name        string
	Host        string
	Output      interface{}
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

// CreateHTTPRequest instantiates a http.Request. method refers to the HTTP
// method; e.g., POST, GET, etc.
//
// path refers to the latter segment of a URL; e.g., /api/customers/john.
// Notice that neither host-name nor scheme are included.
//
// body encapsulates the POST body, if applicable.
//
// headers refers to any HTTP headers applicable to the HTTP request.
//
// The method returns a pointer to the constructed http.Request, or an error,
// if the URL is invalid.
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

// ExecuteHTTPRequest represents a Chain of Responsibility consisting of a
// series of fallback HTTP requests, to augment an initial HTTP request. Should
// the initial HTTP request fail, the next fallback HTTP request in the chain
// will execute. Any number of fallback HTTP requests can be chained sequentially.
//
// ExecuteHTTPRequest initially attempts to construct a HTTP connection. Should
// this fail, the process flow shifts to any fallback method applied to
// Connection. If no fallback method is specified, the method returns.
//
// Unreachable URIs will yield a HTTP 503 response. Invalid URIs will yield a
// 404 response. IT is assumed these response types will by default yield a
// Content-Type of “text/plain”, and will therefore not contain a HTTP Response
// Body to parse.
//
// It is assumed that failed HTTP requests that yield HTTP status codes other
// than 404 or 503 will contain a HTTP Response Body suitable for parsing, and
// applicable to Connection.CustomError in terms of deserialization. If no
// fallback method is specified, the HTTP Response Body will be deserialised to
// Connection.CustomError, and the method will return the HTTP status code.
// Otherwise, fall-back will occur; the process flow will recursively fall back
// to each underlying fallback Connection mechanism until a successful attempt
// is established, or all attempts fail.
//
// It is assumed that successful HTTP requests will contain a HTTP Response
// Body suitable for parsing and applicable to Connection.Output. The HTTP
// Response Body will be deserialised to Connection.Output, and the method will
// return the HTTP status code.
func (connection Connection) ExecuteHTTPRequest(method, path string,
	body []byte, headers map[string]string) (int, error) {

	client := &http.Client{}

	request, err := connection.CreateHTTPRequest(method, path, body, headers)
	if err != nil {
		if connection.Fallback != nil {
			statusCode, err :=
				connection.Fallback.ExecuteHTTPRequest(method, path, body, headers)

			return statusCode, err
		}
		return 400, err
	}

	resp, err := client.Do(request)
	if err != nil {
		if connection.Fallback != nil {
			statusCode, err :=
				connection.Fallback.ExecuteHTTPRequest(method, path, body, headers)

			return statusCode, err
		}
		return 503, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		if connection.Fallback != nil {
			statusCode, err :=
				connection.Fallback.ExecuteHTTPRequest(method, path, body, headers)

			return statusCode, err
		}
		return 404, nil
	} else if resp.StatusCode < 200 || resp.StatusCode > 299 {
		if connection.Fallback != nil {
			statusCode, err :=
				connection.Fallback.ExecuteHTTPRequest(method, path, body, headers)

			return statusCode, err
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

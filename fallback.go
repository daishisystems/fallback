// Package fallback enhances the durability of your API by automatically
// recovering from connectivity failure. It achieves this by providing an
// enhanced degree of redundancy to HTTP requests, introducing a Chain of
// Responsibility, consisting of a series of fallback HTTP requests designed
// to augment an initial HTTP request. Should the initial HTTP request fail,
// the next fallback HTTP request in the chain will execute.
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
	"strconv"
)

// Connecter represents the Handler abstraction in the Chain of Responsibility
// that consists of a series of fallback HTTP requests. It provides a contract
// which defines the prerequisites that implementations must adhere to in order
// to take part in the chain.
type Connecter interface {
	ExecuteHTTPRequest() (int, error)
}

// Logger represents an abstraction providing custom logging. Clients may apply
// custom Logger implementations that publish events in the event of any
// Fallback component failing to execute a HTTP request.
//
// Logger expects a successful response and therefore does not return an error.
type Logger interface {
	Log(message string)
}

// Connection represents a Concrete Handler implementation in the Chain of
// Responsibility that consists of a series of fallback HTTP requests.
// Consuming clients can utilise this class directly, or provide custom
// implementations derived from Connecter.
//
// Name: The name used to describe the Connection
//
// Method: The HTTP Verb (GET, POST, PUT, etc.)
//
// Path: The HTTP URI
//
// Body: The HTTP request Body
//
// Headers: The HTTP request Headers
//
// Output: A custom struct that represents a deserialised object returned as
// result of a successful HTTP request
//
// CustomError: A custom struct that represents a deserialised object returned
// as result of an unsuccessful HTTP request
//
// Fallback: The next link in the Chain of Responsibility. Fallback represents
// an underlying HTTP request that will be invoked in the event of failure
// during execution of this HTTP request
//
// Logger: Custom Logger implementations that publish events in the event of
// any Fallback component failing to execute a HTTP request.
type Connection struct {
	Name, Method, Path  string
	Body                []byte
	Headers             map[string]string
	Output, CustomError interface{}
	Fallback            Connecter
	Logger              Logger
}

// NewConnection returns a new Connection instance based on the specified
// metadata pertaining to Connection.
func NewConnection(name, method, path string, body []byte,
	headers map[string]string, output, customError interface{},
	fallback Connecter, logger Logger) *Connection {

	return &Connection{
		name,
		method,
		path,
		body,
		headers,
		output,
		customError,
		fallback,
		logger,
	}
}

// CreateHTTPRequest instantiates a http.Request based on connection metadata.
// The method returns a pointer to the constructed http.Request, or an error,
// if the URL is invalid.
func (connection Connection) createHTTPRequest() (*http.Request, error) {

	var request *http.Request
	var err error

	if connection.Body == nil {
		request, err = http.NewRequest(connection.Method, connection.Path, nil)
	} else {
		request, err = http.NewRequest(connection.Method, connection.Path,
			bytes.NewBuffer(connection.Body))
	}

	if err != nil {
		return nil, err
	}

	for key := range connection.Headers {
		request.Header.Add(key, connection.Headers[key])
	}
	return request, nil
}

// ExecuteHTTPRequest represents a Chain of Responsibility consisting of a
// series of fallback HTTP requests that augment an initial HTTP request.
// Should the initial HTTP request fail, the next fallback HTTP request in the
// chain will execute. Any number of fallback HTTP requests can be chained
// sequentially. ExecuteHTTPRequest initially attempts to construct a HTTP
// connection. Should this fail, the process flow shifts to any fallback method
// applied to Connection. If no fallback method is specified, the method returns.
//
// Unreachable URIs will yield a HTTP 503 response. Invalid URIs will yield a
// HTTP 400 response. Neither response will yield a response body.
//
// It is assumed that successful HTTP requests will contain a HTTP Response
// Body suitable for parsing and applicable to Connection.Output. The HTTP
// Response Body will be deserialised to Connection.Output, and the method will
// return the HTTP status code in such cases. If the HTTP Response Body is not
// set, or cannot be deserialised to Connection.Output, an error is returned
// along with the HTTP status code.
//
// It is assumed that failed HTTP requests that yield HTTP status codes other
// than 400 or 503 will contain a HTTP Response Body suitable for parsing, and
// applicable to Connection.CustomError in terms of deserialization. If no
// fallback method is specified, the HTTP Response Body will be deserialised to
// Connection.CustomError, and the method will return the HTTP status code.
// Otherwise, fallback will occur; the process flow will recursively fall back
// to each underlying fallback Connection mechanism until a successful attempt
// is established, or all attempts fail. If the HTTP Response Body is not set,
// or cannot be deserialised to Connection.CustomError, an error is returned
// along with the HTTP status code.
func (connection Connection) ExecuteHTTPRequest() (int, error) {

	client := &http.Client{}

	request, err := connection.createHTTPRequest()
	if err != nil {
		if connection.Logger != nil {
			connection.Logger.Log(connection.Name + " failed: " + err.Error())
		}

		if connection.Fallback != nil {
			statusCode, err :=
				connection.Fallback.ExecuteHTTPRequest()

			return statusCode, err
		}
		return 400, err
	}

	resp, err := client.Do(request)
	if err != nil {
		if connection.Logger != nil {
			connection.Logger.Log(connection.Name + " failed: " + err.Error())
		}

		if connection.Fallback != nil {
			statusCode, err :=
				connection.Fallback.ExecuteHTTPRequest()

			return statusCode, err
		}
		return 503, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		if connection.Logger != nil {
			connection.Logger.Log(
				connection.Name + " returned HTTP Error: " +
					strconv.Itoa(resp.StatusCode))
		}

		if connection.Fallback != nil {
			statusCode, err :=
				connection.Fallback.ExecuteHTTPRequest()

			return statusCode, err
		}

		dec := json.NewDecoder(resp.Body)
		err := dec.Decode(connection.CustomError)
		if err != nil {
			return resp.StatusCode,
				errors.New("Unable to parse HTTP Response body.")
		}

		return resp.StatusCode, nil
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(connection.Output)
	if err != nil {
		return resp.StatusCode, errors.New("Unable to parse HTTP Response body.")
	}

	return resp.StatusCode, nil
}

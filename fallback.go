// Package fallback provides an enhanced degree of redundancy to HTTP requests
// by introducing a chain of responsibility, consisting of a series of fallback
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

type Connecter interface {
	GetName() string

	CreateHTTPRequest(method, path string, body []byte,
		headers map[string]string) (*http.Request, error)

	ExecuteHTTPRequest(method, path string, body []byte,
		headers map[string]string) (int, error)
}

type Connection struct {
	name        string
	host        string
	output      interface{}
	customError interface{}
	fallback    Connecter
}

func NewConection(name, host string, output interface{},
	customError interface{}, fallback Connecter) *Connection {

	return &Connection{name, host, output, customError, fallback}
}

func (connection Connection) GetName() string {

	return connection.name
}

func (connection Connection) CreateHTTPRequest(method, path string,
	body []byte, headers map[string]string) (*http.Request, error) {

	// todo: Check for invalid method names

	var request *http.Request
	var err error

	if body == nil {
		request, err = http.NewRequest(method, connection.host+"/"+path, nil)
	} else {
		request, err = http.NewRequest(method, connection.host+"/"+path, bytes.NewBuffer(body))
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

	// Attempt to construct an initial connection. Should this fail,
	// recursively fall back to each underlying connection mechanism until a
	// successful attempt is established, or all attempts fail.

	request, err := connection.CreateHTTPRequest(method, path, body, headers)
	if err != nil {
		if connection.fallback != nil {
			statusCode, err :=
				connection.fallback.ExecuteHTTPRequest(method, path, body, headers)
			if statusCode < 200 || statusCode > 299 {
				return statusCode, err
			}
			return statusCode, nil
		}
		return 400, err // This error will occur if the URI is malformed or otherwise invalid.
	}

	// Attempt to execute the request. Should this fail, recursively fall back
	// to each underlying connection mechanism until a successful attempt is
	// established, or all attempts fail.

	resp, err := client.Do(request)
	if err != nil {
		if connection.fallback != nil {
			statusCode, err :=
				connection.fallback.ExecuteHTTPRequest(method, path, body, headers)
			if statusCode < 200 || statusCode > 299 {
				return statusCode, err
			}
			return statusCode, nil
		}
		return 503, err // This error will occur if the URI is unreachable.
	}
	defer resp.Body.Close()

	// Handle all potential HTTP responses. At this point, the request has
	// executed and will return a valid HTTP status code. Successful HTTP
	// status codes (200 – 202) will yield the necessary payload from Monex,
	// in JSON format. Assuming valid format, this payload will be parsed to
	// output, and returned. Unsuccessful HTTP status codes will cause
	// fall-back; the process flow will recursively fall back to each
	// underlying connection mechanism until a successful attempt is
	// established, or all attempts fail.

	if resp.StatusCode == 404 {
		if connection.fallback != nil {
			statusCode, err :=
				connection.fallback.ExecuteHTTPRequest(method, path, body, headers)
			if statusCode < 200 || statusCode > 299 {
				return statusCode, err
			}
			return statusCode, nil
		}
		return 404, nil
	} else if resp.StatusCode < 200 || resp.StatusCode > 299 {
		if connection.fallback != nil {
			statusCode, err :=
				connection.fallback.ExecuteHTTPRequest(method, path, body, headers)
			if statusCode < 200 || statusCode > 299 {
				return statusCode, err
			}
			return statusCode, nil
		}

		dec := json.NewDecoder(resp.Body)

		err := dec.Decode(connection.customError)
		if err != nil {
			return resp.StatusCode, errors.New("Unable to parse custom error.")
		}
		return resp.StatusCode, nil
	} else {
		dec := json.NewDecoder(resp.Body)

		err := dec.Decode(connection.output)
		if err != nil {
			return resp.StatusCode, errors.New("Unable to parse custom error.")
		}
		return resp.StatusCode, nil
	}
}

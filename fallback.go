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
		headers map[string]string,
		output interface{},
		customError interface{}) (int, error)
}

type Connection struct {
	name     string
	host     string
	fallback Connecter
}

func NewConection(name, host string, fallback Connecter) *Connection {

	return &Connection{name, host, fallback}
}

func (connection Connection) GetName() string {

	return connection.name
}

func (connection Connection) CreateHTTPRequest(method, path string,
	body []byte, headers map[string]string) (*http.Request, error) {

	req, err :=
		http.NewRequest(method, connection.host+"/"+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	for key := range headers {
		req.Header.Add(key, headers[key])
	}
	return req, nil
}

func (connection Connection) ExecuteHTTPRequest(method, path string,
	body []byte, headers map[string]string,
	output interface{},
	customError interface{}) (int, error) {

	client := &http.Client{}

	// Attempt to construct an initial connection. Should this fail,
	// recursively fall back to each underlying connection mechanism until a
	// successful attempt is established, or all attempts fail.

	request, err := connection.CreateHTTPRequest(method, path, body, headers)
	if err != nil {
		if connection.fallback != nil {
			ok, statusCode, err :=
				connection.fallback.ExecuteHTTPRequest(method, path, body, headers, output, customError)
			if !ok {
				return false, statusCode, err
			}
			return true, statusCode, nil
		}
		return false, 0, err
	}

	// Attempt to execute the request. Should this fail, recursively fall back
	// to each underlying connection mechanism until a successful attempt is
	// established, or all attempts fail.

	resp, err := client.Do(request)
	if err != nil {
		if connection.fallback != nil {
			ok, statusCode, err :=
				connection.fallback.ExecuteHTTPRequest(method, path, body, headers, output, customError)
			if !ok {
				return false, statusCode, err
			}
			return true, statusCode, nil
		}
		return false, 0, err
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
			ok, statusCode, err :=
				connection.fallback.ExecuteHTTPRequest(method, path, body, headers, output, customError)
			if !ok {
				return false, statusCode, err
			}
			return true, statusCode, nil
		}
		return false, 404, nil
	} else if resp.StatusCode < 200 || resp.StatusCode > 202 {
		if connection.fallback != nil {
			ok, statusCode, err :=
				connection.fallback.ExecuteHTTPRequest(method, path, body, headers, output, customError)
			if !ok {
				return false, statusCode, err
			}
			return true, statusCode, nil
		}
		dec := json.NewDecoder(resp.Body)

		err := dec.Decode(customError)
		if err != nil {
			return false, resp.StatusCode, errors.New("Unable to parse custom error.")
		}
		return false, resp.StatusCode, nil
	} else {
		dec := json.NewDecoder(resp.Body)

		err := dec.Decode(output)
		if err != nil {
			return false, resp.StatusCode, errors.New("Unable to parse custom error.")
		}
		return true, resp.StatusCode, nil
	}
}

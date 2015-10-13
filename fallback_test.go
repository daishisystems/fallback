package fallback

import (
	"testing"
)

type BasicResponse struct {
	Text   string
	Detail string
}

type BasicError struct {
	Code    int
	Message string
}

func TestSingleHTTPRequest(t *testing.T) {

	basicResponse := &BasicResponse{}
	basicError := &BasicError{}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	conn := NewConnection("HTTP", "GET",
		"http://demo7227109.mockable.io/get-basic", nil, headers,
		basicResponse, basicError, nil)

	statusCode, err := conn.ExecuteHTTPRequest()

	if err != nil {
		t.Fatal(err)
	}

	if statusCode != 200 {
		t.Fatal("For", "Basic GET",
			"expected", 200,
			"got", statusCode)
	}

	if basicResponse.Text != "OK" || basicResponse.Detail != "Successful HTTP request" {
		t.Error("For", "Basic GET",
			"expected", "OK, Successful HTTP request",
			"got", basicResponse.Text, basicResponse.Detail)
	}
}

func TestSimpleFallback(t *testing.T) {

	basicResponse := &BasicResponse{}
	basicError := &BasicError{}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	working := NewConnection("Working", "GET",
		"http://demo7227109.mockable.io/get-basic", nil, headers,
		basicResponse, basicError, nil)

	failing := NewConnection("Failing", "GET",
		"http://demo7227109.mockable.io/fail-basic", nil, headers,
		basicResponse, basicError, working)

	statusCode, err := failing.ExecuteHTTPRequest()

	if err != nil {
		t.Log("HTTP status code:", statusCode)
		t.Fatal(err)
	}

	if statusCode != 200 {
		t.Fatal("For", "Basic GET",
			"expected", 200,
			"got", statusCode)
	}

	if basicResponse.Text != "OK" || basicResponse.Detail != "Successful HTTP request" {
		t.Error("For", "Basic GET",
			"expected", "OK, Successful HTTP request",
			"got", basicResponse.Text, basicResponse.Detail)
	}
}

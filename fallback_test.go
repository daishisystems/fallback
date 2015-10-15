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

func TestComplexFallback(t *testing.T) {

	basicResponse := &BasicResponse{}
	basicError := &BasicError{}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	workingPOST := NewConnection("Working POST", "POST",
		"http://demo7227109.mockable.io/post-basic", nil, headers,
		basicResponse, basicError, nil)

	failingPOST := NewConnection("Failing POST", "POST",
		"http://demo7227109.mockable.io/fail-basic-post", nil, headers,
		basicResponse, basicError, workingPOST)

	failingGet := NewConnection("Failing GET", "GET",
		"http://demo7227109.mockable.io/fail-basic", nil, headers,
		basicResponse, basicError, failingPOST)

	statusCode, err := failingGet.ExecuteHTTPRequest()

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

func TestFallbackBuilder(t *testing.T) {

	path := "http://demo7227109.mockable.io/get-basic"

	basicResponse := &BasicResponse{}
	basicError := &BasicError{}

	builder := ConnectionBuilder{}
	connectionManager := ConnectionManager{&builder}

	connectionManager.CreateConnection("CONN1", "GET", path, nil, nil,
		basicResponse, basicError, nil)

	statusCode, err := builder.Connection.ExecuteHTTPRequest()

	if err != nil {
		t.Fatal(err)
	}

	if statusCode != 200 {
		t.Fatal("For", "Basic GET",
			"expected", 200,
			"got", statusCode)
	}
}

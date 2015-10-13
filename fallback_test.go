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

	conn := NewConnection("HTTP", "http://demo7227109.mockable.io/get-basic", basicResponse, basicError, nil)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	statusCode, err := conn.ExecuteHTTPRequest("GET", nil, headers)

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

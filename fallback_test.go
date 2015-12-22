package fallback

import "testing"

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
		basicResponse, basicError, nil, nil)

	statusCode, err := conn.ExecuteHTTPRequest()

	if err != nil {
		t.Fatal("Status Code:", statusCode, "Error:", err)
	}

	if statusCode != 200 {
		t.Fatal("For", "Basic GET",
			"expected", 200,
			"got", statusCode)
	}

	if basicResponse.Text != "OK" ||
		basicResponse.Detail != "Successful HTTP request" {
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
		basicResponse, basicError, nil, nil)

	failing := NewConnection("Failing", "GET",
		"http://demo7227109.mockable.io/fail-basic", nil, headers,
		basicResponse, basicError, working, nil)

	statusCode, err := failing.ExecuteHTTPRequest()

	if err != nil {
		t.Fatal(err)
	}

	if statusCode != 200 {
		t.Fatal("For", "Basic GET",
			"expected", 200,
			"got", statusCode)
	}

	if basicResponse.Text != "OK" ||
		basicResponse.Detail != "Successful HTTP request" {
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
		basicResponse, basicError, nil, nil)

	failingPOST := NewConnection("Failing POST", "POST",
		"http://demo7227109.mockable.io/fail-basic-post", nil, headers,
		basicResponse, basicError, workingPOST, nil)

	failingGet := NewConnection("Failing GET", "GET",
		"http://demo7227109.mockable.io/fail-basic", nil, headers,
		basicResponse, basicError, failingPOST, nil)

	statusCode, err := failingGet.ExecuteHTTPRequest()

	if err != nil {
		t.Fatal(err)
	}

	if statusCode != 200 {
		t.Fatal("For", "Basic GET",
			"expected", 200,
			"got", statusCode)
	}

	if basicResponse.Text != "OK" ||
		basicResponse.Detail != "Successful HTTP request" {
		t.Error("For", "Basic GET",
			"expected", "OK, Successful HTTP request",
			"got", basicResponse.Text, basicResponse.Detail)
	}
}

func TestFallbackBuilder(t *testing.T) {

	path := "http://demo7227109.mockable.io/get-basic"

	basicResponse := BasicResponse{}
	basicError := BasicError{}

	builder := NewConnectionBuilder("CONN1", "GET", path, true, nil,
		&basicResponse, &basicError, nil, nil, nil)

	connectionManager := ConnectionManager{}
	connectionManager.CreateConnection(builder)

	statusCode, err := builder.Connection.ExecuteHTTPRequest()

	if err != nil {
		t.Fatal(err)
	}

	if statusCode != 200 {
		t.Fatal("For", "Basic GET",
			"expected", 200,
			"got", statusCode)
	}

	if basicResponse.Text != "OK" ||
		basicResponse.Detail != "Successful HTTP request" {
		t.Error("For", "Basic GET",
			"expected", "OK, Successful HTTP request",
			"got", basicResponse.Text, basicResponse.Detail)
	}
}

func TestComplexFallbackBuilder(t *testing.T) {

	passPath := "http://demo7227109.mockable.io/get-basic"
	failPath2 := "http://demo7227109.mockable.io/fail-basic"
	failPath1 := "http://demo7227109.mockable.io/fail-basic-post"

	basicResponse := BasicResponse{}
	basicError := BasicError{}

	connectionManager := ConnectionManager{}

	passBuilder := NewConnectionBuilder("PASS", "GET", passPath, true, nil,
		&basicResponse, &basicError, nil, nil, nil)
	connectionManager.CreateConnection(passBuilder)

	failBuilder2 := NewConnectionBuilder("FAIL2", "POST", failPath2, true, nil,
		&basicResponse, &basicError, nil, passBuilder.Connection, nil)
	connectionManager.CreateConnection(failBuilder2)

	failBuilder1 := NewConnectionBuilder("FAIL1", "POST", failPath1, true, nil,
		&basicResponse, &basicError, nil, failBuilder2.Connection, nil)
	connectionManager.CreateConnection(failBuilder1)

	statusCode, err := failBuilder1.Connection.ExecuteHTTPRequest()

	if err != nil {
		t.Fatal("Status Code:", statusCode, "Error:", err)
	}

	if statusCode != 200 {
		t.Fatal("For", "Basic GET",
			"expected", 200,
			"got", statusCode)
	}

	if basicResponse.Text != "OK" ||
		basicResponse.Detail != "Successful HTTP request" {
		t.Error("For", "Basic GET",
			"expected", "OK, Successful HTTP request",
			"got", basicResponse.Text, basicResponse.Detail)
	}
}

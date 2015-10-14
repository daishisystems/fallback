<a href="http://insidethecpu.com">![Image of insidethecpu](https://dl.dropboxusercontent.com/u/26042707/Daishi%20Systems%20Icon%20with%20Text%20%28really%20tiny%20with%20photo%29.png)</a>
# Go Fallback Package
[![Join the chat at https://gitter.im/daishisystems/month](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/daishisystems/fallback?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Build Status](https://travis-ci.org/daishisystems/fallback.svg?branch=master)](https://travis-ci.org/daishisystems/fallback)
[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/daishisystems/fallback)

Package fallback enhances the durability of your API by automatically recovering from connectivity failure. It achieves this by providing an enhanced degree of redundancy to HTTP requests, introducing a <a href="https://en.wikipedia.org/wiki/Chain-of-responsibility_pattern">Chain of Responsibility</a>, consisting of a series of fallback HTTP requests designed to augment an initial HTTP request. Should the initial HTTP request fail, the next fallback HTTP request in the chain will execute.

Any number of fallback HTTP requests can be chained sequentially. Redundancy is achieved by executing each fallback HTTP request in a recursive manner until one of the requests succeeds, or all requests fail.
![Icon](https://dl.dropboxusercontent.com/u/26042707/Fallback_XS.jpg)
## Installation
go get github.com/daishisystems/fallback
## Sample Code
```go
package main

import (
	"fmt"
	"github.com/daishisystems/fallback"
)

func main() {

	type BasicResponse struct {
		Text   string
		Detail string
	}

	type BasicError struct {
		Code    int
		Message string
	}

	basicResponse := &BasicResponse{}
	basicError := &BasicError{}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	working := fallback.NewConnection("Working", "GET",
		"http://demo7227109.mockable.io/get-basic", nil, headers,
		basicResponse, basicError, nil)

	failing := fallback.NewConnection("Failing", "GET",
		"http://demo7227109.mockable.io/fail-basic", nil, headers,
		basicResponse, basicError, working)

	statusCode, err := failing.ExecuteHTTPRequest()
	if err != nil {
		panic(err)
	}

	fmt.Printf("HTTP status code: %d\n", statusCode)
	fmt.Printf("Text: %s\n", basicResponse.Text)
	fmt.Printf("Detail: %s", basicResponse.Detail)
}
```
## Contact the Developer
Please reach out and contact me for questions, suggestions, or to just talk tech in general.


<a href="http://insidethecpu.com/feed/">![RSS](https://dl.dropboxusercontent.com/u/26042707/rss.png)</a><a href="https://twitter.com/daishisystems">![Twitter](https://dl.dropboxusercontent.com/u/26042707/twitter.png)</a><a href="https://www.linkedin.com/in/daishisystems">![LinkedIn](https://dl.dropboxusercontent.com/u/26042707/linkedin.png)</a><a href="https://plus.google.com/102806071104797194504/posts">![Google+](https://dl.dropboxusercontent.com/u/26042707/g.png)</a><a href="https://www.youtube.com/user/daishisystems">![YouTube](https://dl.dropboxusercontent.com/u/26042707/youtube.png)</a>
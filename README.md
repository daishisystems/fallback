# Go Fallback Package
[![Join the chat at https://gitter.im/daishisystems/month](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/daishisystems/fallback?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Build Status](https://travis-ci.org/daishisystems/fallback.svg?branch=master)](https://travis-ci.org/daishisystems/fallback)
[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/daishisystems/fallback)

Package fallback enhances the durability of your API by automatically recovering from connectivity failure. It achieves this by providing an enhanced degree of redundancy to HTTP requests, introducing a <a href="https://en.wikipedia.org/wiki/Chain-of-responsibility_pattern">Chain of Responsibility</a>, consisting of a series of fallback HTTP requests designed to augment an initial HTTP request. Should the initial HTTP request fail, the next fallback HTTP request in the chain will execute.

Any number of fallback HTTP requests can be chained sequentially. Redundancy is achieved by executing each fallback HTTP request in a recursive manner until one of the requests succeeds, or all requests fail.
![Icon](https://dl.dropboxusercontent.com/u/26042707/Fallback_XS.jpg)
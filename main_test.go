package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type TestSuiteMain struct {
	suite.Suite
}

func (t *TestSuiteMain) SetupSuite() {
}

func TestMainTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuiteMain))
}

func (t *TestSuiteMain) TestCreateRequestExpired() {

	// TEST Create data
	req, err := http.NewRequest("POST", "", bytes.NewBuffer([]byte(`{
		"Key": "key2",
		"Data": "bGtua2pubQ==",
		"Timestamp": "2012-11-01T22:08:41Z"
	}`)))
	if err != nil {
		t.FailNow(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(post)

	// Our handlers satisfy http.IndexHandler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusCreated {
		t.FailNow("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	// TEST Create duplicate data with duplicate key
	req, err = http.NewRequest("POST", "", bytes.NewBuffer([]byte(`{
		"Key": "key2",
		"Data": "bGtua1pubQ==",
		"Timestamp": "2022-11-01T22:08:41Z"
	}`)))
	if err != nil {
		t.FailNow(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(post)

	// Our handlers satisfy http.IndexHandler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusCreated {
		t.FailNow("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	// TEST Get data from duplicate key
	req, err = http.NewRequest("GET", "", bytes.NewBuffer([]byte(`{
		"Key": "key2"
	}`)))
	if err != nil {
		t.FailNow(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(get)

	// Our handlers satisfy http.IndexHandler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.FailNow("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.FailNow(err.Error())
	}

	var rawJsonRequest createRequestRaw
	err = json.Unmarshal(body, &rawJsonRequest)
	if err != nil {
		t.FailNow(err.Error())
	}

	if rawJsonRequest.Key != "key2" {
		t.FailNow("Returned wrong key: got %s want %s",
			rawJsonRequest.Key, "key2")
	}

	if b64.StdEncoding.EncodeToString(rawJsonRequest.Data) != "bGtua1pubQ==" {
		t.FailNow("Returned wrong data: got %v want %v",
			rawJsonRequest.Data, []byte("bGtua1pubQ=="))
	}

	if rawJsonRequest.Timestamp != "2022-11-01T22:08:41Z" {
		t.FailNow("Returned wrong timestamp: got %v want %v",
			rawJsonRequest.Timestamp, "2022-11-01T22:08:41Z")
	}

	// TEST Create data with different key but same value
	req, err = http.NewRequest("POST", "", bytes.NewBuffer([]byte(`{
		"Key": "key3",
		"Data": "bGtua1pubQ==",
		"Timestamp": "2022-11-01T22:08:41Z"
	}`)))
	if err != nil {
		t.FailNow(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(post)
	// Our handlers satisfy http.IndexHandler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusCreated {
		t.FailNow("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	// TEST Get data from another key but with same data
	req, err = http.NewRequest("GET", "", bytes.NewBuffer([]byte(`{
		"Key": "key3"
	}`)))
	if err != nil {
		t.FailNow(err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(get)

	// Our handlers satisfy http.IndexHandler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.FailNow("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body, err = ioutil.ReadAll(rr.Body)
	if err != nil {
		t.FailNow(err.Error())
	}

	err = json.Unmarshal(body, &rawJsonRequest)
	if err != nil {
		t.FailNow(err.Error())
	}

	if rawJsonRequest.Key != "key3" {
		t.FailNow("Returned wrong key: got "+rawJsonRequest.Key+" want %v",
			"key3")
	}

	if b64.StdEncoding.EncodeToString(rawJsonRequest.Data) != "bGtua1pubQ==" {
		t.FailNow("Returned wrong data: got %v want %v",
			rawJsonRequest.Data, []byte("bGtua1pubQ=="))
	}

	if rawJsonRequest.Timestamp != "2022-11-01T22:08:41Z" {
		t.FailNow("Returned wrong timestamp: got %v want %v",
			rawJsonRequest.Timestamp, "2022-11-01T22:08:41Z")
	}

}

package tests

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Gamma169/go-server-helpers/server"
	"github.com/google/jsonapi"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

/*********************************************
 * Helpers
 * *******************************************/
type testStruct struct {
	Id                 string `json:"id" jsonapi:"primary,model"`
	Name               string `json:"name" jsonapi:"attr,name"`
	callWhenValdidated func()
}

func (ts testStruct) Validate() error {
	ts.callWhenValdidated()
	return nil
}

const jsonapiFmt = `{
	"data": {
		"type": "model",
		"id": "%s",
		"attributes": {
			"name": "%s"
		}
	}
}`
const jsonFmt = `{
	"id":"%s",
	"name":"%s"
}
`

/*********************************************
 * Test Preprocess
 * *******************************************/

func TestPreProcessInput(t *testing.T) {
	req, err := http.NewRequest("GET", "/"+randString(25), nil)
	ok(t, err)

	unmarshalFn := func(interface{}, *http.Request) error { return nil }

	validateCalled := false
	ts := testStruct{
		callWhenValdidated: func() {
			validateCalled = true
		},
	}
	assert(t, !validateCalled, "validateCalled should start as false")
	err = server.PreProcessInput(ts, 500, httptest.NewRecorder(), req, unmarshalFn)
	ok(t, err)
	assert(t, validateCalled, "Should call Validate function when processed")
}

func TestUnmarshalObjectFromHeaders(t *testing.T) {

	testCases := []struct {
		id         string
		name       string
		header     string
		fmtStr     string
		shouldPass bool
	}{
		{uuid.New().String(), randString(60), server.JSONContentType, jsonFmt, true},
		{uuid.New().String(), randString(60), server.JSONContentType, jsonFmt, true},
		{uuid.New().String(), randString(60), jsonapi.MediaType, jsonapiFmt, true},
		{uuid.New().String(), randString(60), jsonapi.MediaType, jsonapiFmt, true},
		{uuid.New().String(), randString(60), "another-type", jsonFmt, false},
		{uuid.New().String(), randString(60), "not right", jsonapiFmt, false},
	}

	for _, tc := range testCases {

		bodyStr := fmt.Sprintf(tc.fmtStr, tc.id, tc.name)
		req, err := http.NewRequest("GET", "/"+randString(25), strings.NewReader(bodyStr))
		ok(t, err)
		req.Header.Add(server.ContentTypeHeader, tc.header)

		ts := testStruct{}
		err = server.UnmarshalObjectFromHeaders(&ts, req)
		if tc.shouldPass {
			ok(t, err)
			equals(t, tc.id, ts.Id)
			equals(t, tc.name, ts.Name)
		} else {
			assert(t, err != nil, "Should throw error if header is not right type")
			equals(t, "", ts.Id)
			equals(t, "", ts.Name)
		}
	}
}

/*********************************************
 * Test Error Handling
 * *******************************************/

func TestSendErrorOnError(t *testing.T) {
	testCases := []struct {
		err       error
		status    int
		shouldLog bool
	}{
		{errors.New("some error"), 500, true},
		{errors.New("another error"), 404, true},
		{errors.New("yet another"), 400, true},
		{nil, 200, false},
	}

	for _, tc := range testCases {

		logFnCalled := false
		var errorPassedToLogFn error
		logFn := func(e error, r *http.Request) {
			logFnCalled = true
			errorPassedToLogFn = e
		}

		recorder := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/"+randString(25), nil)
		ok(t, err)
		server.SendErrorOnError(tc.err,
			tc.status,
			recorder,
			req,
			logFn,
		)

		if tc.shouldLog {
			assert(t, logFnCalled, "Should call log function if err")
		} else {
			assert(t, !logFnCalled, "Should not call log function if err")
		}
		equals(t, tc.err, errorPassedToLogFn)
	}

}

/*********************************************
 * Test Writing Outputs
 * *******************************************/

// TODO: This function does not test the jsonapi accept header type
func TestWriteModelToResponseFromHeaders(t *testing.T) {
	testCases := []struct {
		id     string
		name   string
		header string
		status int
		fmtStr string
	}{
		{uuid.New().String(), randString(60), server.JSONContentType, 200, jsonFmt},
		{uuid.New().String(), randString(60), server.JSONContentType, 201, jsonFmt},
		{uuid.New().String(), randString(60), jsonapi.MediaType, 202, jsonapiFmt},
		{uuid.New().String(), randString(60), jsonapi.MediaType, 200, jsonapiFmt},
		{uuid.New().String(), randString(60), server.JSONContentType, 400, jsonFmt},
		{uuid.New().String(), randString(60), server.JSONContentType, 500, jsonFmt},
	}

	for _, tc := range testCases {

		req, err := http.NewRequest("GET", "/"+randString(25), nil)
		ok(t, err)
		req.Header.Add(server.ContentTypeHeader, tc.header)

		ts := testStruct{
			Id:   tc.id,
			Name: tc.name,
		}

		recorder := httptest.NewRecorder()

		err = server.WriteModelToResponseFromHeaders(&ts, tc.status, recorder, req)
		ok(t, err)

		equals(t, tc.status, recorder.Code)
		equals(t, tc.header, recorder.Result().Header.Get(server.ContentTypeHeader))

		expectedBodyStr := fmt.Sprintf(tc.fmtStr, tc.id, tc.name)
		expectedBodyStr = strings.ReplaceAll(expectedBodyStr, " ", "")
		expectedBodyStr = strings.ReplaceAll(expectedBodyStr, "\n", "")
		expectedBodyStr = strings.ReplaceAll(expectedBodyStr, "\t", "")

		buf := new(bytes.Buffer)
		buf.ReadFrom(recorder.Result().Body)
		bodyStr := buf.String()

		equals(t, expectedBodyStr+"\n", bodyStr)
	}

}

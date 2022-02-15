package tests

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Gamma169/go-server-helpers/server"
	"github.com/google/jsonapi"
	"github.com/google/uuid"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

/*********************************************
 * Helpers
 * *******************************************/
type testStruct struct {
	Id                 string `json:"id"   jsonapi:"primary,model"`
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

	unmarshalFnCalled := false
	unmarshalFn := func(interface{}, *http.Request) error {
		unmarshalFnCalled = true
		return nil
	}

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
	assert(t, unmarshalFnCalled, "Should call unmarshalFn passed in")

	// Should return nil and not call unmarshall func if provided nil input
	unmarshalFnCalled = false
	err = server.PreProcessInput(nil, 0, httptest.NewRecorder(), req, unmarshalFn)
	ok(t, err)
	assert(t, !unmarshalFnCalled, "Should NOT call unmarshalFn when input is nil")
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


func TestPreProcessInputFromJSONAPI(t *testing.T) {

	testCases := []struct {
		id         string
		name       string
	}{
		{uuid.New().String(), randString(60)},
		{uuid.New().String(), randString(60)},
		{uuid.New().String(), randString(60)},
	}

	for _, tc := range testCases {

		bodyStr := fmt.Sprintf(jsonapiFmt, tc.id, tc.name)
		req, err := http.NewRequest("GET", "/"+randString(25), strings.NewReader(bodyStr))
		ok(t, err)

		ts := testStruct{callWhenValdidated: func() {}}
		err = server.PreProcessInputFromJSONAPI(&ts, 5000, httptest.NewRecorder(), req)

		ok(t, err)
		equals(t, tc.id, ts.Id)
		equals(t, tc.name, ts.Name)
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
 * Test Standard Handlers
 * *******************************************/

// This function is basically a combination of all the others.  Almost like an integration test
func TestStandardRequestHandler(t *testing.T) {

	val := randString(60)
	var valReceived string
	ts := testStruct{
		callWhenValdidated: func() {
			valReceived = val
		},
	}
	max := rand.Intn(10000)
	preProcessCalled := false
	var maxReceived int
	preprocessFunc := func(i server.InputObject, maxB int, w http.ResponseWriter, r *http.Request) error {
		maxReceived = maxB
		preProcessCalled = true
		return nil
	}

	statusToReturn := rand.Intn(401) + 100
	logicOutput := randString(600)
	// Calling validate in logic in order to test server.InputObject properly
	logicFunc := func(i server.InputObject, r *http.Request) (interface{}, int, error) {
		i.Validate()
		return &logicOutput, statusToReturn, nil
	}

	recorder := httptest.NewRecorder()
	var statusRecieved int
	var responseReceived interface{}
	responseFunc := func(i interface{}, status int, w http.ResponseWriter, r *http.Request) error {
		responseReceived = i
		statusRecieved = status
		recorder.Code = status
		return nil
	}

	req, err := http.NewRequest("GET", "/"+randString(25), nil)
	ok(t, err)

	logFnCalled := false
	logFn := func(e error, r *http.Request) {
		logFnCalled = true
	}

	// FUNCTION TO TEST:
	err = server.StandardRequestHandler(
		ts,
		max,
		preprocessFunc,
		logicFunc,
		responseFunc,
		recorder,
		req,
		logFn,
	)
	ok(t, err)

	assert(t, val == valReceived, "Assert that logic function has been called")
	equals(t, max, maxReceived)
	assert(t, preProcessCalled, "Should have called PreprocessFunc")

	// If I change test to return a pointer
	equals(t, logicOutput, *(responseReceived.(*string)))
	// If I change test to be not a pointer
	// equals(t, logicOutput, responseReceived)

	equals(t, statusToReturn, statusRecieved)
	equals(t, statusToReturn, recorder.Code)
	assert(t, !logFnCalled, "Should not have called log fn")

	// Checking returns 400 on preprocess error

	// Important to use new recorder
	recorder = httptest.NewRecorder()
	errStr := randString(60)
	badPreprocessFunc := func(i server.InputObject, maxB int, w http.ResponseWriter, r *http.Request) error {
		return errors.New(errStr)
	}

	// FUNCTION TO TEST:
	err = server.StandardRequestHandler(
		ts,
		max,
		badPreprocessFunc,
		logicFunc,
		responseFunc,
		recorder,
		req,
		logFn,
	)

	equals(t, 400, recorder.Code)
	assert(t, err != nil, "Should return the error")
	equals(t, errStr, err.Error())
	assert(t, logFnCalled, "Should call error logfn when errored")
	buf := new(bytes.Buffer)
	buf.ReadFrom(recorder.Result().Body)
	bodyStr := buf.String()
	equals(t, errStr+"\n", bodyStr)

	// Checking returned status on logic error

	logFnCalled = false
	statusRecieved = 0
	err = nil
	statusToReturn = rand.Intn(401) + 100
	// Important to use new recorder
	recorder = httptest.NewRecorder()

	errStr = randString(60)
	badLogicFunc := func(i server.InputObject, r *http.Request) (interface{}, int, error) {
		return nil, statusToReturn, errors.New(errStr)
	}

	// FUNCTION TO TEST:
	err = server.StandardRequestHandler(
		&ts,
		max,
		preprocessFunc,
		badLogicFunc,
		responseFunc,
		recorder,
		req,
		logFn,
	)

	assert(t, err != nil, "Should return the error")
	equals(t, errStr, err.Error())

	equals(t, statusToReturn, recorder.Code)
	assert(t, logFnCalled, "Should call error logfn when errored")
	buf = new(bytes.Buffer)
	buf.ReadFrom(recorder.Result().Body)
	bodyStr = buf.String()
	equals(t, errStr+"\n", bodyStr)
}

// This test checks the AgnosticHandler, but effectively tests everything else under the hood
// Acts as a kind of integration test
// If broken, fix any other lower-level test
func TestStandardAgnosticRequestHandler(t *testing.T) {

	testCases := []struct {
		id     string
		name   string
		header string
		status int
		fmtStr string
	}{
		{uuid.New().String(), randString(60), server.JSONContentType, 200, jsonFmt},
		{uuid.New().String(), randString(60), server.JSONContentType, 201, jsonFmt},
		{uuid.New().String(), randString(60), jsonapi.MediaType, 200, jsonapiFmt},
		{uuid.New().String(), randString(60), jsonapi.MediaType, 201, jsonapiFmt},
		{uuid.New().String(), randString(60), server.JSONContentType, rand.Intn(401) + 100, jsonFmt},
		{uuid.New().String(), randString(60), server.JSONContentType, rand.Intn(401) + 100, jsonFmt},
		{uuid.New().String(), randString(60), jsonapi.MediaType, rand.Intn(401) + 100, jsonapiFmt},
		{uuid.New().String(), randString(60), jsonapi.MediaType, rand.Intn(401) + 100, jsonapiFmt},
	}

	for _, tc := range testCases {

		val := randString(60)
		var valReceived string
		ts := testStruct{
			callWhenValdidated: func() {
				valReceived = val
			},
		}

		outputStruct := testStruct{}
		populateOutput := func(id, name string) {
			outputStruct.Id = id + "_returned"
			outputStruct.Name = name + "_returned"
		}
		var idSet string
		var nameSet string
		// Calling validate in logic in order to test server.InputObject properly
		logicFunc := func(i server.InputObject, r *http.Request) (interface{}, int, error) {
			i.Validate()
			tsType := i.(*testStruct)
			idSet = tsType.Id
			nameSet = tsType.Name
			populateOutput(tsType.Id, tsType.Name)
			// Must return pointer in logif function
			return &outputStruct, tc.status, nil
		}

		bodyStr := fmt.Sprintf(tc.fmtStr, tc.id, tc.name)
		req, err := http.NewRequest("GET", "/"+randString(25), strings.NewReader(bodyStr))
		ok(t, err)
		req.Header.Add(server.ContentTypeHeader, tc.header)

		recorder := httptest.NewRecorder()

		logFnCalled := false
		logFn := func(e error, r *http.Request) {
			logFnCalled = true
		}

		// FUNCTION TO TEST:
		err = server.StandardAgnosticRequestHandler(
			&ts,
			rand.Intn(10000)+2000,
			logicFunc,
			recorder,
			req,
			logFn,
		)
		ok(t, err)

		// Check status first because it is easiest to diagnose
		// (like if preprocess inputs fails)
		equals(t, tc.status, recorder.Code)

		// Checks that vals have been assigned in logic func
		equals(t, tc.id, idSet)
		equals(t, tc.name, nameSet)
		// Checks vals have been assiged at the end of the func
		equals(t, tc.id, ts.Id)
		equals(t, tc.name, ts.Name)
		// Checks logic func was called
		assert(t, val == valReceived, "Assert that logic function has been called")

		equals(t, tc.header, recorder.Result().Header.Get(server.ContentTypeHeader))
		assert(t, !logFnCalled, "Should not have called log fn")

		expectedBodyStr := fmt.Sprintf(tc.fmtStr, outputStruct.Id, outputStruct.Name)
		expectedBodyStr = strings.ReplaceAll(expectedBodyStr, " ", "")
		expectedBodyStr = strings.ReplaceAll(expectedBodyStr, "\n", "")
		expectedBodyStr = strings.ReplaceAll(expectedBodyStr, "\t", "")

		buf := new(bytes.Buffer)
		buf.ReadFrom(recorder.Result().Body)
		bodyStr = buf.String()

		equals(t, expectedBodyStr, bodyStr)
	}

}

/*********************************************
 * Test Writing Outputs
 * *******************************************/

// TODO: This function does not test the jsonapi ACCEPT header only the Content-Type header
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
		// Testcase for nil pointer
		{"", "", "", rand.Intn(401) + 100, ""},
		{"", "", "", rand.Intn(401) + 100, ""},
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

		if tc.id == "" {
			err = server.WriteModelToResponseFromHeaders(nil, tc.status, recorder, req)
			ok(t, err)

			equals(t, tc.status, recorder.Code)
		} else {
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

			equals(t, expectedBodyStr, bodyStr)
		}
	}

}

func TestCheckJSONMarshalAndWrite(t *testing.T) {
	testCases := []struct {
		data       interface{}
		status     int
		expected   string
		shouldPass bool
	}{
		{"some-val", 200, `"some-val"`, true},
		{"another-val", 201, `"another-val"`, true},
		{4, rand.Intn(401) + 100, "4", true},
		{map[string]interface{}{"foo": "bar", "baz": 6}, rand.Intn(401) + 100, `{"baz":6,"foo":"bar"}`, true},
		{0.254, rand.Intn(401) + 100, "0.254", true},
		{0.254, rand.Intn(401) + 100, "0.254", true},
		{math.Inf(1), 0, "", false},
		// Prob can add more failing cases
	}

	for _, tc := range testCases {
		recorder := httptest.NewRecorder()
		err := server.CheckJSONMarshalAndWrite(tc.data, tc.status, recorder)
		if tc.shouldPass {
			ok(t, err)
			buf := new(bytes.Buffer)
			buf.ReadFrom(recorder.Result().Body)
			bodyStr := buf.String()
			equals(t, tc.expected, bodyStr)
		} else {
			assert(t, err != nil, "Should return err if json marshal error")
		}
	}
}

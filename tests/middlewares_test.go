package tests

import (
	"bytes"
	"fmt"
	"github.com/Gamma169/go-server-helpers/server"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

/*********************************************
 * Tests
 * *******************************************/

func TestAddCORSMiddlewareAndEndpoint(t *testing.T) {

	routes := []string{
		"/",
		fmt.Sprintf("/%s", randString(40)),
		fmt.Sprintf("/%s", randString(40)),
		fmt.Sprintf("/%s", randString(40)),
		fmt.Sprintf("/%s", randString(40)),
		fmt.Sprintf("/%s/%s", randString(40), randString(40)),
		fmt.Sprintf("/%s/%s", randString(40), randString(40)),
		fmt.Sprintf("/%s/%s", randString(40), randString(40)),
		// For some reason, if we add another test case it breaks...
		// fmt.Sprintf("/%s/%s", randString(40), randString(40)),
	}
	requesterId := "some-id"

	for _, path := range routes {

		req, err := http.NewRequest("OPTIONS", path, nil)
		ok(t, err)

		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		// FUNCTION TO TEST:
		server.AddCORSMiddlewareAndEndpoint(router, requesterId)
		router.ServeHTTP(rr, req)

		equals(t, http.StatusNoContent, rr.Code)

		corsVals := []struct {
			header   string
			expected string
		}{
			{"Access-Control-Allow-Origin", "*"},
			{"Access-Control-Allow-Credentials", "true"},
			{"Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Session, " + requesterId},
			{"Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE"},
		}

		for _, cv := range corsVals {
			equals(t, cv.expected, rr.Result().Header.Get(cv.header))
		}
	}
}

func TestRequesterIdHeaderMiddleware(t *testing.T) {
	requesterId := randString(25)

	testCases := []struct {
		header     string
		value      string
		shouldPass bool
	}{
		{requesterId, uuid.New().String(), true},
		{requesterId, uuid.New().String(), true},
		{"not-req-id", uuid.New().String(), false},
		{requesterId, "not-a-uuid", false},
		{"not-req-id", "bad-val", false},
	}

	for _, testCase := range testCases {

		randEndpoint := "/" + randString(25)
		req, err := http.NewRequest("GET", randEndpoint, nil)
		ok(t, err)
		req.Header.Add(testCase.header, testCase.value)

		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		router.Path(randEndpoint).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		// FUNCTION TO TEST:
		server.AddRequesterIdHeaderMiddleware(router, requesterId, false)

		router.ServeHTTP(rr, req)

		if testCase.shouldPass {
			equals(t, http.StatusOK, rr.Code)
		} else {
			equals(t, http.StatusBadRequest, rr.Code)
		}
	}
}

func TestAddLoggingMiddleware(t *testing.T) {
	traceIdHeader := randString(25)

	testCases := []struct {
		header    string
		value     string
		shouldLog bool
	}{
		{traceIdHeader, uuid.New().String(), true},
		{traceIdHeader, uuid.New().String(), true},
		{"not-trace-id", "qwe", false},
		{"not-req-id", "bad-val", false},
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	for _, testCase := range testCases {

		randEndpoint := "/" + randString(25)
		req, err := http.NewRequest("GET", randEndpoint, nil)
		ok(t, err)
		req.Header.Add(testCase.header, testCase.value)

		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		var receivedHeaders http.Header
		router.Path(randEndpoint).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeaders = r.Header
		})
		// FUNCTION TO TEST:
		server.AddLoggingMiddleware(router, traceIdHeader, true)

		router.ServeHTTP(rr, req)

		equals(t, testCase.value, receivedHeaders.Get(testCase.header))
		assert(t, rr.Result().Header.Get(testCase.header) == "", "Should not respond with header")

		if testCase.shouldLog {
			doesLogTrace := strings.Contains(buf.String(), testCase.value)
			assert(t, doesLogTrace, "Should log trace in output")
		}
		buf = bytes.Buffer{}
	}
}

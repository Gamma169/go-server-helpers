package server

import (
	"encoding/json"
	"errors"
	"github.com/google/jsonapi"
	"net/http"
)

const ContentTypeHeader = "Content-Type"
const AcceptContentTypeHeader = "Accept"

const JSONContentType = "application/json"

type InputObject interface {
	Validate() error
}

/*********************************************
 * Pre-process + Read Inputs
 * *******************************************/

// Be sure to pass in pointer to InputObject
// Usage:  PreProcessInput(&model, 500, w, r, fn)
func PreProcessInput(input InputObject, maxBytes int, w http.ResponseWriter, r *http.Request, unmarshalFn func(interface{}, *http.Request) error) error {
	if input == nil {
		return nil
	}

	max := 524288
	if maxBytes != 0 {
		max = maxBytes
	}
	// Block the read of any body too large in order to help prevent DoS attacks
	r.Body = http.MaxBytesReader(w, r.Body, int64(max))
	if err := unmarshalFn(input, r); err != nil {
		return err
	}

	return input.Validate()
}

// Common convenience functions

func PreProcessInputFromHeaders(input InputObject, maxBytes int, w http.ResponseWriter, r *http.Request) error {
	return PreProcessInput(input, maxBytes, w, r, UnmarshalObjectFromHeaders)
}

func PreProcessInputFromJSON(input InputObject, maxBytes int, w http.ResponseWriter, r *http.Request) error {
	return PreProcessInput(input, maxBytes, w, r, UnmarshalObjectFromJSONStrict)
}

func PreProcessInputFromJSONAPI(input InputObject, maxBytes int, w http.ResponseWriter, r *http.Request) error {
	return PreProcessInput(input, maxBytes, w, r, func(interface{}, *http.Request) error {
		return jsonapi.UnmarshalPayload(r.Body, &input)
	})
}

// Unmarshalling logic

func UnmarshalObjectFromHeaders(input interface{}, r *http.Request) error {
	header := r.Header.Get(ContentTypeHeader)
	if header == JSONContentType {
		return UnmarshalObjectFromJSONStrict(input, r)
	} else if header == jsonapi.MediaType {
		return jsonapi.UnmarshalPayload(r.Body, input)
	} else {
		return errors.New("Content-Type header is not json or jsonapi standard")
	}
}

func UnmarshalObjectFromJSON(input interface{}, r *http.Request) error {
	dec := json.NewDecoder(r.Body)
	return dec.Decode(input)
}

func UnmarshalObjectFromJSONStrict(input interface{}, r *http.Request) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(input)
}

/*********************************************
 * Error Handling
 * *******************************************/

func SendErrorOnError(err error, status int, w http.ResponseWriter, r *http.Request, logError func(error, *http.Request)) {
	if err != nil {
		logError(err, r)
		http.Error(w, err.Error(), status)
	}
}

/*********************************************
 * Standard Handlers
 *
 * These ones are the trickiest to use since they're the most abstracted
 * Pass in functions (or use the convenience versions in this file) in order to reduce boilerplate code
 *
 * They work on the idea that every handler has 3 parts-
 *   - Preprocess (read input)
 *   - Business Logic
 *   - Response
 * Pass in functions that do each of the three things, and only focus on actual code
 *
 * NOTE: Be sure to pass in a pointer to inputPtr or else they won't work
 * Can pass in <nil> to inputPtr in order to avoid calling preprocessFunc
 * (useful for json handler that outputs json but doesn't require input)
 * *******************************************/

func StandardRequestHandler(
	inputPtr InputObject,
	maxBytes int,
	preprocessFunc func(InputObject, int, http.ResponseWriter, *http.Request) error,
	logicFunc func(InputObject, *http.Request) (interface{}, int, error),
	responseFunc func(interface{}, int, http.ResponseWriter, *http.Request) error,
	w http.ResponseWriter,
	r *http.Request,
	logError func(error, *http.Request),
) {
	var err error
	var errStatus = http.StatusInternalServerError

	defer func() { SendErrorOnError(err, errStatus, w, r, logError) }()

	if err = preprocessFunc(inputPtr, maxBytes, w, r); err != nil {
		errStatus = http.StatusBadRequest
		return
	}

	var status int
	var outputStruct interface{}
	if outputStruct, status, err = logicFunc(inputPtr, r); err != nil {
		errStatus = status
		return
	}

	err = responseFunc(outputStruct, status, w, r)
}

func StandardJSONRequestHandler(
	inputPtr InputObject,
	maxBytes int,
	logicFunc func(InputObject, *http.Request) (interface{}, int, error),
	w http.ResponseWriter,
	r *http.Request,
	logError func(error, *http.Request),
) {
	// Note that the json write function below does not need the request
	// So we wrap it in an anonymous function in order to fit the area 'StandardRequestHandler' expects
	jsonRespFuncWrapper := func(input interface{}, status int, w http.ResponseWriter, r *http.Request) error {
		return WriteModelToResponseJSON(input, status, w)
	}
	StandardRequestHandler(inputPtr, maxBytes, PreProcessInputFromJSON, logicFunc, jsonRespFuncWrapper, w, r, logError)
}

func StandardAgnosticRequestHandler(
	inputPtr InputObject,
	maxBytes int,
	logicFunc func(InputObject, *http.Request) (interface{}, int, error),
	w http.ResponseWriter,
	r *http.Request,
	logError func(error, *http.Request),
) {
	StandardRequestHandler(inputPtr, maxBytes, PreProcessInputFromHeaders, logicFunc, WriteModelToResponseFromHeaders, w, r, logError)
}

/*********************************************
 * Writing Outputs
 * *******************************************/

func WriteModelToResponseJSON(dataToSend interface{}, status int, w http.ResponseWriter) (err error) {
	w.Header().Set(ContentTypeHeader, JSONContentType)
	return CheckJSONMarshalAndWrite(dataToSend, status, w)
}

func WriteModelToResponseJSONAPI(dataToSend interface{}, status int, w http.ResponseWriter) (err error) {
	w.Header().Set(ContentTypeHeader, jsonapi.MediaType)
	var jsonAPIPayload jsonapi.Payloader
	if jsonAPIPayload, err = jsonapi.Marshal(dataToSend); err != nil {
		return
	}
	return CheckJSONMarshalAndWrite(jsonAPIPayload, status, w)
}

func WriteModelToResponseFromHeaders(dataToSend interface{}, status int, w http.ResponseWriter, r *http.Request) error {
	header := r.Header.Get(ContentTypeHeader)
	acceptHeader := r.Header.Get(AcceptContentTypeHeader)
	if header == jsonapi.MediaType || acceptHeader == jsonapi.MediaType {
		return WriteModelToResponseJSONAPI(dataToSend, status, w)
	}
	return WriteModelToResponseJSON(dataToSend, status, w)
}

// Convenience function so you don't have to write your own wrapper function if you don't want to return NoContent
func WriteNoContentToResponse(dataToSend interface{}, status int, w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}

/*
	This function manually marshals the json payload, and checks to see if it marshals to valid json before writing the header and response
	We use this function in order to write a proper error status if the passed-in data does not marshal properly

	If we do this:
		w.WriteHeader(status)
		return json.NewEncoder(w).Encode(payload)
	Then, if the Encode function errors, the status will already have been set, and mask the output from the server
	(For example, we write 200 status, then the function errors, and an error bubbles-up, but a 500 error code cannot be written because the 200 was already written)

	Likewise if we do this:
		if err = json.NewEncoder(w).Encode(payload); err == nil {
			w.WriteHeader(status)
		}
	Then, only a 200 code is ever written and the WriteHeader call is ignored because data has already been written to the ResponseWriter

	Thus we can use this function to capture the error appropriately

	See golang ResponseWriter for how information on the `WriteHeader` works
	https://pkg.go.dev/net/http@go1.17.1#ResponseWriter

	NOTE:  The `w.Write` call can also fail (due to network isues for example )
	in which case the server will not respond with the correct status code regardless (because it cannot write ANY status code)
	But we still capture the error to bubble it up and do any logging/handling on the server side
*/
func CheckJSONMarshalAndWrite(data interface{}, status int, w http.ResponseWriter) (err error) {
	var jsonData []byte
	if jsonData, err = json.Marshal(data); err != nil {
		return
	}
	w.WriteHeader(status)
	// NOTE: this can also fail, see note above
	_, err = w.Write(jsonData)
	return
}

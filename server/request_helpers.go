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

func PreProcessInput(input InputObject, maxBytes int, w http.ResponseWriter, r *http.Request, unmarshalFn func(interface{}, *http.Request) error) error {

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
 * Writing Outputs
 * *******************************************/

func WriteModelToResponseJSON(dataToSend interface{}, status int, w http.ResponseWriter) error {
	w.Header().Set(ContentTypeHeader, JSONContentType)
	// NOTE: w.WriteHeader must go after all other header writes but before json encoder or else it/headers will not be picked up
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(dataToSend)
}

func WriteModelToResponseJSONAPI(dataToSend interface{}, status int, w http.ResponseWriter) error {
	w.Header().Set(ContentTypeHeader, jsonapi.MediaType)
	// NOTE: w.WriteHeader must go after all other header writes but before json encoder or else it/headers will not be picked up
	w.WriteHeader(status)
	return jsonapi.MarshalPayload(w, dataToSend)
}

func WriteModelToResponseFromHeaders(dataToSend interface{}, status int, w http.ResponseWriter, r *http.Request) error {
	header := r.Header.Get(ContentTypeHeader)
	acceptHeader := r.Header.Get(AcceptContentTypeHeader)
	if header == jsonapi.MediaType || acceptHeader == jsonapi.MediaType {
		return WriteModelToResponseJSONAPI(dataToSend, status, w)
	}
	return WriteModelToResponseJSON(dataToSend, status, w)
}

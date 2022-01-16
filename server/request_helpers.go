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

func PreProcessInput(input InputObject, w http.ResponseWriter, r *http.Request, maxBytes int) (error, int) {
	var err error

	max := 131072
	if maxBytes != 0 {
		max = maxBytes
	}
	// Block the read of any body too large in order to help prevent DoS attacks
	r.Body = http.MaxBytesReader(w, r.Body, int64(max))

	header := r.Header.Get(ContentTypeHeader)
	if header == JSONContentType {

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		if err = dec.Decode(input); err != nil {
			return err, http.StatusBadRequest
		}

	} else if header == jsonapi.MediaType {

		if err := jsonapi.UnmarshalPayload(r.Body, &input); err != nil {
			return err, http.StatusBadRequest
		}

	} else {
		return errors.New("Content-Type header is not json or jsonapi standard"), http.StatusBadRequest
	}

	if err = input.Validate(); err != nil {
		return err, http.StatusBadRequest
	}

	return nil, 0
}

func SendErrorOnError(err error, status int, w http.ResponseWriter, r *http.Request, logError func(error, *http.Request)) {
	if err != nil {
		logError(err, r)
		http.Error(w, err.Error(), status)
	}
}

func WriteJSONModelToResponse(dataToSend interface{}, status int, w http.ResponseWriter) error {
	w.Header().Set(ContentTypeHeader, JSONContentType)
	// NOTE: w.WriteHeader must go after all other header writes but before json encoder or else it/headers will not be picked up
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(dataToSend)
}

func WriteJSONAPIModelToResponse(dataToSend interface{}, status int, w http.ResponseWriter) error {
	w.Header().Set(ContentTypeHeader, jsonapi.MediaType)
	// NOTE: w.WriteHeader must go after all other header writes but before json encoder or else it/headers will not be picked up
	w.WriteHeader(status)
	return jsonapi.MarshalPayload(w, dataToSend)
}

func WriteAgnosticModelToResponse(dataToSend interface{}, status int, w http.ResponseWriter, r *http.Request) error {
	header := r.Header.Get(ContentTypeHeader)
	acceptHeader := r.Header.Get(AcceptContentTypeHeader)
	if header == jsonapi.MediaType || acceptHeader == jsonapi.MediaType {
		return WriteJSONAPIModelToResponse(dataToSend, status, w)
	}
	return WriteJSONModelToResponse(dataToSend, status, w)
}

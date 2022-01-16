package server

import (
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const boldPrint = "\033[1m"
const headerPrint = "\033[95m"
const endPrint = "\033[0m"

// This adds CORS headers for all requests (to use if web server is running locally)
func AddCORSMiddlewareAndEndpoint(router *mux.Router, requesterIdHeader string) {
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Session, "+requesterIdHeader)
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")
			next.ServeHTTP(w, r)
		})
	})
	// Any options requests return 204-- to be used with above CORS stuff
	router.PathPrefix("/").Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
}

// Check if the requesterIdHeader exists and is valid and return 400 if not
func AddRequesterIdHeaderMiddleware(router *mux.Router, requesterIdHeader string, debug bool) {
	router.Use(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				requesterId := r.Header.Get(requesterIdHeader)
				if requesterId == "" {
					msg := "No '" + requesterIdHeader + "' header"
					if debug {
						log.Println(msg)
					}
					http.Error(w, msg, http.StatusBadRequest)
					return
				}
				if _, err := uuid.Parse(requesterId); err != nil {
					msg := requesterIdHeader + "- is not valid UUID"
					if debug {
						log.Println(msg)
					}
					http.Error(w, msg, http.StatusBadRequest)
					return
				}

				// Call the next handler, which can be another middleware in the chain, or the final handler.
				next.ServeHTTP(w, r)
			})
		},
	)
}

func AddLoggingMiddleware(router *mux.Router, traceIdHeader string, debug bool) {
	router.Use(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestId := r.Header.Get(traceIdHeader)
				if requestId == "" {
					requestId = uuid.New().String()
					r.Header.Set(traceIdHeader, requestId)
				}

				if debug {
					log.Println(boldPrint, headerPrint, "Received:", r.RequestURI, "--", requestId, endPrint)
				}
				next.ServeHTTP(w, r)
				// TODO: Prob implement custom responseWriter to be able to get status from request
				// https://github.com/Gamma169/go-service-template/issues/2
				if debug {
					log.Println(boldPrint, headerPrint, "Finished:", r.RequestURI, "--", requestId, "--", "[]", endPrint)
				}
			})
		},
	)
}

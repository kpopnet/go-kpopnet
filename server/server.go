package server

import (
	"net/http"

	"github.com/dimfeld/httptreemux/v5"
)

// Start starts HTTP server on specified address.
func Start(address string) (err error) {
	router := createRouter()
	return http.ListenAndServe(address, router)
}

func createRouter() http.Handler {
	r := httptreemux.New()

	api := r.UsingContext().NewGroup("/api")
	api.GET("/profiles", ServeProfiles)
	api.POST("/recognize", ServeRecognize)

	return http.Handler(r)
}

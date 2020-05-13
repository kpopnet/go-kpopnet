package server

import (
	"net/http"

	"github.com/dimfeld/httptreemux/v5"
)

// StartServer starts HTTP server on specified address.
func Start(address string) (err error) {
	router := createRouter()
	return http.ListenAndServe(address, router)
}

func createRouter() http.Handler {
	r := httptreemux.New()

	r.GET("/api/idols/profiles", ServeProfiles)
	r.POST("/api/idols/recognize", ServeRecognize)

	return http.Handler(r)
}

package server

import (
	"encoding/json"
	"net/http"

	"github.com/kpopnet/go-kpopnet"
	"github.com/kpopnet/go-kpopnet/cache"
	"github.com/kpopnet/go-kpopnet/db"
	"github.com/kpopnet/go-kpopnet/facerec"
)

const (
	maxOverheadSize = int64(10 * 1024)
	maxFileSize     = int64(5 * 1024 * 1024)
	maxBodySize     = maxFileSize + maxOverheadSize
)

// ServeProfiles returns a JSON object with information about all profiles.
func ServeProfiles(w http.ResponseWriter, r *http.Request) {
	// TODO(Kagami): For some reason cached request is not fast enough.
	// TODO(Kagami): Use some trigger to invalidate cache.
	v, err := cache.Cached(cache.ProfileCacheKey, func() (v interface{}, err error) {
		ps, err := db.GetProfiles()
		if err != nil {
			return
		}
		// Takes ~5ms so better to store encoded.
		return json.Marshal(ps)
	})
	if err != nil {
		serve500(w, r, err)
		return
	}
	serveEncodedJSON(w, r, v.([]byte))
}

// ServeRecognize recognizes image uploaded via HTTP.
func ServeRecognize(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	if err := r.ParseMultipartForm(0); err != nil {
		serveError(w, r, kpopnet.ErrParseForm, 400)
		return
	}
	fhs := r.MultipartForm.File["files[]"]
	if len(fhs) != 1 {
		serve400(w, r, kpopnet.ErrParseFile)
		return
	}
	idolID, err := facerec.RequestRecognizeMultipart(fhs[0])
	switch err {
	case kpopnet.ErrParseFile:
		serve400(w, r, err)
		return
	case kpopnet.ErrBadImage:
		serve400(w, r, err)
		return
	case kpopnet.ErrNoIdol:
		serve400(w, r, err)
		return
	case nil:
		// Do nothing.
	default:
		serve500(w, r, err)
		return
	}
	if idolID == nil {
		serve400(w, r, kpopnet.ErrNoSingleFace)
		return
	}
	result := map[string]string{"id": *idolID}
	serveJSON(w, r, result)
}

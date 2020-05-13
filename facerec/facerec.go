package facerec

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg" // JPEG decoder
	"io/ioutil"
	"mime/multipart"

	"github.com/kpopnet/go-kpopnet"
	"github.com/kpopnet/go-kpopnet/cache"
	"github.com/kpopnet/go-kpopnet/db"

	"github.com/Kagami/go-face"
)

const (
	// Maximum number of recognizer threads executing at the same time.
	numRecWorkers = 1

	minDimension = 300
	maxDimension = 5000
)

var (
	faceRec *face.Recognizer
	recJobs = make(chan recRequest)
)

type recRequest struct {
	fh *multipart.FileHeader
	ch chan<- recResult
}

type recResult struct {
	idolID *string
	err    error
}

// StartFaceRec initializes face recognition.
func StartFaceRec(modelDir string) (err error) {
	faceRec, err = face.NewRecognizer(modelDir)
	if err != nil {
		return fmt.Errorf("error initializing face recognizer: %v", err)
	}
	for i := 0; i < numRecWorkers; i++ {
		go recWorker()
	}
	return
}

// Execute recognizing jobs.
func recWorker() {
	for {
		req := <-recJobs
		idolID, err := recognizeMultipart(req.fh)
		req.ch <- recResult{idolID, err}
	}
}

// RequestRecognizeMultipart recognizes provided image.
func RequestRecognizeMultipart(fh *multipart.FileHeader) (idolID *string, err error) {
	ch := make(chan recResult)
	go func() {
		recJobs <- recRequest{fh, ch}
	}()
	res := <-ch
	return res.idolID, res.err
}

// Simple wrapper to work with uploaded files.
// Recognize immediately.
func recognizeMultipart(fh *multipart.FileHeader) (idolID *string, err error) {
	fd, err := fh.Open()
	if err != nil {
		err = kpopnet.ErrParseFile
		return
	}
	defer fd.Close()
	imgData, err := ioutil.ReadAll(fd)
	if err != nil {
		err = kpopnet.ErrParseFile
		return
	}
	idolID, err = recognize(imgData)
	return
}

// Recognize immediately.
// TODO(Kagami): Search for already recognized idol using imageId.
func recognize(imgData []byte) (idolID *string, err error) {
	// TODO(Kagami): Invalidate?
	v, err := cache.Cached(cache.TrainDataCacheKey, func() (interface{}, error) {
		data, err := db.GetTrainData()
		if err == nil {
			faceRec.SetSamples(data.Samples, data.Cats)
		}
		return data, err
	})
	if err != nil {
		return
	}
	data := v.(*kpopnet.TrainData)

	r := bytes.NewReader(imgData)
	c, typ, err := image.DecodeConfig(r)
	if err != nil || typ != "jpeg" ||
		c.Width < minDimension ||
		c.Height < minDimension ||
		c.Width > maxDimension ||
		c.Height > maxDimension ||
		c.ColorModel != color.YCbCrModel {
		err = kpopnet.ErrBadImage
		return
	}

	f, err := faceRec.RecognizeSingle(imgData)
	if _, ok := err.(face.ImageLoadError); ok {
		err = kpopnet.ErrBadImage
	}
	if err != nil || f == nil {
		return
	}

	catID := faceRec.Classify(f.Descriptor)
	if catID < 0 {
		err = kpopnet.ErrNoIdol
		return
	}
	id := data.Labels[catID]
	return &id, nil
}

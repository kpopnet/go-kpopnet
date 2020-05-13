package kpopnet

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg" // JPEG decoder
	"io/ioutil"
	"mime/multipart"
	"unsafe"

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

type trainData struct {
	samples []face.Descriptor
	cats    []int32
	labels  map[int]string
}

// StartFaceRec initializes facerec module.
func StartFaceRec(dataDir string) error {
	return startFaceRec(getModelsDir(dataDir))
}

// Useful for tests.
func startFaceRec(modelsDir string) (err error) {
	faceRec, err = face.NewRecognizer(modelsDir)
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
		err = errParseFile
		return
	}
	defer fd.Close()
	imgData, err := ioutil.ReadAll(fd)
	if err != nil {
		err = errParseFile
		return
	}
	idolID, err = recognize(imgData)
	return
}

// Recognize immediately.
// TODO(Kagami): Search for already recognized idol using imageId.
func recognize(imgData []byte) (idolID *string, err error) {
	// TODO(Kagami): Invalidate?
	v, err := cached(trainDataCacheKey, func() (interface{}, error) {
		data, err := getTrainData()
		if err == nil {
			// NOTE(Kagami): We don't copy this data to C++ side so need to
			// keep in cache to prevent GC.
			faceRec.SetSamples(data.samples, data.cats)
		}
		return data, err
	})
	if err != nil {
		return
	}
	data := v.(*trainData)

	r := bytes.NewReader(imgData)
	c, typ, err := image.DecodeConfig(r)
	if err != nil || typ != "jpeg" ||
		c.Width < minDimension ||
		c.Height < minDimension ||
		c.Width > maxDimension ||
		c.Height > maxDimension ||
		c.ColorModel != color.YCbCrModel {
		err = errBadImage
		return
	}

	f, err := faceRec.RecognizeSingle(imgData)
	if _, ok := err.(face.ImageLoadError); ok {
		err = errBadImage
	}
	if err != nil || f == nil {
		return
	}

	catID := faceRec.Classify(f.Descriptor)
	if catID < 0 {
		err = errNoIdol
		return
	}
	id := data.labels[catID]
	return &id, nil
}

// Get all confirmed face descriptors.
func getTrainData() (data *trainData, err error) {
	var samples []face.Descriptor
	var cats []int32
	labels := make(map[int]string)

	rs, err := prepared["get_train_data"].Query()
	if err != nil {
		return
	}
	defer rs.Close()
	var catID int32
	var prevIdolID string
	catID = -1
	for rs.Next() {
		var idolID string
		var descrBytes []byte
		if err = rs.Scan(&idolID, &descrBytes); err != nil {
			return
		}
		descriptor := bytes2descr(descrBytes)
		samples = append(samples, descriptor)
		if idolID != prevIdolID {
			catID++
			labels[int(catID)] = idolID
		}
		cats = append(cats, catID)
		prevIdolID = idolID
	}
	if err = rs.Err(); err != nil {
		return
	}

	data = &trainData{
		samples: samples,
		cats:    cats,
		labels:  labels,
	}
	return
}

// Zero-copy conversions.

func descr2bytes(d face.Descriptor) []byte {
	size := unsafe.Sizeof(d)
	return (*[1 << 30]byte)(unsafe.Pointer(&d))[:size:size]
}

func bytes2descr(b []byte) face.Descriptor {
	return *(*face.Descriptor)(unsafe.Pointer(&b[0]))
}

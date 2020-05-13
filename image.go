package kpopnet

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/Kagami/go-face"
)

type namesKey [2]string
type idolNamesMap map[namesKey]Idol

func getImagesDir(d string) string {
	return filepath.Join(d, "images")
}

func getSha1(data []byte) string {
	hash := sha1.Sum(data)
	return hex.EncodeToString(hash[:])
}

func getNamesKey(bname, iname string) namesKey {
	return [2]string{bname, iname}
}

func getIdolNamesKey(idol Idol, bandByID map[string]Band) (key namesKey, ok bool) {
	iname, ok := idol["name"].(string)
	if !ok {
		return
	}
	bandID, ok := idol["band_id"].(string)
	if !ok {
		return
	}
	band, ok := bandByID[bandID]
	if !ok {
		return
	}
	bname, ok := band["name"].(string)
	if !ok {
		return
	}
	return getNamesKey(bname, iname), true
}

func getIdolNamesMap() (idolByNames idolNamesMap, err error) {
	tx, err := beginTx()
	if err != nil {
		return
	}
	defer endTx(tx, &err)
	if err = setReadOnly(tx); err != nil {
		return
	}
	_, bandByID, err := getBands(tx)
	if err != nil {
		return
	}
	idols, _, err := getIdols(tx)
	if err != nil {
		return
	}
	idolByNames = make(idolNamesMap)
	for _, idol := range idols {
		if key, ok := getIdolNamesKey(idol, bandByID); ok {
			idolByNames[key] = idol
		}
	}
	return
}

func recognizeIdolImage(ipath string) (f *face.Face, id string, err error) {
	fd, err := os.Open(ipath)
	if err != nil {
		return
	}
	imgData, err := ioutil.ReadAll(fd)
	if err != nil {
		return
	}
	f, err = faceRec.RecognizeSingle(imgData)
	if err != nil || f == nil {
		return
	}
	id = getSha1(imgData)
	return
}

// TODO(Kagami): Use multiple threads?
func recognizeIdolImages(idir string) (faces []face.Face, ids []string, err error) {
	idolImages, err := ioutil.ReadDir(idir)
	if err != nil {
		return
	}
	// No need to validate names/formats because everything was checked by
	// Python spider.
	for _, file := range idolImages {
		var f *face.Face
		var imageID string
		fname := file.Name()
		ipath := filepath.Join(idir, fname)
		f, imageID, err = recognizeIdolImage(ipath)
		if err != nil {
			return
		}
		if f == nil {
			log.Printf("Not a single face on %s", ipath)
			continue
		}
		faces = append(faces, *f)
		ids = append(ids, imageID)
	}
	return
}

func importIdolImages(st *sql.Stmt, idir string, idol Idol) (err error) {
	faces, imageIDs, err := recognizeIdolImages(idir)
	if err != nil {
		return
	}
	idolID := idol["id"].(string)
	for i, f := range faces {
		rectStr := rect2str(f.Rectangle)
		descrBytes := descr2bytes(f.Descriptor)
		imageID := imageIDs[i]
		confirmed := true
		source := "googleimages"
		_, err = st.Exec(rectStr, descrBytes, imageID, idolID, confirmed, source)
		if err != nil {
			return
		}
	}
	return
}

func importBandImages(bdir, bname string, idolByNames idolNamesMap) (err error) {
	// Use single transaction per band.
	tx, err := beginTx()
	if err != nil {
		return
	}
	defer endTx(tx, &err)
	st := tx.Stmt(prepared["upsert_face"])

	idolDirs, err := ioutil.ReadDir(bdir)
	if err != nil {
		return
	}
	for _, dir := range idolDirs {
		iname := dir.Name()
		idir := filepath.Join(bdir, iname)
		key := getNamesKey(bname, iname)
		idol, ok := idolByNames[key]
		if !ok {
			log.Printf("Can't find %s (%s)", iname, bname)
			continue
		}
		log.Printf("Importing %s", idir)
		if err = importIdolImages(st, idir, idol); err != nil {
			return
		}
	}
	return
}

// ImportImages reads and updates idols recognition info in DB.
func ImportImages(connStr string, modelDir string, imageDir string, onlyBands []string) (err error) {
	if err = StartDB(nil, connStr); err != nil {
		return
	}

	if err = StartFaceRec(modelDir); err != nil {
		return
	}

	idolByNames, err := getIdolNamesMap()
	if err != nil {
		err = fmt.Errorf("error querying idols: %v", err)
		return
	}

	bandDirs, err := ioutil.ReadDir(imageDir)
	if err != nil {
		err = fmt.Errorf("error reading bands: %v", err)
		return
	}

	bandFilter := make(map[string]bool)
	for _, bname := range onlyBands {
		bandFilter[bname] = true
	}

	for _, dir := range bandDirs {
		bname := dir.Name()
		if len(bandFilter) > 0 && !bandFilter[bname] {
			continue
		}
		bdir := filepath.Join(imageDir, bname)
		if err = importBandImages(bdir, bname, idolByNames); err != nil {
			err = fmt.Errorf("error importing %s images: %v", bname, err)
			return
		}
	}
	return
}

package db

import (
	"database/sql"
	"encoding/json"

	"github.com/Kagami/go-face"
	k "github.com/kpopnet/go-kpopnet"
)

// Get all bands.
func getBands(tx *sql.Tx) (bands []k.Band, bandByID map[string]k.Band, err error) {
	bands = make([]k.Band, 0)
	bandByID = make(map[string]k.Band)
	rs, err := tx.Stmt(prepared["get_bands"]).Query()
	if err != nil {
		return
	}
	defer rs.Close()
	for rs.Next() {
		var id string
		var data []byte
		var band k.Band
		if err = rs.Scan(&id, &data); err != nil {
			return
		}
		if err = json.Unmarshal(data, &band); err != nil {
			return
		}
		band["id"] = id
		bands = append(bands, band)
		bandByID[id] = band
	}
	if err = rs.Err(); err != nil {
		return
	}
	return
}

// Get all idols.
func getIdols(tx *sql.Tx) (idols []k.Idol, idolByID map[string]k.Idol, err error) {
	idols = make([]k.Idol, 0)
	idolByID = make(map[string]k.Idol)
	rs, err := tx.Stmt(prepared["get_idols"]).Query()
	if err != nil {
		return
	}
	defer rs.Close()
	for rs.Next() {
		var id string
		var bandID string
		var data []byte
		var idol k.Idol
		if err = rs.Scan(&id, &bandID, &data); err != nil {
			return
		}
		if err = json.Unmarshal(data, &idol); err != nil {
			return
		}
		idol["id"] = id
		idol["band_id"] = bandID
		idols = append(idols, idol)
		idolByID[id] = idol
	}
	if err = rs.Err(); err != nil {
		return
	}
	return
}

// Get and set idol preview property.
func getIdolPreviews(tx *sql.Tx, idolByID map[string]k.Idol) (err error) {
	rs, err := tx.Stmt(prepared["get_idol_previews"]).Query()
	if err != nil {
		return
	}
	defer rs.Close()
	for rs.Next() {
		var idolID string
		var imageID string
		if err = rs.Scan(&idolID, &imageID); err != nil {
			return
		}
		if idol, ok := idolByID[idolID]; ok {
			idol["image_id"] = imageID
		}
	}
	if err = rs.Err(); err != nil {
		return
	}
	return
}

// GetProfiles queries all profiles.
func GetProfiles() (ps *k.Profiles, err error) {
	tx, err := beginTx()
	if err != nil {
		return
	}
	defer endTx(tx, &err)
	bands, _, err := getBands(tx)
	if err != nil {
		return
	}
	idols, idolByID, err := getIdols(tx)
	if err != nil {
		return
	}
	err = getIdolPreviews(tx, idolByID)
	if err != nil {
		return
	}

	ps = &k.Profiles{
		Bands: bands,
		Idols: idols,
	}
	return
}

// GetMaps returns idols/bands maps accessable by ID.
func GetMaps() (idolByID map[string]k.Idol, bandByID map[string]k.Band, err error) {
	tx, err := beginTx()
	if err != nil {
		return
	}
	defer endTx(tx, &err)
	if _, idolByID, err = getIdols(tx); err != nil {
		return
	}
	if _, bandByID, err = getBands(tx); err != nil {
		return
	}
	return
}

// GetTrainData returns confirmed face descriptors.
func GetTrainData() (data *k.TrainData, err error) {
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

	data = &k.TrainData{
		Samples: samples,
		Cats:    cats,
		Labels:  labels,
	}
	return
}

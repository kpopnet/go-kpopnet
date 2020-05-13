package db

import (
	"database/sql"
	"encoding/json"

	"github.com/Kagami/go-face"
	"github.com/kpopnet/go-kpopnet"
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

// Prepare band structure to be stored in DB.
// ID fields are removed to avoid duplication.
func getBandData(band k.Band) (data []byte, err error) {
	delete(band, "id")
	delete(band, "urls") // Don't need this
	data, err = json.Marshal(band)
	return
}

// Prepare idol structure to be stored in DB.
// ID fields are removed to avoid duplication.
func getIdolData(idol k.Idol) (data []byte, err error) {
	delete(idol, "id")
	delete(idol, "band_id")
	data, err = json.Marshal(idol)
	return
}

// UpdateProfiles inserts or updates database profiles.
func UpdateProfiles(ps *k.Profiles) (err error) {
	tx, err := beginTx()
	if err != nil {
		return
	}
	defer endTx(tx, &err)

	st := tx.Stmt(prepared["upsert_band"])
	for _, band := range ps.Bands {
		id := band["id"]
		var data []byte
		data, err = getBandData(band)
		if err != nil {
			return
		}
		if _, err = st.Exec(id, data); err != nil {
			return
		}
	}

	st = tx.Stmt(prepared["upsert_idol"])
	for _, idol := range ps.Idols {
		id := idol["id"]
		bandID := idol["band_id"]
		var data []byte
		data, err = getIdolData(idol)
		if err != nil {
			return
		}
		if _, err = st.Exec(id, bandID, data); err != nil {
			return
		}
	}

	return
}

// GetImageInfo returns recognition info for the specified image.
func GetImageInfo(imageID string) (info *k.ImageInfo, err error) {
	var rectStr string
	var idolID string
	var confirmed bool
	err = prepared["get_face"].QueryRow(imageID).Scan(&rectStr, &idolID, &confirmed)
	if err != nil {
		if err == sql.ErrNoRows {
			err = kpopnet.ErrNoIdol
		}
		return
	}
	rect := str2rect(rectStr)
	info = &k.ImageInfo{Rectangle: rect, IdolID: idolID, Confirmed: confirmed}
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

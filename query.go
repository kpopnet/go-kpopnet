package kpopnet

import (
	"database/sql"
	"encoding/json"
)

// Get all bands.
func getBands(tx *sql.Tx) (bands []Band, bandByID map[string]Band, err error) {
	bands = make([]Band, 0)
	bandByID = make(map[string]Band)
	rs, err := tx.Stmt(prepared["get_bands"]).Query()
	if err != nil {
		return
	}
	defer rs.Close()
	for rs.Next() {
		var id string
		var data []byte
		var band Band
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
func getIdols(tx *sql.Tx) (idols []Idol, idolByID map[string]Idol, err error) {
	idols = make([]Idol, 0)
	idolByID = make(map[string]Idol)
	rs, err := tx.Stmt(prepared["get_idols"]).Query()
	if err != nil {
		return
	}
	defer rs.Close()
	for rs.Next() {
		var id string
		var bandID string
		var data []byte
		var idol Idol
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
func getIdolPreviews(tx *sql.Tx, idolByID map[string]Idol) (err error) {
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
func GetProfiles() (ps *Profiles, err error) {
	tx, err := beginTx()
	if err != nil {
		return
	}
	defer endTx(tx, &err)
	if err = setReadOnly(tx); err != nil {
		return
	}
	if err = setRepeatableRead(tx); err != nil {
		return
	}

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

	ps = &Profiles{
		Bands: bands,
		Idols: idols,
	}
	return
}

// Prepare band structure to be stored in DB.
// ID fields are removed to avoid duplication.
func getBandData(band Band) (data []byte, err error) {
	delete(band, "id")
	delete(band, "urls") // Don't need this
	data, err = json.Marshal(band)
	return
}

// Prepare idol structure to be stored in DB.
// ID fields are removed to avoid duplication.
func getIdolData(idol Idol) (data []byte, err error) {
	delete(idol, "id")
	delete(idol, "band_id")
	data, err = json.Marshal(idol)
	return
}

// UpdateProfiles inserts or updates database profiles.
func UpdateProfiles(ps *Profiles) (err error) {
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

func getImageInfo(imageID string) (info *ImageInfo, err error) {
	var rectStr string
	var idolID string
	var confirmed bool
	err = prepared["get_face"].QueryRow(imageID).Scan(&rectStr, &idolID, &confirmed)
	if err != nil {
		if err == sql.ErrNoRows {
			err = errNoIdol
		}
		return
	}
	rect := str2rect(rectStr)
	info = &ImageInfo{rect, idolID, confirmed}
	return
}

package facerec

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kpopnet/go-kpopnet/db"
)

const (
	testDir  = "../testdata"
	testConn = "user=meguca password=meguca dbname=meguca sslmode=disable"
)

var (
	// TODO(Kagami): Grab a lot of test data to test against regressions.
	testData = map[string]string{
		"elkie.jpg":      "Elkie, CLC",
		"chaeyoung.jpg":  "Chaeyoung, Twice",
		"chaeyoung2.jpg": "Chaeyoung, Twice",
		"sejeong.jpg":    "Sejeong, Gugudan",
		"jimin.jpg":      "Jimin, AOA",
		"jimin2.jpg":     "Jimin, AOA",
		"jimin4.jpg":     "Jimin, AOA",
		"meiqi.jpg":      "Mei Qi, WJSN",
		"chaeyeon.jpg":   "Chaeyeon, DIA",
		"chaeyeon3.jpg":  "Chaeyeon, DIA",
		"tzuyu2.jpg":     "Tzuyu, Twice",
		"nayoung.jpg":    "Nayoung, PRISTIN",
		"luda2.jpg":      "Luda, WJSN",
		"joy.jpg":        "Joy, Red Velvet",
		// Currently failing.
		// "bona.jpg": "Bona, WJSN",
		// "bona2.jpg": "Bona, WJSN",
		// "bona3.jpg": "Bona, WJSN",
		// "bona4.jpg": "Bona, WJSN",
		// "nana.jpg": "Nana, After School",
		// "chaeyeon2.jpg": "Chaeyeon, DIA",
		// "luda.jpg": "Luda, WJSN",
		// "eunseo2.jpg": "Eunseo, WJSN",
		// "eunseo3.jpg": "Eunseo, WJSN",
		// "yujin.jpg": "Yujin, CLC",
		// "tzuyu.jpg": "Tzuyu, Twice",
		// "seulgi.jpg": "Seulgi, Red Velvet",
		// "eunwoo.jpg": "Eunwoo, PRISTIN",
		// "rena.jpg": "Rena, PRISTIN",
		// "jimin5.jpg": "Jimin, AOA",
		// "jimin6.jpg": "Jimin, AOA",
		// "jimin7.jpg": "Jimin, AOA",
	}
)

func getTestFilePath(fname string) string {
	return filepath.Join(testDir, "images", fname)
}

func recognizeFile(fpath string) (idolID *string, err error) {
	fd, err := os.Open(fpath)
	if err != nil {
		return
	}
	imgData, err := ioutil.ReadAll(fd)
	if err != nil {
		return
	}
	return recognize(imgData)
}

func TestIdols(t *testing.T) {
	if err := db.Start(nil, testConn); err != nil {
		t.Fatal(err)
	}
	if err := Start(filepath.Join(testDir, "models")); err != nil {
		t.Fatal(err)
	}
	idolByID, bandByID, err := db.GetMaps()
	if err != nil {
		t.Fatal(err)
	}
	for fname, expected := range testData {
		t.Run(fname, func(t *testing.T) {
			names := strings.Split(expected, ", ")
			expectedIname := names[0]
			expectedBname := names[1]

			actualIdolID, err := recognizeFile(getTestFilePath(fname))
			if err != nil {
				t.Fatal(err)
			}
			if actualIdolID == nil {
				t.Errorf("%s: expected “%s” but not recognized", fname, expected)
				return
			}

			idol := idolByID[*actualIdolID]
			band := bandByID[idol["band_id"].(string)]
			actualIname := idol["name"]
			actualBname := band["name"]
			if expectedIname != actualIname || expectedBname != actualBname {
				actual := fmt.Sprintf("%s, %s", actualIname, actualBname)
				t.Errorf("%s: expected “%s” but got “%s”", fname, expected, actual)
			}
		})
	}
}

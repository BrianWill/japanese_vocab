package main

import (
	"archive/zip"
	"fmt"
	"io"
	"math"
)

func unzipSource(path string) ([]byte, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	bytes, err := unzipFile(reader.File[0])
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func unzipFile(f *zip.File) ([]byte, error) {
	zippedFile, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer zippedFile.Close()

	bytes, err := io.ReadAll(zippedFile)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func secondsToTimestamp(seconds float64) string {
	minutes := math.Floor(seconds / 60)
	seconds -= minutes * 60
	wholeSeconds := math.Floor(seconds)
	fracSeconds := seconds - wholeSeconds
	s := fmt.Sprintf("%d:%02d", int(minutes), int(seconds))
	if fracSeconds > 0 {
		s += ".5"
	}
	return s
}

func getVerbCategory(sense JMDictSense) int {
	category := 0
	for _, pos := range sense.Pos {
		switch pos {
		case "verb-ichidan":
			category |= DRILL_CATEGORY_ICHIDAN
		case "verb-godan-su":
			category |= DRILL_CATEGORY_GODAN_SU
		case "verb-godan-ku":
			category |= DRILL_CATEGORY_GODAN_KU
		case "verb-godan-gu":
			category |= DRILL_CATEGORY_GODAN_GU
		case "verb-godan-ru":
			category |= DRILL_CATEGORY_GODAN_RU
		case "verb-godan-u":
			category |= DRILL_CATEGORY_GODAN_U
		case "verb-godan-tsu":
			category |= DRILL_CATEGORY_GODAN_TSU
		case "verb-godan-mu":
			category |= DRILL_CATEGORY_GODAN_MU
		case "verb-godan-nu":
			category |= DRILL_CATEGORY_GODAN_NU
		case "verb-godan-bu":
			category |= DRILL_CATEGORY_GODAN_BU
		}
	}
	return category
}

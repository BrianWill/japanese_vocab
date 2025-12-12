package main

import (
	"archive/zip"
	"io"
	"net"
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

// Get preferred outbound ip of this machine
func GetOutboundIP() ([]string, error) {

	strs := make([]string, 0)

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	// handle err
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if !ip.IsLoopback() && ip.To4() != nil {
				strs = append(strs, ip.String())
			}
		}
	}

	return strs, nil
}

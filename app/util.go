package main

import (
	"archive/zip"
	"fmt"
	"io"
)

func unzipSource(path string) ([]byte, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	for _, f := range reader.File {
		bytes, err := unzipFile(f)
		if err != nil {
			return nil, err
		}
		return bytes, nil
	}

	return nil, fmt.Errorf("No content read from zip file: %s\n", path)
}

func unzipFile(f *zip.File) ([]byte, error) {

	// 7. Unzip the content of a file and copy it to the destination file
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

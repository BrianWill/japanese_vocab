package main

import (
	"archive/zip"
	"io"
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

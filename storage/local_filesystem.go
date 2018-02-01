package storage

import (
	b64 "encoding/base64"
	"fmt"
	"path"

	"arthurgustin.fr/teddycare/shared"
	"io/ioutil"
	"os"
)

const (
	jpegMimetype = "image/jpeg"
)

var (
	ErrUnsupportedFileFormat = fmt.Errorf("unsupported format. The only accepted format is %s", jpegMimetype)
)

type Storage interface {
	Store(encodedImage, mimeType string) (string, error)
	Get(filename string) (string, error)
	Delete(filename string) error
}

type LocalStorage struct {
	Config          *shared.AppConfig `inject:""`
	StringGenerator interface {
		GenerateUuid() string
	} `inject:""`
}

func (s *LocalStorage) Store(encodedImage, mimeType string) (string, error) {
	if mimeType != jpegMimetype {
		return "", ErrUnsupportedFileFormat
	}

	decoded, err := b64.StdEncoding.DecodeString(encodedImage)
	if err != nil {
		return "", err
	}

	id := s.StringGenerator.GenerateUuid()

	filePath := path.Clean(s.Config.LocalStoragePath + "/" + id + ".jpg")

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.Write(decoded)
	if err != nil {
		return "", err
	}
	file.Sync()

	return filePath, nil
}

func (s *LocalStorage) Get(filePath string) (string, error) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return b64.StdEncoding.EncodeToString(file), nil
}

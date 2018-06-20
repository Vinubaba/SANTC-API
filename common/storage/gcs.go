package storage

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type Options struct {
	CredentialsFile string
	BucketName      string
}

func New(ctx context.Context, options Options) (*GoogleStorage, error) {
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(options.CredentialsFile))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}
	gs := &GoogleStorage{
		client: client,
		bucket: options.BucketName,
	}

	b, err := ioutil.ReadFile(options.CredentialsFile)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(b, &gs.serviceAccountDetails); err != nil {
		return nil, err
	}

	return gs, nil
}

type GoogleStorage struct {
	client                *storage.Client
	bucket                string
	serviceAccountDetails serviceAccountDetails
	StringGenerator       interface {
		GenerateUuid() string
	} `inject:""`
}

func (s *GoogleStorage) Store(ctx context.Context, b64image string, folder string) (string, error) {
	if b64image == "" {
		return "", nil
	}
	encoded, err := s.validate64EncodedPhoto(b64image)
	if err != nil {
		return "", err
	}

	decoded, err := b64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	fileName := s.StringGenerator.GenerateUuid() + ".jpg"
	if folder != "" {
		fileName = folder + "/" + fileName
	}
	w := s.client.Bucket(s.bucket).Object(fileName).NewWriter(ctx)

	if _, err = w.Write(decoded); err != nil {
		return "", err
	}

	return fileName, w.Close()
}

func (s *GoogleStorage) validate64EncodedPhoto(photo string) (encoded string, err error) {
	if strings.HasPrefix(photo, "data:image/jpeg;base64,") {
		encoded = strings.TrimPrefix(photo, "data:image/jpeg;base64,")
	} else {
		err = ErrUnsupportedFileFormat
	}
	return
}

// returns signedUrls
func (s *GoogleStorage) Get(ctx context.Context, fileName string) (string, error) {
	if fileName == "" {
		return "", nil
	}
	url, err := storage.SignedURL(s.bucket, fileName, &storage.SignedURLOptions{
		GoogleAccessID: s.serviceAccountDetails.ClientEmail,
		PrivateKey:     []byte(s.serviceAccountDetails.PrivateKey),
		Method:         http.MethodGet,
		Expires:        time.Now().Add(time.Second * 180),
	})
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *GoogleStorage) Delete(ctx context.Context, fileName string) error {
	if fileName == "" {
		return nil
	}

	return s.client.Bucket(s.bucket).Object(fileName).Delete(ctx)
}

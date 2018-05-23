package storage

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Vinubaba/SANTC-API/api/shared"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type GoogleStorage struct {
	Config          *shared.AppConfig `inject:""`
	StringGenerator interface {
		GenerateUuid() string
	} `inject:""`
}

func (s *GoogleStorage) Store(ctx context.Context, b64image string) (string, error) {
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

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(s.Config.BucketServiceAccount))
	if err != nil {
		return "", fmt.Errorf("failed to create client: %v", err)
	}

	fileName := s.StringGenerator.GenerateUuid() + ".jpg"
	w := client.Bucket(s.Config.BucketImagesName).Object(fileName).NewWriter(ctx)
	defer w.Close()
	if _, err = w.Write(decoded); err != nil {
		return "", err
	}

	return fileName, nil
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
	url, err := storage.SignedURL(s.Config.BucketImagesName, fileName, &storage.SignedURLOptions{
		GoogleAccessID: s.Config.BucketServiceAccountDetails.ClientEmail,
		PrivateKey:     []byte(s.Config.BucketServiceAccountDetails.PrivateKey),
		Method:         http.MethodGet,
		Expires:        time.Now().Add(time.Second * 180),
	})
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *GoogleStorage) Delete(ctx context.Context, fileName string) error {
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(s.Config.BucketServiceAccount))
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	return client.Bucket(s.Config.BucketImagesName).Object(fileName).Delete(ctx)
}

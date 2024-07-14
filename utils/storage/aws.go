package storage

import (
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type AWS struct {
	Bucket  string
	Session *session.Session
}

func NewAWSStorage() *AWS {
	storage := &AWS{
		Bucket: os.Getenv("AWS_BUCKET"),
	}

	config := &aws.Config{
		Region: aws.String("eu-central-1"),
	}

	if os.Getenv("ENVIRONMENT") == "LOCAL DEV" {
		config.Credentials = credentials.NewStaticCredentials(
			os.Getenv("ACCESS_KEY"),
			os.Getenv("SECRET_KEY"),
			"",
		)
	}

	storage.Session, _ = session.NewSession(config)
	return storage
}

func (storage *AWS) GetFile(key string) ([]byte, error) {
	svc := s3.New(storage.Session)

	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(storage.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

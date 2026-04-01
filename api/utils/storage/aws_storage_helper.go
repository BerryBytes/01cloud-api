package storage

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
)

func CreateAwsStorageClient(storageMap map[string]string) (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(storageMap["region"]),
		Credentials: credentials.NewStaticCredentials(storageMap["access_key"], storageMap["secret_key"], ""),
	})
	if err != nil {
		log.Warnf("Could not create storage Client: %v", err)
	}

	return sess, err
}

func CreateAwsBucket(storageMap map[string]string, id uint) error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(fmt.Sprintf("zerone-bucket-%d", id)),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(storageMap["region"]),
		},
	}
	ses, err := CreateAwsStorageClient(storageMap)
	if err != nil {
		return err
	}
	svc := s3.New(ses)
	_, err = svc.CreateBucket(input)
	if err != nil {
		return err
	}
	return nil
}

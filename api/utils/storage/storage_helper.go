package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"01cloud-api/api/utils/helper"

	"cloud.google.com/go/storage"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/gosimple/slug"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

func CreateStorageClient() (*storage.Client, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		logrus.Error("Could not create storage Client:", err)
	}
	return client, err
}

func CreateStorageClientWithCredential(credentialFile string) (*storage.Client, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialFile))
	if err != nil {
		logrus.Error("Could not create storage Client:", err)
	}
	return client, err
}

func CreateBucket(client *storage.Client, project string, bucketName string) error {
	ctx := context.Background()
	bucket := client.Bucket(bucketName)
	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()
	if err := bucket.Create(ctx, project, nil); err != nil {
		return err
	}
	return nil
}
func DownloadFile(client *storage.Client, bucket, object string) (string, error) {
	ctx := context.Background()
	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return "", err
	}
	path := "./temp/terraform/" + bucket + "/" + object
	tarPath := fmt.Sprintf("%s/%s.tgz", path, object)
	helper.ExtractTarGz(rc, path)
	if _, er := os.Stat(path + "/s3"); er == nil {
		_ = os.Remove(path + "/s3/zerone-devops-lab.json")
		_ = os.RemoveAll(path + "/s3/.terraform")
		_ = os.Remove(path + "/s3/values.schema.json")
		_ = os.Remove(path + "/s3/credentials.json")
		err = helper.CreateTarFile(path+"/s3", tarPath)
	} else {
		_ = os.Remove(path + "/zerone-devops-lab.json")
		_ = os.RemoveAll(path + "/.terraform")
		_ = os.Remove(path + "/values.schema.json")
		_ = os.Remove(path + "/credentials.json")
		err = helper.CreateTarFile(path, tarPath)
	}
	if err != nil {
		return "", err
	}
	_ = helper.CreateIfNotExists(fmt.Sprintf("/data/uploads/terraform/%s/", bucket), 0775)
	downloadPath := fmt.Sprintf("/data/uploads/terraform/%s/%s.tgz", bucket, object)
	_, err = fileutils.CopyFile(tarPath, downloadPath)
	if err != nil {
		return "", err
	}
	downloadUrl := fmt.Sprintf("%s/uploads/terraform/%s/%s.tgz", os.Getenv("BASE_URL"), bucket, object)
	return downloadUrl, nil
}
func DownloadKubeConfigFile(client *storage.Client, bucket, object, name string) (string, error) {
	ctx := context.Background()
	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return "", err
	}
	path := "./temp/terraform/" + bucket + "/" + object
	helper.ExtractTarGz(rc, path)
	_ = helper.CreateIfNotExists(fmt.Sprintf("/data/config/%s/%s/", bucket, object), 0775)
	downloadPath := fmt.Sprintf("/data/config/%s/%s/kubeconfig", bucket, object)
	if _, err = os.Stat(path + "/s3"); err == nil {
		_, err = fileutils.CopyFile(fmt.Sprintf("%s/s3/kubeconfig_%s", path, slug.Make(name)), downloadPath)
	} else {
		_, err = fileutils.CopyFile(fmt.Sprintf("%s/kubeconfig_%s", path, slug.Make(name)), downloadPath)
	}
	if err != nil {
		return "", err
	}
	return downloadPath, nil
}

func DeleteFile(client *storage.Client, bucket, object string) {
	ctx := context.Background()
	_ = client.Bucket(bucket).Object(object).Delete(ctx)
}
func UploadFile(client *storage.Client, bucket, object string, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("os.Open: %v", err)
	}
	defer f.Close()
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()
	_ = client.Bucket(bucket).Object(object).Delete(ctx)
	wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}
	logrus.Info("File uploaded")
	return nil
}

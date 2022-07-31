package storage

import (
	"bytes"
	"io/ioutil"
	"log"
	"path"

	"github.com/minio/minio-go"
	"github.com/teamyapp/cloud/app/config"
)

const appDataRoot = "appData"

type S3Bucket struct {
	client     *minio.Client
	env        config.Environment
	bucketName string
}

var _ MapBackend = (*S3Bucket)(nil)

func (s S3Bucket) Get(key string) ([]byte, error) {
	fullPath := path.Join(appDataRoot, string(s.env), key)
	obj, err := s.client.GetObject(s.bucketName, fullPath, minio.GetObjectOptions{})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return ioutil.ReadAll(obj)
}

func (s S3Bucket) Put(key string, data []byte) error {
	objSize := int64(len(data))
	fullPath := path.Join(appDataRoot, string(s.env), key)
	_, err := s.client.PutObject(s.bucketName, fullPath, bytes.NewReader(data), objSize, minio.PutObjectOptions{})
	return err
}

func (s S3Bucket) Delete(key string) error {
	fullPath := path.Join(appDataRoot, string(s.env), key)
	return s.client.RemoveObject(s.bucketName, fullPath)
}

func NewS3Bucket(endpoint string, accessKeyID string, accessKey string, env config.Environment, bucketName string) (S3Bucket, error) {
	client, err := minio.New(endpoint, accessKeyID, accessKey, true)
	if err != nil {
		return S3Bucket{}, err
	}

	return S3Bucket{
		client:     client,
		env:        env,
		bucketName: bucketName,
	}, nil
}

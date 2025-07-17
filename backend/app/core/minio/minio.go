package minio

import (
	"context"
	"fmt"
	"os"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinioClient() (*minio.Client, error) {
	containerName := os.Getenv("MINIO_CONTAINER_NAME")
	port := os.Getenv("MINIO_PORT")
	endpoint := containerName + ":" + port

	accessKeyID := os.Getenv("MINIO_ROOT_USER")
	secretAccessKey := os.Getenv("MINIO_ROOT_PASSWORD")
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		fmt.Printf("Error creating MinIO client: %v", err)
		return nil, err
	}

	return minioClient, nil
}

func UploadFile(ctx context.Context, minioClient *minio.Client, bucketName string, objectName string, filePath string, contentType string) error {
	_, err := minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func DeleteFile(ctx context.Context, minioClient *minio.Client, bucketName, objectName string) error {
	return minioClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}


func GetFileContent(ctx context.Context, minioClient *minio.Client, bucketName, objectName string) ([]byte, error) {
	// Get object from MinIO
	obj, err := minioClient.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	// Read object content
	return io.ReadAll(obj)
}

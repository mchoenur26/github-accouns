package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// inisialisasi UI dan struktur
type FileUploader interface {
	UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, slug string) (string, error)
}

// struktur data
type S3Service struct {
	Client     *s3.Client
	BucketName string
}

// inisialisasi membuka data dan jika gagal
func (s *S3Service) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, slug string) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("gagal membuka file: %w", err)
	}
	defer file.Close()
	//read data dan mengubah gambar menjadi huruf kecil
	fileExtension := strings.ToLower(filepath.Ext(fileHeader.Filename))
	uniqueFileName := uuid.New().String() + fileExtension
	//memastikan penamaan file tidak berubah jika dicari
	cleanSlug := strings.ReplaceAll(slug, " ", "-")
	key := fmt.Sprintf("logo/%s/%s", cleanSlug, uniqueFileName)
	//inisialisasi penentuan tipe file
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	//upload ke database AWS
	_, err = s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.BucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil { //jika gagal
		return "", fmt.Errorf("gagal mengunggah ke S3: %w", err)
	} //membuat url file untuk diakses
	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.BucketName, key)

	return url, nil
}

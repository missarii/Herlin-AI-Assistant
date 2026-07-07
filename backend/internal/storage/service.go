// Herlin AI Assistant - Backend Service
// Copyright 2026 Herlin AI. All rights reserved.

package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"time"

	"github.com/herlin-ai/herlin-assistant/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Service struct {
	client *minio.Client
	cfg    *config.Config
	ctx    context.Context
}

type FileInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ContentType string
}

func NewService(cfg *config.Config) (*Service, error) {
	// Parse endpoint
	endpoint := cfg.Storage.Endpoint
	if !isValidEndpoint(endpoint) {
		return nil, fmt.Errorf("invalid storage endpoint: %s", endpoint)
	}

	// Initialize MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Storage.AccessKey, cfg.Storage.SecretKey, ""),
		Secure: false, // Set to true for HTTPS
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage client: %w", err)
	}

	// Create bucket if it doesn't exist
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Storage.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.Storage.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &Service{
		client: client,
		cfg:    cfg,
		ctx:    context.Background(),
	}, nil
}

func (s *Service) UploadFile(key string, data []byte, contentType string) error {
	reader := bytes.NewReader(data)
	_, err := s.client.PutObject(s.ctx, s.cfg.Storage.Bucket, key, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	return nil
}

func (s *Service) UploadFileWithMetadata(key string, data []byte, contentType string, metadata map[string]string) error {
	reader := bytes.NewReader(data)
	_, err := s.client.PutObject(s.ctx, s.cfg.Storage.Bucket, key, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType:  contentType,
		UserMetadata: metadata,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file with metadata: %w", err)
	}
	return nil
}

func (s *Service) DownloadFile(key string) ([]byte, error) {
	obj, err := s.client.GetObject(s.ctx, s.cfg.Storage.Bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	return data, nil
}

func (s *Service) DownloadFileAsStream(key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(s.ctx, s.cfg.Storage.Bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	return obj, nil
}

func (s *Service) DeleteFile(key string) error {
	err := s.client.RemoveObject(s.ctx, s.cfg.Storage.Bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *Service) DeleteFiles(keys []string) error {
	objectsCh := make(chan minio.ObjectInfo)
	go func() {
		for _, key := range keys {
			obj, err := s.client.StatObject(s.ctx, s.cfg.Storage.Bucket, key, minio.StatObjectOptions{})
			if err == nil {
				objectsCh <- obj
			}
		}
		close(objectsCh)
	}()

	errorCh := s.client.RemoveObjects(s.ctx, s.cfg.Storage.Bucket, objectsCh, minio.RemoveObjectsOptions{})

	for err := range errorCh {
		if err.Err != nil {
			return fmt.Errorf("failed to delete file %s: %w", err.ObjectName, err.Err)
		}
	}

	return nil
}

func (s *Service) FileExists(key string) (bool, error) {
	_, err := s.client.StatObject(s.ctx, s.cfg.Storage.Bucket, key, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}
	return true, nil
}

func (s *Service) GetFileInfo(key string) (*FileInfo, error) {
	stat, err := s.client.StatObject(s.ctx, s.cfg.Storage.Bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &FileInfo{
		Key:          key,
		Size:         stat.Size,
		LastModified: stat.LastModified,
		ContentType:  stat.ContentType,
	}, nil
}

func (s *Service) ListFiles(prefix string, recursive bool) ([]FileInfo, error) {
	var files []FileInfo

	objectsCh := s.client.ListObjects(s.ctx, s.cfg.Storage.Bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	})

	for object := range objectsCh {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list files: %w", object.Err)
		}

		files = append(files, FileInfo{
			Key:          object.Key,
			Size:         object.Size,
			LastModified: object.LastModified,
			ContentType:  object.ContentType,
		})
	}

	return files, nil
}

func (s *Service) GeneratePresignedURL(key string, expiration time.Duration) (string, error) {
	presignedURL, err := s.client.PresignedGetObject(s.ctx, s.cfg.Storage.Bucket, key, expiration, url.Values{})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}

func (s *Service) GeneratePresignedUploadURL(key string, expiration time.Duration) (string, error) {
	presignedURL, err := s.client.PresignedPutObject(s.ctx, s.cfg.Storage.Bucket, key, expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}
	return presignedURL.String(), nil
}

func (s *Service) CopyFile(sourceKey, destKey string) error {
	_, err := s.client.CopyObject(s.ctx, minio.CopyDestOptions{
		Bucket: s.cfg.Storage.Bucket,
		Object: destKey,
	}, minio.CopySrcOptions{
		Bucket: s.cfg.Storage.Bucket,
		Object: sourceKey,
	})
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	return nil
}

func (s *Service) MoveFile(sourceKey, destKey string) error {
	if err := s.CopyFile(sourceKey, destKey); err != nil {
		return err
	}
	return s.DeleteFile(sourceKey)
}

func (s *Service) GetBucketSize() (int64, error) {
	var totalSize int64

	objectsCh := s.client.ListObjects(s.ctx, s.cfg.Storage.Bucket, minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectsCh {
		if object.Err != nil {
			return 0, fmt.Errorf("failed to list files: %w", object.Err)
		}
		totalSize += object.Size
	}

	return totalSize, nil
}

func (s *Service) GetFileCount() (int, error) {
	count := 0

	objectsCh := s.client.ListObjects(s.ctx, s.cfg.Storage.Bucket, minio.ListObjectsOptions{
		Recursive: true,
	})

	for range objectsCh {
		count++
	}

	return count, nil
}

func (s *Service) CreateFolder(path string) error {
	key := filepath.Join(path, ".folder")
	return s.UploadFile(key, []byte(""), "application/x-directory")
}

func (s *Service) DeleteFolder(path string) error {
	files, err := s.ListFiles(path, true)
	if err != nil {
		return err
	}

	keys := make([]string, len(files))
	for i, file := range files {
		keys[i] = file.Key
	}

	return s.DeleteFiles(keys)
}

// Helper functions
func isValidEndpoint(endpoint string) bool {
	u, err := url.Parse("http://" + endpoint)
	if err != nil {
		return false
	}
	return u.Host != ""
}

func (s *Service) GenerateKey(userID uint, filename string) string {
	timestamp := time.Now().Format("20060102-150405")
	ext := filepath.Ext(filename)
	baseName := filename[:len(filename)-len(ext)]
	return fmt.Sprintf("users/%d/%s-%s%s", userID, baseName, timestamp, ext)
}

func (s *Service) GenerateDocumentKey(userID uint, documentID uint, filename string) string {
	return fmt.Sprintf("documents/%d/%d/%s", userID, documentID, filename)
}

func (s *Service) GenerateAvatarKey(userID uint, filename string) string {
	return fmt.Sprintf("avatars/%d/%s", userID, filename)
}

func (s *Service) GenerateMemoryKey(userID uint, memoryID uint) string {
	return fmt.Sprintf("memories/%d/%d", userID, memoryID)
}

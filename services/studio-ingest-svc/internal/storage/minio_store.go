package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStore struct {
	client *minio.Client
}

func NewMinioStore(endpoint, accessKey, secretKey string, useSSL bool) (*MinioStore, error) {
	cli, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &MinioStore{client: cli}, nil
}

func (m *MinioStore) PutChunkPart(ctx context.Context, bucket, sessionID string, chunkIdx int32, payload []byte) (string, error) {
	partKey := fmt.Sprintf("tus-parts/%s/part_%06d.bin", sessionID, chunkIdx)
	info, err := m.client.PutObject(
		ctx,
		bucket,
		partKey,
		bytes.NewReader(payload),
		int64(len(payload)),
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)
	if err != nil {
		return "", err
	}
	return info.ETag, nil
}

func (m *MinioStore) ComposeAndHashObject(ctx context.Context, bucket, sessionID, destKey string, totalParts int) (string, error) {
	srcs := make([]minio.CopySrcOptions, 0, totalParts)
	for i := 0; i < totalParts; i++ {
		srcs = append(srcs, minio.CopySrcOptions{
			Bucket: bucket,
			Object: fmt.Sprintf("tus-parts/%s/part_%06d.bin", sessionID, i),
		})
	}
	dst := minio.CopyDestOptions{Bucket: bucket, Object: destKey}
	if _, err := m.client.ComposeObject(ctx, dst, srcs...); err != nil {
		return "", err
	}
	obj, err := m.client.GetObject(ctx, bucket, destKey, minio.GetObjectOptions{})
	if err != nil {
		return "", err
	}
	defer obj.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, obj); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (m *MinioStore) PutWebVTT(ctx context.Context, bucket, key string, content []byte) error {
	_, err := m.client.PutObject(
		ctx,
		bucket,
		key,
		bytes.NewReader(content),
		int64(len(content)),
		minio.PutObjectOptions{ContentType: "text/vtt; charset=utf-8"},
	)
	return err
}


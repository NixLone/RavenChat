package files

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type File struct {
	ID          uuid.UUID `json:"id"`
	OwnerID     uuid.UUID `json:"owner_id"`
	ObjectKey   string    `json:"object_key"`
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	CreatedAt   time.Time `json:"created_at"`
}

type Service struct {
	db     *pgxpool.Pool
	client *minio.Client
	bucket string
}

func NewService(db *pgxpool.Pool, endpoint, access, secret, bucket string, useSSL bool) (*Service, error) {
	c, err := minio.New(endpoint, &minio.Options{Creds: credentials.NewStaticV4(access, secret, ""), Secure: useSSL})
	if err != nil {
		return nil, err
	}
	return &Service{db: db, client: c, bucket: bucket}, nil
}

func (s *Service) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
}

func (s *Service) Upload(ctx context.Context, ownerID uuid.UUID, file multipart.File, header *multipart.FileHeader, contentType string) (File, error) {
	id := uuid.New()
	objectKey := fmt.Sprintf("uploads/%s%s", id, filepath.Ext(header.Filename))
	info, err := s.client.PutObject(ctx, s.bucket, objectKey, file, header.Size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return File{}, err
	}
	stored := File{ID: id, OwnerID: ownerID, ObjectKey: objectKey, FileName: header.Filename, ContentType: contentType, SizeBytes: info.Size}
	err = s.db.QueryRow(ctx, `INSERT INTO files(id, owner_id, object_key, file_name, content_type, size_bytes)
	VALUES($1,$2,$3,$4,$5,$6) RETURNING created_at`, stored.ID, stored.OwnerID, stored.ObjectKey, stored.FileName, stored.ContentType, stored.SizeBytes).Scan(&stored.CreatedAt)
	return stored, err
}

package filesrv

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
)

type fileService struct {
	metadata *mongo.Collection
	fsBucket *gridfs.Bucket
	tempDir  string
}

func NewFileService(metaColl *mongo.Collection, fsBucket *gridfs.Bucket) FileService {
	return &fileService{
		metadata: metaColl,
		fsBucket: fsBucket,
		tempDir:  "./tmp_uploads",
	}
}

func (s *fileService) InitUpload(ctx context.Context, filename string, totalChunks, chunkSize int) (string, error) {
	sessionID := primitive.NewObjectID().Hex()

	meta := UploadMetadata{
		ID:             sessionID,
		Filename:       filename,
		TotalChunks:    totalChunks,
		UploadedChunks: []int{},
		ChunkSize:      chunkSize,
		Status:         "in_progress",
		CreatedAt:      time.Now(),
	}
	_, err := s.metadata.InsertOne(ctx, meta)
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func (s *fileService) UploadChunk(ctx context.Context, sessionID string, chunkNum int, data []byte) error {
	meta := UploadMetadata{}
	err := s.metadata.FindOne(ctx, bson.M{"_id": sessionID}).Decode(&meta)
	if err != nil {
		return err
	}

	dir := filepath.Join(s.tempDir, sessionID)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	chunkPath := filepath.Join(dir, fmt.Sprintf("%d.chunk", chunkNum))
	if err := os.WriteFile(chunkPath, data, 0600); err != nil {
		return err
	}

	_, err = s.metadata.UpdateOne(ctx,
		bson.M{"_id": sessionID},
		bson.M{"$addToSet": bson.M{"uploaded_chunks": chunkNum}},
	)

	return err
}

func (s *fileService) FinalizeUpload(ctx context.Context, sessionID string) (string, error) {
	meta := UploadMetadata{}
	err := s.metadata.FindOne(ctx, bson.M{"_id": sessionID}).Decode(&meta)
	if err != nil {
		return "", err
	}

	if len(meta.UploadedChunks) != meta.TotalChunks {
		return "", errors.New("not all chunks uploaded")
	}

	uploadStream, err := s.fsBucket.OpenUploadStream(meta.Filename)
	if err != nil {
		return "", err
	}
	defer uploadStream.Close()

	for i := range meta.TotalChunks {
		f, err := safeOpenChunk(s.tempDir, sessionID, i)
		if err != nil {
			return "", err
		}
		_, err = io.Copy(uploadStream, f)
		if cerr := f.Close(); cerr != nil {
			return "", cerr
		}
		if err != nil {
			return "", err
		}
	}

	_, err = s.metadata.UpdateOne(ctx,
		bson.M{"_id": sessionID},
		bson.M{
			"$set": bson.M{
				"status":        "completed",
				"final_file_id": uploadStream.FileID,
			},
		},
	)

	return uploadStream.FileID.(primitive.ObjectID).Hex(), err
}

func (s *fileService) AbortUpload(ctx context.Context, sessionID string) error {
	_ = os.RemoveAll(filepath.Join(s.tempDir, sessionID))
	_, err := s.metadata.UpdateOne(ctx,
		bson.M{"_id": sessionID},
		bson.M{"$set": bson.M{"status": "aborted"}},
	)
	return err
}

func (s *fileService) DownloadFile(ctx context.Context, filename string) ([]byte, error) {
	fileStream, err := s.fsBucket.OpenDownloadStreamByName(filename)
	if err != nil {
		return nil, err
	}
	defer fileStream.Close()

	data, err := io.ReadAll(fileStream)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func safeOpenChunk(baseDir, sessionID string, chunkNum int) (*os.File, error) {
	chunkFilename := fmt.Sprintf("%d.chunk", chunkNum)
	dir := filepath.Join(baseDir, filepath.Base(sessionID))
	safePath := filepath.Join(dir, filepath.Base(chunkFilename))
	// #nosec G304 -- path sanitized with filepath.Base
	return os.Open(safePath)
}

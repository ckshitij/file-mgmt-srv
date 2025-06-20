// Package filesrv provides a file upload and download service
// built on top of MongoDB GridFS. It supports chunked uploads
// with resumability, metadata tracking, and safe disk buffering.
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

// fileService implements the FileService interface and handles
// file uploads, chunk buffering, metadata storage, and downloads.
type fileService struct {
	metadata *mongo.Collection // MongoDB collection to track upload metadata
	fsBucket *gridfs.Bucket    // GridFS bucket for final file storage
	tempDir  string            // Directory to temporarily buffer chunk files
}

// NewFileService creates a new instance of fileService.
func NewFileService(metaColl *mongo.Collection, fsBucket *gridfs.Bucket) FileService {
	return &fileService{
		metadata: metaColl,
		fsBucket: fsBucket,
		tempDir:  "./tmp_uploads",
	}
}

// InitUpload initializes a new upload session by storing
// metadata such as filename, chunk size, and total chunks.
// It returns a session ID to be used for uploading chunks.
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

// UploadChunk saves an individual chunk of a file to local
// disk and updates the metadata to mark the chunk as received.
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

// FinalizeUpload assembles all uploaded chunks in order,
// streams them to GridFS, marks the upload as complete,
// and removes local chunk files from disk.
// Returns the final file's ObjectID as a hex string.
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

	go s.removedProcessedChunks(sessionID)

	return uploadStream.FileID.(primitive.ObjectID).Hex(), err
}

// AbortUpload cancels an in-progress upload and cleans up
// any locally stored chunk files. The metadata status is
// marked as "aborted".
func (s *fileService) AbortUpload(ctx context.Context, sessionID string) error {
	if err := s.removedProcessedChunks(sessionID); err != nil {
		return err
	}
	_, err := s.metadata.UpdateOne(ctx,
		bson.M{"_id": sessionID},
		bson.M{"$set": bson.M{"status": "aborted"}},
	)
	return err
}

// DownloadFile retrieves a complete file from GridFS
// using the filename and returns its content in memory.
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

// safeOpenChunk sanitizes the file path and safely opens a chunk file
// from disk. It prevents path traversal using filepath.Base.
func safeOpenChunk(baseDir, sessionID string, chunkNum int) (*os.File, error) {
	chunkFilename := fmt.Sprintf("%d.chunk", chunkNum)
	dir := filepath.Join(baseDir, filepath.Base(sessionID))
	safePath := filepath.Join(dir, filepath.Base(chunkFilename))
	// #nosec G304 -- path sanitized with filepath.Base
	return os.Open(safePath)
}

// removedProcessedChunks deletes all buffered chunk files
// for the specified upload session from the local temp directory.
func (s *fileService) removedProcessedChunks(sessionID string) error {
	return os.RemoveAll(filepath.Join(s.tempDir, sessionID))
}

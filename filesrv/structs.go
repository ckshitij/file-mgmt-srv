package filesrv

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UploadMetadata struct {
	ID             string              `bson:"_id"`
	Filename       string              `bson:"filename"`
	TotalChunks    int                 `bson:"total_chunks"`
	UploadedChunks []int               `bson:"uploaded_chunks"`
	ChunkSize      int                 `bson:"chunk_size"`
	Status         string              `bson:"status"`
	CreatedAt      time.Time           `bson:"created_at"`
	FinalFileID    *primitive.ObjectID `bson:"final_file_id,omitempty"`
}

type AbortRequest struct {
	SessionID string `json:"session_id"`
}

type FinalizeRequest struct {
	SessionID string `json:"session_id"`
}

type FinalizeResponse struct {
	FileID string `json:"file_id"`
	Err    error  `json:"err,omitempty"`
}

type UploadChunkRequest struct {
	SessionID string `json:"session_id"`
	ChunkNum  int    `json:"chunk"`
	Data      []byte
}

type InitUploadRequest struct {
	Filename    string `json:"filename"`
	TotalChunks int    `json:"total_chunks"`
	ChunkSize   int    `json:"chunk_size"`
}

type InitUploadResponse struct {
	SessionID string `json:"session_id"`
	Err       error  `json:"err,omitempty"`
}

type GenericResponse struct {
	Err error `json:"err,omitempty"`
}

// DownloadRequest represents a request to download a file by name
type DownloadRequest struct {
	Filename string `json:"filename"`
}

// DownloadResponse represents the file data and filename
type DownloadResponse struct {
	Filename string
	Data     []byte
}

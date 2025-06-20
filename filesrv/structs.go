// Package filesrv defines types used for file upload sessions,
// including metadata tracking, API request/response payloads,
// and download handling.
package filesrv

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UploadMetadata represents the state of a file upload session,
// stored in MongoDB to track progress and finalize uploads.
type UploadMetadata struct {
	ID             string              `bson:"_id"`                     // Unique session ID for the upload
	Filename       string              `bson:"filename"`                // Original file name
	TotalChunks    int                 `bson:"total_chunks"`            // Expected number of chunks
	UploadedChunks []int               `bson:"uploaded_chunks"`         // Chunks successfully uploaded
	ChunkSize      int                 `bson:"chunk_size"`              // Size of each chunk in bytes
	Status         string              `bson:"status"`                  // Upload status: in_progress, completed, or aborted
	CreatedAt      time.Time           `bson:"created_at"`              // Timestamp of session creation
	FinalFileID    *primitive.ObjectID `bson:"final_file_id,omitempty"` // ID of the final GridFS file (if completed)
}

// AbortRequest is the payload to abort an upload session.
type AbortRequest struct {
	SessionID string `json:"session_id"` // Unique session ID to abort
}

// FinalizeRequest is the payload to finalize and store a complete file
// from uploaded chunks.
type FinalizeRequest struct {
	SessionID string `json:"session_id"` // Upload session to finalize
}

// FinalizeResponse contains the result of finalizing a file upload.
type FinalizeResponse struct {
	FileID string `json:"file_id"`       // ID of the finalized GridFS file
	Err    error  `json:"err,omitempty"` // Optional error
}

// UploadChunkRequest is the request body for uploading a single chunk
// as part of a multipart file upload.
type UploadChunkRequest struct {
	SessionID string `json:"session_id"` // ID of the upload session
	ChunkNum  int    `json:"chunk"`      // Chunk number (0-based)
	Data      []byte // Raw binary data of the chunk
}

// InitUploadRequest contains parameters to start a new upload session.
type InitUploadRequest struct {
	Filename    string `json:"filename"`     // Name of the file to be uploaded
	TotalChunks int    `json:"total_chunks"` // Total number of expected chunks
	ChunkSize   int    `json:"chunk_size"`   // Size of each chunk in bytes
}

// InitUploadResponse is returned after a new upload session is created.
type InitUploadResponse struct {
	SessionID string `json:"session_id"`    // Generated session ID
	Err       error  `json:"err,omitempty"` // Optional error
}

// GenericResponse is a common response structure for APIs that return
// no additional data, just success or error.
type GenericResponse struct {
	Err error `json:"err,omitempty"` // Optional error
}

// DownloadRequest represents a request to download a file by name.
type DownloadRequest struct {
	Filename string `json:"filename"` // Name of the file to retrieve
}

// DownloadResponse represents the response to a file download request,
// containing the file data and its name.
type DownloadResponse struct {
	Filename string // Original file name
	Data     []byte // Raw file content
}

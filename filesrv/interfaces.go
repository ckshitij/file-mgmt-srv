// Package filesrv defines the core FileService interface and error handling
// contract for chunked file upload and download services.
package filesrv

import "context"

// errorer is an interface used for transport-level error propagation.
// Implementations can return a concrete error via Err(), enabling Go-Kit
// transport layers to inspect and return appropriate HTTP errors.
type errorer interface {
	// Err returns a non-nil error if one occurred during service execution.
	Err() error
}

// FileService defines the operations supported by the file upload and
// download system, including support for resumable, chunked uploads.
type FileService interface {
	// InitUpload starts a new upload session by recording file metadata
	// and returns a unique session ID.
	//
	// filename     - the name of the file to be uploaded
	// totalChunks  - the total number of chunks expected
	// chunkSize    - the size in bytes of each chunk
	InitUpload(ctx context.Context, filename string, totalChunks int, chunkSize int) (string, error)

	// UploadChunk stores a chunk of the file associated with a session ID.
	//
	// sessionID - ID of the upload session
	// chunkNum  - the index of this chunk (0-based)
	// data      - raw binary data of the chunk
	UploadChunk(ctx context.Context, sessionID string, chunkNum int, data []byte) error

	// FinalizeUpload assembles all chunks for a session into a complete file
	// and stores it in GridFS. Returns the ID of the stored file.
	//
	// sessionID - ID of the upload session
	FinalizeUpload(ctx context.Context, sessionID string) (string, error)

	// AbortUpload cancels the upload session and deletes any stored chunks.
	//
	// sessionID - ID of the upload session
	AbortUpload(ctx context.Context, sessionID string) error

	// DownloadFile retrieves a complete file by its name from GridFS.
	//
	// filename - the original name of the file to download
	DownloadFile(ctx context.Context, filename string) ([]byte, error)
}

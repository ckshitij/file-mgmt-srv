package filesrv

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-kit/kit/transport"
	kitHttp "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
)

func MakeHTTPHandler(e Endpoints, logger log.Logger) http.Handler {
	mux := http.NewServeMux()

	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	mux.Handle("/init-upload", kitHttp.NewServer(
		e.InitUpload,
		decodeInitUploadRequest,
		encodeResponse,
		options...,
	))

	mux.Handle("/upload-chunk", kitHttp.NewServer(
		e.UploadChunk,
		decodeUploadChunkRequest,
		encodeResponse,
		options...,
	))

	mux.Handle("/finalize-upload", kitHttp.NewServer(
		e.FinalizeUpload,
		decodeFinalizeRequest,
		encodeResponse,
		options...,
	))

	mux.Handle("/abort-upload", kitHttp.NewServer(
		e.AbortUpload,
		decodeAbortRequest,
		encodeResponse,
		options...,
	))

	mux.Handle("/download", kitHttp.NewServer(
		e.Download,
		decodeDownloadRequest,
		encodeDownloadResponse,
		options...,
	))

	// ✅ Register HTML UI route on correct mux
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./index.html")
	})

	return mux
}

func decodeInitUploadRequest(_ context.Context, r *http.Request) (any, error) {
	var req InitUploadRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeUploadChunkRequest(_ context.Context, r *http.Request) (any, error) {
	sessionID := r.URL.Query().Get("session_id")
	chunkNum, _ := strconv.Atoi(r.URL.Query().Get("chunk"))

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return UploadChunkRequest{
		SessionID: sessionID,
		ChunkNum:  chunkNum,
		Data:      data,
	}, nil
}

func decodeFinalizeRequest(_ context.Context, r *http.Request) (any, error) {
	return FinalizeRequest{
		SessionID: r.URL.Query().Get("session_id"),
	}, nil
}

func decodeAbortRequest(_ context.Context, r *http.Request) (any, error) {
	return AbortRequest{
		SessionID: r.URL.Query().Get("session_id"),
	}, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response any) error {
	if errResp, ok := response.(errorer); ok && errResp.Err() != nil {
		http.Error(w, errResp.Err().Error(), http.StatusInternalServerError)
		return nil
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}

func decodeDownloadRequest(_ context.Context, r *http.Request) (any, error) {
	filename := r.URL.Query().Get("filename")
	return DownloadRequest{Filename: filename}, nil
}

func encodeDownloadResponse(ctx context.Context, w http.ResponseWriter, response any) error {
	resp, ok := response.(DownloadResponse)
	if !ok {
		http.Error(w, "invalid response type", http.StatusInternalServerError)
		return nil
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\""+resp.Filename+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	_, err := w.Write(resp.Data)
	return err
}

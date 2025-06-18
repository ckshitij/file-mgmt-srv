package filesrv

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	InitUpload     endpoint.Endpoint
	UploadChunk    endpoint.Endpoint
	FinalizeUpload endpoint.Endpoint
	AbortUpload    endpoint.Endpoint
	Download       endpoint.Endpoint
}

func MakeEndpoints(svc FileService) Endpoints {
	return Endpoints{
		InitUpload:     InitUploadEndpoint(svc),
		UploadChunk:    UploadChunkEndpoint(svc),
		FinalizeUpload: FinalizeEndpoint(svc),
		AbortUpload:    AbortEndpoint(svc),
		Download:       DownloadEndpoint(svc),
	}
}

func InitUploadEndpoint(svc FileService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(InitUploadRequest)
		fmt.Printf("request init %+v\n", req)
		id, err := svc.InitUpload(ctx, req.Filename, req.TotalChunks, req.ChunkSize)
		return InitUploadResponse{SessionID: id, Err: err}, err
	}
}

func UploadChunkEndpoint(svc FileService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(UploadChunkRequest)
		err := svc.UploadChunk(ctx, req.SessionID, req.ChunkNum, req.Data)
		return GenericResponse{Err: err}, err
	}
}

func FinalizeEndpoint(svc FileService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(FinalizeRequest)
		id, err := svc.FinalizeUpload(ctx, req.SessionID)
		return FinalizeResponse{FileID: id, Err: err}, err
	}
}

func AbortEndpoint(svc FileService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(AbortRequest)
		err := svc.AbortUpload(ctx, req.SessionID)
		return GenericResponse{Err: err}, err
	}
}

func DownloadEndpoint(svc FileService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DownloadRequest)
		data, err := svc.DownloadFile(ctx, req.Filename)
		if err != nil {
			return nil, err
		}
		return DownloadResponse{Filename: req.Filename, Data: data}, nil
	}
}

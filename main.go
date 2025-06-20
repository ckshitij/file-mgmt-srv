package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ckshitij/file-mgmt-srv/config"
	dbmongo "github.com/ckshitij/file-mgmt-srv/db-mongo"
	"github.com/ckshitij/file-mgmt-srv/filesrv"
	logkit "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	CollectionName  string = "uploads"
	GridFSChunkSize int32  = 8 * 1024 * 1024 // 8MB
)

func main() {

	cfg, err := config.LoadConfig("resource/config.yml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()
	uri := cfg.MongoDB.URI
	client, err := dbmongo.NewMongoDBClient(ctx, uri)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx)
	if err != nil {
		log.Fatal(err)
	}

	db := client.GetDatabase("upload_service")
	fsBucket, _ := gridfs.NewBucket(db, &options.BucketOptions{
		Name:           &CollectionName,
		ChunkSizeBytes: &GridFSChunkSize,
	})
	uploadsCollection := db.Collection(CollectionName)

	var logger logkit.Logger
	{
		logger = logkit.NewLogfmtLogger(os.Stderr)
		logger = logkit.With(logger, "ts", logkit.DefaultTimestampUTC)
		logger = logkit.With(logger, "caller", logkit.DefaultCaller)
	}

	var svc filesrv.FileService
	{
		svc = filesrv.NewFileService(uploadsCollection, fsBucket)
		svc = filesrv.LoggingMiddleware(logger)(svc)
	}

	endpoints := filesrv.MakeEndpoints(svc)

	var handler http.Handler
	{
		handler = filesrv.MakeHTTPHandler(endpoints, logkit.With(logger, "component", "HTTP"))
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 300 * time.Second,
	}

	level.Info(logger).Log("msg", "Server listening on :", cfg.Server.Port)
	log.Fatal(srv.ListenAndServe())
}

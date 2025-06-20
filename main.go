package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func(handler http.Handler) {
		httpAddress := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
		srv := &http.Server{
			Addr:         httpAddress,
			Handler:      handler,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 300 * time.Second,
		}
		if err := level.Info(logger).Log("msg", "server starting to listen", "addr", httpAddress); err != nil {
			fmt.Println("error will logging server start listening")
		}
		errs <- srv.ListenAndServe()
	}(handler)

	if err := logger.Log("exit", <-errs); err != nil {
		fmt.Println("error will logging server exit error")
	}
}

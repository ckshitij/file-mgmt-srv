package main

import (
	"context"
	"log"
	"net/http"
	"time"

	dbmongo "github.com/ckshitij/media-mgmt-srv/db-mongo"
	"github.com/ckshitij/media-mgmt-srv/filesrv"
	logkit "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
)

func main() {
	ctx := context.Background()
	uri := ""
	client, err := dbmongo.NewMongoDBClient(ctx, uri)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx)
	if err != nil {
		log.Fatal(err)
	}

	db := client.GetDatabase("upload_service")
	fsBucket, _ := gridfs.NewBucket(db)
	metadataColl := db.Collection("uploads")

	logger := logkit.NewLogfmtLogger(logkit.NewSyncWriter(log.Writer()))

	svc := filesrv.NewFileService(metadataColl, fsBucket)
	endpoints := filesrv.MakeEndpoints(svc)
	handler := filesrv.MakeHTTPHandler(endpoints, logger)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 300 * time.Second,
	}

	level.Info(logger).Log("msg", "Server listening on :8080")
	log.Fatal(srv.ListenAndServe())
}

package main

import (
	"context"
	"log"

	dbmongo "github.com/ckshitij/media-mgmt-srv/db-mongo"
)

func main() {
	ctx := context.Background()
	client, err := dbmongo.NewMongoDBClient(ctx, "uri")
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx)
	if err != nil {
		log.Fatal(err)
	}

	client.GetDatabase("upload_service")

}

package main

import (
	"context"
	"flag"
	"log"

	iodb "github.com/godyy/ggs/internal/core/db/io"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	mgoUri := flag.String("mongo-uri", "", "mongo uri")
	flag.Parse()

	if *mgoUri == "" {
		flag.Usage()
		return
	}

	cli, err := mongo.Connect(options.Client().ApplyURI(*mgoUri))
	if err != nil {
		log.Fatalf("connect mongo failed, err: %v", err)
	}
	defer cli.Disconnect(context.Background())

	if err := iodb.InitMongoIndexes(context.Background(), cli); err != nil {
		log.Fatalf("init mongo indexes failed, err: %v", err)
	}
}

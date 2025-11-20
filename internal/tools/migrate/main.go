package main

import (
	"context"
	"flag"
	"log"

	"github.com/godyy/ggs/internal/data/migrate"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	mgoUri := flag.String("mongo-uri", "", "mongo uri")
	flag.Parse()

	if *mgoUri != "" {
		migrateMongo(*mgoUri)
	}
}

func migrateMongo(uri string) {
	cli, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("connect mongo failed, err: %v", err)
	}
	defer cli.Disconnect(context.Background())

	if err := migrate.Mongo(context.Background(), cli); err != nil {
		log.Fatalf("migrate mongo failed, err: %v", err)
	}
}

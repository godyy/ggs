package main

import (
	"flag"
	"log"

	"github.com/godyy/gexcels"
	"github.com/godyy/gexcels/export"
	exportcode "github.com/godyy/gexcels/export/code"
	exportdata "github.com/godyy/gexcels/export/data"
	"github.com/godyy/gexcels/parse"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

const (
	packageName = "gdconf"
	tag         = gexcels.Tag("s")
)

func main() {
	excelPath := flag.String("excel-path", "", "excel path")
	codePath := flag.String("code-path", "", "code path")
	mongoURI := flag.String("mongo-uri", "", "mongo uri")
	mongoDB := flag.String("mongo-db", "", "mongo db")
	flag.Parse()

	if *excelPath == "" {
		log.Fatalf("-excel-path is required")
	}
	if *codePath == "" {
		log.Fatalf("-code-path is required")
	}
	if *mongoURI == "" {
		log.Fatalf("-mongo-uri is required")
	}
	if *mongoDB == "" {
		log.Fatalf("-mongo-db is required")
	}

	mongoCliOpts := options.Client().
		ApplyURI(*mongoURI).
		SetWriteConcern(writeconcern.Majority())
	mongoCli, err := mongo.Connect(mongoCliOpts)
	if err != nil {
		log.Fatalf("connect mongo at %s failed: %v", *mongoURI, err)
	}
	mgoDB := mongoCli.Database(*mongoDB)

	parser, err := parse.Parse(*excelPath, &parse.Options{Tags: []gexcels.Tag{tag}})
	if err != nil {
		log.Fatalf("parse excel at %s failed: %v", *excelPath, err)
	}

	if err := exportcode.ExportGo(parser, *codePath, &exportcode.Options{DataKind: export.DataBson}, &exportcode.GoOptions{PkgName: packageName}); err != nil {
		log.Fatalf("export code to %s failed: %v", *codePath, err)
	}

	if err := exportdata.ExportBson(parser, mgoDB); err != nil {
		log.Fatalf("export bson to mongo [%s][%s] failed: %v", *mongoURI, *mongoDB, err)
	}
}

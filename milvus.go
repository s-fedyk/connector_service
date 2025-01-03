package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

var milvusClient *client.Client

func init() {
	log.Print("Initializing Milvus...")

	milvusURL := os.Getenv("MILVUS_URL")
	if milvusURL == "" {
		log.Fatal("MILVUS_URL environment variable is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := client.NewClient(ctx,
		client.Config{
			Address: milvusURL,
			DBName:  "default",
		},
	)

	if err != nil {
		log.Fatalf("Milvus connection failed. err=(%v)", err)
	}

	milvusClient = &client

	log.Print("Milvus initialization success!")

	databasePresent, err := collectionPresent()

	if databasePresent {
		log.Print("Collection present!")
	} else {
		log.Fatal("Database empty!")
	}

	if err != nil {
		log.Fatalf("Error retrieving collection status, err=(%v)", err)
	}
}

func collectionPresent() (bool, error) {
	return (*milvusClient).HasCollection(context.Background(), "image_embeddings")
}

func querySimilar(embedding []float32, context context.Context) {
	log.Printf("querySimilar")

	sp, _ := entity.NewIndexFlatSearchParam()

	res, err := (*milvusClient).Search(
		context,
		"image_embeddings",
		[]string{},
		"",
		[]string{"filename"},
		[]entity.Vector{entity.FloatVector(embedding)},
		"vector",
		entity.L2,
		10,
		sp,
	)

	if err != nil {
		log.Printf("Search error, err=(%v)", err)
	}

	log.Printf("similar filename is %v", res)

	return
}

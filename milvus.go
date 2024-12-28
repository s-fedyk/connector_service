package main

import (
	"context"
	"log"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

var milvusClient *client.Client

func init() {
	log.Print("Initializing Milvus...")

	client, err := client.NewClient(context.Background(),
		client.Config{
			Address: "milvus-demo.milvus.svc.cluster.local:19530",
		},
	)

	milvusClient = &client

	if err != nil {
		log.Fatalf("Milvus connection failed. err=(%v)", err)
	}

	log.Print("Milvus initialization success!")
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
		"Vector",
		entity.COSINE,
		10,
		sp,
	)

	if err != nil {
		log.Printf("Search error, err=(%v)", err)
	}

	log.Printf("similar filename is %v", res)

	return
}

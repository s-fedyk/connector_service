package main

import (
	"context"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"log"
)

var milvusClient client.Client

func init() {
	log.Print("Initializing Milvus...")

	var err error = nil
	milvusClient, err = client.NewClient(context.Background(),
		client.Config{
			Address: "milvus-demo.milvus.svc.cluster.local:19530",
		},
	)

	if err != nil {
		log.Fatalf("Milvus connection failed. err=(%v)", err)
	}

	log.Print("Milvus initialization success!")
}

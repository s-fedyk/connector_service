package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

var milvusClient *client.Client

type milvusSecret struct {
	MILVUS_URI  string
	MILVUS_USER string
	MILVUS_PASS string
}

func getAWSSecret() milvusSecret {
	secretName := "prod/similarity/milvus"
	region := "us-east-2"

	config, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		log.Fatal(err)
	}

	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := svc.GetSecretValue(context.Background(), input)
	if err != nil {
		log.Fatal(err.Error())
	}

	var secret milvusSecret
	err = json.Unmarshal([]byte(*result.SecretString), &secret)
	if err != nil {
		log.Fatal(err.Error())
	}
	return secret
}

func init() {
	log.Print("Initializing Milvus...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var secret milvusSecret

	if os.Getenv("ENV") == "DEV" {
		secret.MILVUS_URI = os.Getenv("MILVUS_URI")
		secret.MILVUS_USER = os.Getenv("MILVUS_USER")
		secret.MILVUS_PASS = os.Getenv("MILVUS_PASS")
	} else {
		secret = getAWSSecret()
	}

	client, err := client.NewClient(ctx,
		client.Config{
			Address:  secret.MILVUS_URI,
			Username: secret.MILVUS_USER,
			Password: secret.MILVUS_PASS,
			DBName:   "default",
		},
	)

	if err != nil {
		log.Fatalf("Milvus connection failed. err=(%v)", err)
	}

	milvusClient = &client

	log.Print("Milvus initialization success!")

	databasePresent, err := collectionPresent()

	if err != nil {
		log.Fatalf("Error retrieving collection status, err=(%v)", err)
	}

	if databasePresent {
		log.Print("Collection present!")
	} else {
		log.Fatal("Database empty!")
	}
}

func collectionPresent() (bool, error) {
	return (*milvusClient).HasCollection(context.Background(), "facenet_embeddings")
}

func querySimilar(embedding []float32, context context.Context) []string {
	log.Printf("querySimilar")
	log.Printf("%v", embedding)

	sp, _ := entity.NewIndexFlatSearchParam()

	res, err := (*milvusClient).Search(
		context,
		"facenet_embeddings",
		[]string{},
		"",
		[]string{"filepath"},
		[]entity.Vector{entity.FloatVector(embedding)},
		"vector",
		entity.L2,
		10,
		sp,
	)

	if err != nil {
		log.Printf("Search error, err=(%v)", err)
	}

	URLs := make([]string, 10)

	for _, searchResult := range res {
		for idx := range 10 {
			URL, _ := searchResult.Fields.GetColumn("filepath").GetAsString(idx)
			URLs[idx] = URL
		}
	}

	log.Printf("querySimilar success")

	return URLs
}

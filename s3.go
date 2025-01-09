package main

import (
	"bytes"
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	s3Client   *s3.S3
	bucketName string
	accessKey  string
)

func init() {
	log.Print("Initializing S3...")

	// Load AWS configuration
	bucketName = os.Getenv("S3_BUCKET")
	if bucketName == "" {
		log.Fatal("S3_BUCKET environment variable is not set")
	}

	awsRegion := "us-east-2"

	accessKey := os.Getenv("S3_ACCESS_KEY")
	accessSecret := os.Getenv("S3_ACCESS_SECRET")

	var sess *session.Session
	var err error

	if os.Getenv("ENV") == "DEV" {
		// Initialize S3 client
		sess, err = session.NewSession(&aws.Config{
			Region:      aws.String(awsRegion),
			Credentials: credentials.NewStaticCredentials(accessKey, accessSecret, ""),
		})
	} else {
		sess, err = session.NewSession(&aws.Config{
			Region: aws.String(awsRegion),
		})
	}

	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err)
	}

	s3Client = s3.New(sess)
	log.Println("S3 client initialized successfully!")
}

func storeS3(buf []byte, filename string) (bool, error) {
	ctx := context.Background()

	_, err := s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(filename),
		Body:          bytes.NewReader(buf),
		ContentLength: aws.Int64(int64(len(buf))),
		ContentType:   aws.String("image/jpeg"),
	})

	if err != nil {
		log.Printf("Failed to upload to S3: %v", err)
		return false, err
	}

	log.Printf("Successfully uploaded %s to S3 bucket %s", filename, bucketName)
	return true, nil
}

func deleteS3(filename string) (bool, error) {
	ctx := context.Background()

	_, err := s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
	})
	if err != nil {
		log.Printf("Failed to delete %s from S3: %v", filename, err)
		return false, err
	}

	// Wait until the object is actually deleted. This step is optional but ensures
	// the delete operation is fully confirmed.
	err = s3Client.WaitUntilObjectNotExistsWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
	})
	if err != nil {
		log.Printf("Error while waiting for object %s to be deleted: %v", filename, err)
		return false, err
	}

	log.Printf("Successfully deleted %s from S3 bucket %s", filename, bucketName)
	return true, nil
}

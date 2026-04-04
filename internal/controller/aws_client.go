package controller

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// sqsClient creates a new AWS SQS client for the given region.
// Credentials are read from the AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
// environment variables, which are injected into the pod via a Kubernetes Secret.
func sqsClient(region string) *sqs.Client {
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		),
	)
	if err != nil {
		fmt.Println("Error loading AWS config:", err)
		os.Exit(1)
	}

	return sqs.NewFromConfig(cfg)
}

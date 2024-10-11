package kms

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
)

func NewKMSClient() (*kms.Client, error) {
	// Create a new AWS session with explicit credentials and region
	awsCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		fmt.Println("Error loading AWS configuration:", err)
		return &kms.Client{}, err
	}
	kmsClient := kms.NewFromConfig(awsCfg)
	fmt.Println("KMS client created")
	return kmsClient, nil
}

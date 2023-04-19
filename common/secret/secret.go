package secret

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"log"
	"os"
)

const (
	DefaultVersionStage = "AWSCURRENT"
	AwsRegion           = "AWS_REGION"
)

func LoadSecretData(secretName string) (map[string]string, error) {

	awsRegion := os.Getenv(AwsRegion)
	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	if err != nil {
		log.Fatal(err)
	}

	secretManager := secretsmanager.NewFromConfig(config)
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String(DefaultVersionStage),
	}

	result, err := secretManager.GetSecretValue(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	resultMap := make(map[string]string)
	err = json.Unmarshal(result.SecretBinary, &resultMap)
	if err != nil {
		return nil, err
	}

	return resultMap, nil
}

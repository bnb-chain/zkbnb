package secret

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"os"
)

const (
	VersionStage     = "AWSCURRENT"
	AwsRegion        = "AWS_REGION"
	AwsProfile       = "AWS_PROFILE"
	AwsSdkLoadConfig = "AWS_SDK_LOAD_CONFIG"
)

func LoadSecretData(secretName string) (map[string]string, error) {

	awsRegion := os.Getenv(AwsRegion)
	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	if err != nil {
		return nil, err
	}

	secretManager := secretsmanager.NewFromConfig(config)
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String(VersionStage),
	}

	result, err := secretManager.GetSecretValue(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	resultMap := make(map[string]string)
	secretBytes := []byte(*result.SecretString)
	err = json.Unmarshal(secretBytes, &resultMap)

	if err != nil {
		return nil, err
	}

	return resultMap, nil
}

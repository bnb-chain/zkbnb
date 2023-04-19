package secret

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/zeromicro/go-zero/core/logx"
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
	secretBytes := []byte(*result.SecretString)
	err = json.Unmarshal(secretBytes, &resultMap)
	if err != nil {

		logx.Errorf("result.SecretString:%s", *result.SecretString)

		return nil, err
	}

	return resultMap, nil
}

package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/chime"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AWSConfig struct {
	Region    string
	ChimeSDK  *chime.Client
	DynamoDB  *dynamodb.Client
	S3        *s3.Client
	Lambda    *lambda.Client
	Cache     *elasticache.Client
}

func NewAWSConfig(ctx context.Context, region string) (*AWSConfig, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	return &AWSConfig{
		Region:    region,
		ChimeSDK:  chime.NewFromConfig(cfg),
		DynamoDB:  dynamodb.NewFromConfig(cfg),
		S3:        s3.NewFromConfig(cfg),
		Lambda:    lambda.NewFromConfig(cfg),
		Cache:     elasticache.NewFromConfig(cfg),
	}, nil
}

package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// SoloGame represents a practice game in DynamoDB
type SoloGame struct {
	ID            string    `dynamodbav:"id"`
	UserID        string    `dynamodbav:"user_id"`
	Status        string    `dynamodbav:"status"`
	WordID        string    `dynamodbav:"word_id"`
	Word          string    `dynamodbav:"word"`
	Attempts      []Attempt `dynamodbav:"attempts"`
	HintsUsed     int       `dynamodbav:"hints_used"`
	Score         int       `dynamodbav:"score"`
	StartedAt     time.Time `dynamodbav:"started_at"`
	CompletedAt   time.Time `dynamodbav:"completed_at,omitempty"`
	CreatedAt     time.Time `dynamodbav:"created_at"`
}

type Attempt struct {
	Word      string    `dynamodbav:"word"`
	Type      string    `dynamodbav:"type"` // voice or text
	IsCorrect bool      `dynamodbav:"is_correct"`
	Timestamp time.Time `dynamodbav:"timestamp"`
}

// UserWordStats tracks a user's performance with specific words
type UserWordStats struct {
	UserID           string    `dynamodbav:"user_id"`
	WordID           string    `dynamodbav:"word_id"`
	CorrectAttempts  int       `dynamodbav:"correct_attempts"`
	IncorrectAttempts int      `dynamodbav:"incorrect_attempts"`
	LastAttemptAt    time.Time `dynamodbav:"last_attempt_at"`
	NextReviewAt     time.Time `dynamodbav:"next_review_at"`
}

type DynamoDBService struct {
	client *dynamodb.Client
}

func NewDynamoDBService(client *dynamodb.Client) *DynamoDBService {
	return &DynamoDBService{client: client}
}

// CreateTables creates the required DynamoDB tables
func (s *DynamoDBService) CreateTables(ctx context.Context) error {
	tables := []struct {
		name       string
		attributes []types.AttributeDefinition
		keySchema  []types.KeySchemaElement
		gsi       []types.GlobalSecondaryIndex
	}{
		{
			name: "solo_games",
			attributes: []types.AttributeDefinition{
				{
					AttributeName: aws.String("id"),
					AttributeType: types.ScalarAttributeTypeS,
				},
				{
					AttributeName: aws.String("user_id"),
					AttributeType: types.ScalarAttributeTypeS,
				},
				{
					AttributeName: aws.String("created_at"),
					AttributeType: types.ScalarAttributeTypeS,
				},
			},
			keySchema: []types.KeySchemaElement{
				{
					AttributeName: aws.String("id"),
					KeyType:      types.KeyTypeHash,
				},
			},
			gsi: []types.GlobalSecondaryIndex{
				{
					IndexName: aws.String("user_games"),
					KeySchema: []types.KeySchemaElement{
						{
							AttributeName: aws.String("user_id"),
							KeyType:      types.KeyTypeHash,
						},
						{
							AttributeName: aws.String("created_at"),
							KeyType:      types.KeyTypeRange,
						},
					},
					Projection: &types.Projection{
						ProjectionType: types.ProjectionTypeAll,
					},
				},
			},
		},
		{
			name: "user_word_stats",
			attributes: []types.AttributeDefinition{
				{
					AttributeName: aws.String("user_id"),
					AttributeType: types.ScalarAttributeTypeS,
				},
				{
					AttributeName: aws.String("word_id"),
					AttributeType: types.ScalarAttributeTypeS,
				},
				{
					AttributeName: aws.String("next_review_at"),
					AttributeType: types.ScalarAttributeTypeS,
				},
			},
			keySchema: []types.KeySchemaElement{
				{
					AttributeName: aws.String("user_id"),
					KeyType:      types.KeyTypeHash,
				},
				{
					AttributeName: aws.String("word_id"),
					KeyType:      types.KeyTypeRange,
				},
			},
			gsi: []types.GlobalSecondaryIndex{
				{
					IndexName: aws.String("review_schedule"),
					KeySchema: []types.KeySchemaElement{
						{
							AttributeName: aws.String("user_id"),
							KeyType:      types.KeyTypeHash,
						},
						{
							AttributeName: aws.String("next_review_at"),
							KeyType:      types.KeyTypeRange,
						},
					},
					Projection: &types.Projection{
						ProjectionType: types.ProjectionTypeAll,
					},
				},
			},
		},
	}

	for _, table := range tables {
		_, err := s.client.CreateTable(ctx, &dynamodb.CreateTableInput{
			TableName:            aws.String(table.name),
			AttributeDefinitions: table.attributes,
			KeySchema:           table.keySchema,
			GlobalSecondaryIndexes: table.gsi,
			BillingMode:         types.BillingModePayPerRequest,
		})
		if err != nil {
			return fmt.Errorf("failed to create table %s: %w", table.name, err)
		}
	}

	return nil
}

func (s *DynamoDBService) getTableSchema() *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("user_id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:      types.KeyTypeHash,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("UserIDIndex"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("user_id"),
						KeyType:      types.KeyTypeHash,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
		BillingMode: types.BillingModeProvisioned,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
		TableName: aws.String("solo_games"),
	}
}

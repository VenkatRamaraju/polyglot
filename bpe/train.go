// byte pair encoding package
package bpe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// creates config object
func CreateAWSConfigFromEnv(region string) (aws.Config, error) {
	// get environment variables
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	// validations
	if accessKey == "" || secretKey == "" {
		return aws.Config{}, fmt.Errorf("missing AWS credentials in environment")
	}

	// establish configuration
	creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""))
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// return configuration
	return cfg, nil
}

// get all keys in an s3 bucket
func listS3Keys(bucket, region string) ([]string, error) {
	// get configuration
	cfg, err := CreateAWSConfigFromEnv(region)
	if err != nil {
		return nil, err
	}

	// get client
	client := s3.NewFromConfig(cfg)

	// track variables
	var keys []string
	var continuationToken *string = nil
	for {
		// service call
		resp, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}
		for _, obj := range resp.Contents {
			keys = append(keys, aws.ToString(obj.Key))
		}

		// end
		if !*resp.IsTruncated {
			break
		}
		continuationToken = resp.NextContinuationToken
	}

	// return keys
	return keys, nil
}

// converts s3 files to map
func fetchJSONFromS3(bucket string, key string) (map[string]interface{}, error) {
	// context
	ctx := context.Background()

	// state
	region := "us-east-1"

	// Read credentials from environment variables
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	// validation
	if accessKey == "" || secretKey == "" || region == "" {
		return nil, fmt.Errorf("missing AWS credentials or region in environment variables")
	}

	// set up custom AWS config with static credentials
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// create S3 client
	client := s3.NewFromConfig(cfg)

	// fetch object
	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer resp.Body.Close()

	// read and unmarshal JSON
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return result, nil
}

// Orchestrate the trianing process
func Train() error {
	// populate merges
	maxToken, err := populateMerges()
	if err != nil {
		return fmt.Errorf("error populating the merges: %w", err)
	}

	// perform merges on the statistics
	fmt.Println(maxToken)

	return nil
}

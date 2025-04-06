// byte pair encoding package
package bpe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type pair struct {
	iFirst  int
	iSecond int
}

// fetchJSONFromS3 pulls a JSON file from S3 and unmarshals it into a map.
func fetchJSONFromS3(bucket, key string) (map[string]interface{}, error) {
	// grab background context
	ctx := context.Background()

	// Read credentials from environment variables
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := "us-east-1"

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

func merge(sentences []string, merges *map[pair]int) error {
	for _, sentence := range sentences {
		fmt.Println(sentence)
		os.Exit(1)
	}

	return nil
}

// Orchestrate the trianing process
func Train() error {
	// get training dataset
	languageToSentence, err := fetchJSONFromS3("tknzr", "raw.json")
	if err != nil {
		return fmt.Errorf("failed fetching JSON from S3: %w", err)
	}

	// Initialize the map to store training merges
	merges := make(map[pair]int)

	// iterate over all languages
	for language := range languageToSentence {
		if language != "English" {
			continue
		}

		// grab sentence list
		sentences, valid := languageToSentence[language].([]string)
		if !valid {
			return errors.New("Unable to parse JSON for " + language + " into a string array")
		}

		// create merges
		err = merge(sentences, &merges)
		if err != nil {
			return fmt.Errorf("failed running the merging algorithm: %w", err)
		}
	}

	return nil
}

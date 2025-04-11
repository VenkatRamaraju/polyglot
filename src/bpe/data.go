package bpe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// CreateAWSConfigFromEnv creates an AWS config object from environment variables
func CreateAWSConfigFromEnv(region string) (aws.Config, error) {
	// Get environment variables
	sAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	sSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	// Validate credentials
	if sAccessKey == "" || sSecretKey == "" {
		return aws.Config{}, fmt.Errorf("missing AWS credentials in environment")
	}

	// Establish configuration
	pdCredentials := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(sAccessKey, sSecretKey, ""))
	pdConfiguration, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(pdCredentials),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Return configuration
	return pdConfiguration, nil
}

// listS3Keys retrieves all keys in an S3 bucket
func listS3Keys(sBucket string, sRegion string) ([]string, error) {
	// Get configuration
	pdConfiguration, err := CreateAWSConfigFromEnv(sRegion)
	if err != nil {
		return nil, err
	}

	// Get S3 client
	dataClient := s3.NewFromConfig(pdConfiguration)

	// Track variables
	var asKeys []string
	var pdContinuationToken *string = nil
	for {
		// Service call
		pdResponse, err := dataClient.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket:            aws.String(sBucket),
			ContinuationToken: pdContinuationToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}
		for _, dataContents := range pdResponse.Contents {
			asKeys = append(asKeys, aws.ToString(dataContents.Key))
		}

		// End
		if !*pdResponse.IsTruncated {
			break
		}
		pdContinuationToken = pdResponse.NextContinuationToken
	}

	// Return keys
	return asKeys, nil
}

// fetchJSONFromS3 retrieves a JSON object from S3 and unmarshals it into a map
func fetchJSONFromS3(sBucket string, sKey string) (map[string]interface{}, error) {
	// Context
	dataContext := context.Background()

	// State
	sRegion := "us-east-1"

	// Read credentials from environment variables
	sAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	sSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	// Validation
	if sAccessKey == "" || sSecretKey == "" || sRegion == "" {
		return nil, fmt.Errorf("missing AWS credentials or region in environment variables")
	}

	// Set up custom AWS config with static credentials
	pdConfiguration, err := config.LoadDefaultConfig(dataContext,
		config.WithRegion(sRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(sAccessKey, sSecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	pdClient := s3.NewFromConfig(pdConfiguration)

	// Fetch object
	dataResponse, err := pdClient.GetObject(dataContext, &s3.GetObjectInput{
		Bucket: aws.String(sBucket),
		Key:    aws.String(sKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer dataResponse.Body.Close()

	// Read and unmarshal JSON
	abBody, err := io.ReadAll(dataResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	// unmarshal into map
	var mapResult map[string]interface{}
	if err := json.Unmarshal(abBody, &mapResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return mapResult, nil
}

// getData retrieves all sentences from S3
func getData() (*dataDataset, error) {
	dataDataset := &dataDataset{pdMutex: &sync.Mutex{}}

	// Get files
	asJSONFiles, err := listS3Keys("tknzr", "us-east-1")
	if err != nil {
		return nil, fmt.Errorf("unable to pull s3 files: %w", err)
	}

	// Get training dataset - we can populate map in parallel by maintaining a mutex
	var dWg sync.WaitGroup
	ch := make(chan error)
	for _, sJSONFile := range asJSONFiles {
		dWg.Add(1)
		go func(qsFileName string) {
			// Defer completion of routine
			defer dWg.Done()

			// Get the file contents
			mapLanguageToSentence, err := fetchJSONFromS3("tknzr", qsFileName)
			if err != nil {
				ch <- err
				return
			}

			// Iterate over all languages
			for sLanguage := range mapLanguageToSentence {
				// Grab sentence list
				dLanguages, valid := mapLanguageToSentence[sLanguage].([]interface{})
				if !valid {
					ch <- errors.New("Unable to parse JSON for " + sLanguage + " into a string array")
				}

				// Add to list
				dataDataset.AddList(dLanguages)
			}
		}(sJSONFile)
	}
	go func() {
		dWg.Wait()
		close(ch)
	}()

	// Collect errors
	for err := range ch {
		if err != nil {
			return nil, err
		}
	}

	return dataDataset, nil
}

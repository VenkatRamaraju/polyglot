package bpe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"normalize"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Dataset struct {
	sentences [][]int64
	mutex     sync.Mutex
}

type Statistics struct {
	pairFrequency map[[2]int64]int
	mutex         sync.Mutex
	maxPair       *[2]int64
	maxCount      *int
}

// insertMerge increments the count for a given pair in StatisticsMap
// also tracks the most frequently occurring pair
func insertMerge(statistics *Statistics, pair [2]int64) {
	statistics.mutex.Lock()
	defer statistics.mutex.Unlock()

	// Increment pair
	statistics.pairFrequency[pair]++

	// update max pair
	if statistics.pairFrequency[pair] > *statistics.maxCount {
		*statistics.maxPair = pair
		*statistics.maxCount = statistics.pairFrequency[pair]
	}
}

func (d *Dataset) add(item []int64) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.sentences = append(d.sentences, item)
}

func (d *Dataset) AddList(sentences []interface{}) {
	for index := range sentences {
		// Normalize the sentence
		sentence := normalize.Normalize(sentences[index].(string))

		// Convert to unicode integers
		var unicodePoints []int64
		for _, r := range sentence {
			unicodePoints = append(unicodePoints, int64(r))
		}

		// Add to 2d list
		d.add(unicodePoints)

	}

}

// CreateAWSConfigFromEnv creates an AWS config object from environment variables
func CreateAWSConfigFromEnv(region string) (aws.Config, error) {
	// Get environment variables
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	// Validate credentials
	if accessKey == "" || secretKey == "" {
		return aws.Config{}, fmt.Errorf("missing AWS credentials in environment")
	}

	// Establish configuration
	creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""))
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Return configuration
	return cfg, nil
}

// listS3Keys retrieves all keys in an S3 bucket
func listS3Keys(bucket, region string) ([]string, error) {
	// Get configuration
	cfg, err := CreateAWSConfigFromEnv(region)
	if err != nil {
		return nil, err
	}

	// Get S3 client
	client := s3.NewFromConfig(cfg)

	// Track variables
	var keys []string
	var continuationToken *string = nil
	for {
		// Service call
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

		// End
		if !*resp.IsTruncated {
			break
		}
		continuationToken = resp.NextContinuationToken
	}

	// Return keys
	return keys, nil
}

// fetchJSONFromS3 retrieves a JSON object from S3 and unmarshals it into a map
func fetchJSONFromS3(bucket string, key string) (map[string]interface{}, error) {
	// Context
	ctx := context.Background()

	// State
	region := "us-east-1"

	// Read credentials from environment variables
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	// Validation
	if accessKey == "" || secretKey == "" || region == "" {
		return nil, fmt.Errorf("missing AWS credentials or region in environment variables")
	}

	// Set up custom AWS config with static credentials
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg)

	// Fetch object
	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer resp.Body.Close()

	// Read and unmarshal JSON
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

// getMaxToken scans a list of unicode point sequences and returns the highest token value.
func getMaxToken(dataset *Dataset) int64 {
	var maxToken int64 = -1
	for _, sequence := range dataset.sentences {
		for _, token := range sequence {
			if token > maxToken {
				maxToken = token
			}
		}
	}
	return maxToken
}

// replaces one token with another
func replace(pair [2]int64, mintToken int64, dataset *Dataset) {
	// new list
	var newList [][]int64

	for _, sequence := range dataset.sentences {
		index := 0
		var newSequence []int64
		for index < len(sequence)-1 {
			if sequence[index] == pair[0] && sequence[index+1] == pair[1] {
				newSequence = append(newSequence, mintToken)
				index += 2
			} else {
				newSequence = append(newSequence, sequence[index])
				index += 1
			}
		}
		newList = append(newList, newSequence)
	}

	// reassign
	dataset.sentences = newList
}

// get the vocab size
func getVocabSize(dataset *Dataset) int {
	unique := make(map[int64]bool)
	for _, sequence := range dataset.sentences {
		for index := range sequence {
			unique[sequence[index]] = true
		}
	}
	return len(unique)
}

// get sequence length
func getTotalSequenceLength(dataset *Dataset) int64 {
	var count int64
	for _, sequence := range dataset.sentences {
		count += int64(len(sequence))
	}
	return count
}

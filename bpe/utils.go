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

// dataDataset holds the sentences and a mutex for concurrent access.
type dataDataset struct {
	aalSentences [][]int64
	pdMutex      *sync.Mutex
}

// dataStatistics holds the frequency of pairs and a mutex for concurrent access.
type dataStatistics struct {
	mapPairFrequency map[[2]int64]int
	pdMutex          *sync.Mutex
	palMaxPair       *[2]int64
	piMaxCount       *int
}

// insertMerge increments the count for a given pair in StatisticsMap
// also tracks the most frequently occurring pair
func insertMerge(dataStatistics *dataStatistics, alPair [2]int64) {
	dataStatistics.pdMutex.Lock()
	defer dataStatistics.pdMutex.Unlock()

	// Increment pair
	dataStatistics.mapPairFrequency[alPair]++

	// update max pair
	if dataStatistics.mapPairFrequency[alPair] > *dataStatistics.piMaxCount {
		*dataStatistics.palMaxPair = alPair
		*dataStatistics.piMaxCount = dataStatistics.mapPairFrequency[alPair]
	}
}

// add a single sentence to a list
func (d *dataDataset) add(alSentence []int64) {
	d.pdMutex.Lock()
	defer d.pdMutex.Unlock()
	d.aalSentences = append(d.aalSentences, alSentence)
}

// add set of sentences to a list
func (d *dataDataset) AddList(adataSentences []interface{}) {
	for index := range adataSentences {
		// Normalize the sentence
		sentence := normalize.Normalize(adataSentences[index].(string))

		// Convert to unicode integers
		var unicodePoints []int64
		for _, r := range sentence {
			unicodePoints = append(unicodePoints, int64(r))
		}

		// Add to list
		d.add(unicodePoints)
	}
}

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
	dataCredentials := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(sAccessKey, sSecretKey, ""))
	dataConfiguration, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(dataCredentials),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Return configuration
	return dataConfiguration, nil
}

// listS3Keys retrieves all keys in an S3 bucket
func listS3Keys(sBucket string, sRegion string) ([]string, error) {
	// Get configuration
	dataConfiguration, err := CreateAWSConfigFromEnv(sRegion)
	if err != nil {
		return nil, err
	}

	// Get S3 client
	dataClient := s3.NewFromConfig(dataConfiguration)

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
		for _, obj := range pdResponse.Contents {
			asKeys = append(asKeys, aws.ToString(obj.Key))
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
	dataConfiguration, err := config.LoadDefaultConfig(dataContext,
		config.WithRegion(sRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(sAccessKey, sSecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	pdClient := s3.NewFromConfig(dataConfiguration)

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
	var mapResult map[string]interface{}
	if err := json.Unmarshal(abBody, &mapResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return mapResult, nil
}

// getMaxToken scans a list of unicode point sequences and returns the highest token value.
func getMaxToken(dataset *dataDataset) int64 {
	var lMaxToken int64 = -1
	for _, alSentence := range dataset.aalSentences {
		for _, lToken := range alSentence {
			if lToken > lMaxToken {
				lMaxToken = lToken
			}
		}
	}
	return lMaxToken
}

// replaces one token with another
func replace(alPair [2]int64, lMintToken int64, dataset *dataDataset) {
	// new list
	var aalNewList [][]int64

	for _, sequence := range dataset.aalSentences {
		index := 0
		var alNewSequence []int64
		for index < len(sequence)-1 {
			if sequence[index] == alPair[0] && sequence[index+1] == alPair[1] {
				alNewSequence = append(alNewSequence, lMintToken)
				index += 2
			} else {
				alNewSequence = append(alNewSequence, sequence[index])
				index += 1
			}
		}
		aalNewList = append(aalNewList, alNewSequence)
	}

	// reassign
	dataset.aalSentences = aalNewList
}

// get the vocab size
func getVocabSize(dataset *dataDataset) int {
	mapUnique := make(map[int64]bool)
	for _, alSequence := range dataset.aalSentences {
		for index := range alSequence {
			mapUnique[alSequence[index]] = true
		}
	}
	return len(mapUnique)
}

// get sequence length
func getTotalSequenceLength(dataset *dataDataset) int64 {
	var lCount int64
	for _, alSequence := range dataset.aalSentences {
		lCount += int64(len(alSequence))
	}
	return lCount
}

package bpe

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// merge implements the byte pair encoding algorithm and returns an error if the merge process fails.
func merge(dataDataset *dataDataset) (*Merges, error) {
	// Initialize max token value
	lMintToken := getMaxToken(dataDataset) + 1

	// Before vocab size
	lOldSequenceLength := getTotalSequenceLength(dataDataset)

	// store merges
	dataMerges := &Merges{
		mapMerges: make(map[[2]int64]int64),
		alKeys:    [][2]int64{},
	}

	// start time
	mainStart := time.Now()

	for {
		start := time.Now()

		// initialize a map
		pdMergeStatistics := &dataStatistics{
			mapPairFrequency: make(map[[2]int64]int),
			pdMutex:          &sync.Mutex{},
		}

		// Populate merge pairs based on current sentences
		// Store the most frequently occurring pair
		err := countStatistics(pdMergeStatistics, dataDataset)
		if err != nil {
			return nil, fmt.Errorf("failed to generate merge pairs: %w", err)
		}

		fmt.Println("Step 1", time.Since(start))

		// store merges
		dataMerges.insertMerge(*pdMergeStatistics.palMaxPair, lMintToken)

		fmt.Println("Step 2", time.Since(start))

		// replace max pair with the minted token
		replace(*pdMergeStatistics.palMaxPair, lMintToken, dataDataset)

		fmt.Println("Step 3", time.Since(start))

		// After vocab size
		newSequence := getTotalSequenceLength(dataDataset)

		fmt.Println("Step 4", time.Since(start))

		// calculate compression ratio
		fCompressionRatio := float64(lOldSequenceLength) / float64(newSequence)
		fmt.Println(time.Since(mainStart), fCompressionRatio, string(pdMergeStatistics.palMaxPair[0]), string(pdMergeStatistics.palMaxPair[1]))

		fmt.Println("Step 5", time.Since(start))
		fmt.Println("======================================")

		// Break after a certain ratio
		if fCompressionRatio > 5 {
			break
		}

		// next minted token
		lMintToken += 1
	}
	return dataMerges, nil
}

// countStatistics analyzes the dataset's sentences to create and track pairs of adjacent unicode points.
func countStatistics(dataStatistics *dataStatistics, dataDataset *dataDataset) error {
	// Use all available CPU cores with a small buffer
	numWorkers := runtime.NumCPU()

	// Create channels for work distribution and results collection
	jobs := make(chan []int64, numWorkers)
	results := make(chan map[[2]int64]int, numWorkers)

	// Start worker goroutines
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			localFrequency := make(map[[2]int64]int)

			// Process each sentence assigned to this worker
			for sentence := range jobs {
				sentenceLen := len(sentence)
				for j := 0; j < sentenceLen-1; j++ {
					alPair := [2]int64{sentence[j], sentence[j+1]}
					localFrequency[alPair]++
				}
			}

			// Send local results back
			results <- localFrequency
		}()
	}

	// Distribute work across goroutines
	go func() {
		// Calculate optimal batch size - not too small to minimize overhead
		// but not too large to ensure good distribution
		totalSentences := len(dataDataset.aalSentences)
		batchSize := calculateBatchSize(totalSentences, numWorkers)

		for i := 0; i < totalSentences; i += batchSize {
			end := i + batchSize
			if end > totalSentences {
				end = totalSentences
			}

			// Process a batch of sentences at a time
			for j := i; j < end; j++ {
				jobs <- dataDataset.aalSentences[j]
			}
		}
		close(jobs)
	}()

	// Close results channel when all workers finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Initialize map if needed
	if dataStatistics.mapPairFrequency == nil {
		dataStatistics.mapPairFrequency = make(map[[2]int64]int)
	}

	// Collect and merge results
	iMaxCount := 0
	alMaxPair := [2]int64{-1, -1}

	for localFreq := range results {
		for pair, count := range localFreq {
			dataStatistics.mapPairFrequency[pair] += count

			// Update max if needed
			if dataStatistics.mapPairFrequency[pair] > iMaxCount {
				iMaxCount = dataStatistics.mapPairFrequency[pair]
				alMaxPair = pair
			}
		}
	}

	// Set final values
	dataStatistics.piMaxCount = &iMaxCount
	dataStatistics.palMaxPair = &alMaxPair

	return nil
}

// calculateBatchSize determines optimal batch size based on dataset size and workers
func calculateBatchSize(totalItems, numWorkers int) int {
	// Aim for each worker to process multiple batches for better load balancing
	// but keep batches large enough to minimize overhead
	desiredBatchesPerWorker := 4
	batchSize := totalItems / (numWorkers * desiredBatchesPerWorker)

	// Enforce minimum and maximum batch sizes
	minBatchSize := 100
	maxBatchSize := 10000

	if batchSize < minBatchSize {
		return minBatchSize
	}
	if batchSize > maxBatchSize {
		return maxBatchSize
	}
	return batchSize
}

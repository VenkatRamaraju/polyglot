module tokenizer

require (
normalize v0.0.0
bpe v0.0.0
)

require golang.org/x/text v0.23.0 // indirect

replace (
normalize => ./normalize
bpe => ./bpe
)

go 1.24.2

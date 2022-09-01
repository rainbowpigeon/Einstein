package neurons

import (
	"math/rand"
	"strconv"
	"time"
)

const yearLimit = 2018

var timeLimit = time.Date(yearLimit, time.January, 1, 1, 0, 0, 0, time.UTC)
var utilRand *rand.Rand

func init() {
	initUtilRand()
}

func initUtilRand() {
	// TODO: there should honestly really be a cleaner way of handling this init dependency
	if utilRand == nil {
		utilRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
}

// adapted from google/uuid/version4.go -> NewRandomFromReader()
func generateUUID() ([]byte, error) {
	UUID := make([]byte, 16)
	_, err := utilRand.Read(UUID)
	if err != nil {
		return nil, err
	}
	UUID[6] = (UUID[6] & 0x0f) | 0x40 // Version 4
	UUID[8] = (UUID[8] & 0x3f) | 0x80 // Variant is 10
	return UUID, nil
}

// ref: https://stackoverflow.com/a/51327810
func generateRandomBool() bool {
	return utilRand.Int63()&(1<<62) == 0
}

func generateRandomTimestamp(timeType time.Duration) int64 {
	randomTime := utilRand.Int63n(time.Now().Unix()-timeLimit.Unix()) + timeLimit.Unix() // randomTime will be in seconds
	return randomTime * int64(time.Second/timeType)
}

func isReasonableTimestamp(timestamp string, timeType time.Duration) bool {
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}

	var unixTime time.Time

	switch timeType {
	case time.Second:
		unixTime = time.Unix(i, 0)
	case time.Millisecond:
		unixTime = time.UnixMilli(i)
	case time.Microsecond:
		unixTime = time.UnixMicro(i)
	case time.Nanosecond:
		unixTime = time.Unix(0, i)
	}

	if unixTime.Before(timeLimit) {
		return false
	} else {
		return true
	}
}

// generics were introduced in 1.18 yay!
func chooseRandom[T any](genericSlice []T) T {
	return genericSlice[utilRand.Intn(len(genericSlice))]
}

// TODO: add variant of createChunks that takes in numberOfChunks instead of chunkSize to generate
// modified from improved variant here: https://stackoverflow.com/a/61469854
func createChunks(s string, chunkSize int) []string {
	stringLength := len(s)

	if stringLength == 0 {
		return nil
	}
	if chunkSize >= stringLength {
		return []string{s}
	}
	var chunks []string = make([]string, 0, (stringLength-1)/chunkSize+1)
	currentLen := 0
	currentStart := 0
	for i := range s {
		if currentLen == chunkSize {
			chunks = append(chunks, s[currentStart:i])
			currentLen = 0
			currentStart = i
		}
		currentLen++
	}
	chunks = append(chunks, s[currentStart:])
	return chunks
}

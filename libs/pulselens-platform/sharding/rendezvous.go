package sharding

import (
	"hash/fnv"
	"strconv"
)

func BucketForKey(key string, buckets int) int {
	if buckets <= 1 {
		return 0
	}

	selectedBucket := 0
	var selectedScore uint64
	for bucket := 0; bucket < buckets; bucket++ {
		score := scoreForNode(key, bucket)
		if bucket == 0 || score > selectedScore {
			selectedBucket = bucket
			selectedScore = score
		}
	}

	return selectedBucket
}

func scoreForNode(key string, bucket int) uint64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(key))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write([]byte(strconv.Itoa(bucket)))
	return hasher.Sum64()
}

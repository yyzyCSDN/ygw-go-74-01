package particle

import (
	"github.com/cespare/xxhash/v2"

	"cleanroomorcontrol/internal/model"
)

func PointHash(point model.PointID) uint64 {
	return xxhash.Sum64String(string(point))
}

func PointBucket(point model.PointID, buckets int) int {
	if buckets <= 0 {
		return 0
	}
	return int(PointHash(point) % uint64(buckets))
}

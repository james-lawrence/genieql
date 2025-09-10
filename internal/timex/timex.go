package timex

import (
	"math"
	"time"
)

// Inf - positive infinity no time can be larger.
// see https://stackoverflow.com/questions/25065055/what-is-the-maximum-time-time-in-go/32620397
func Inf() time.Time {
	return time.Unix(math.MaxInt64-62135596800, 999999999)
}

func NegInf() time.Time {
	return time.Unix(math.MinInt64, math.MinInt64)
}

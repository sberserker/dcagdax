package gemini

import (
	"time"
)

type Args map[string]interface{}

// Nonce returns a generic nonce based on unix timestamp
func nonce() int64 {
	return time.Now().UnixNano()
}

func msToTime(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}

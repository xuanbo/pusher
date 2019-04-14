package pusher

import "time"

// current timestamp
func Timestamp() int64 {
	return time.Now().Unix()
}

package metrics

import "sync/atomic"

var visits int64
var generates int64

func IncVisit() { atomic.AddInt64(&visits, 1) }
func IncGenerate() { atomic.AddInt64(&generates, 1) }
func Snapshot() (int64, int64) { return atomic.LoadInt64(&visits), atomic.LoadInt64(&generates) }

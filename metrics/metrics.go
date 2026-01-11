package metrics

import (
	"database/sql"
	"log"
	"sync/atomic"
)

var visits int64
var generates int64
var db *sql.DB

func useDB() bool { return db != nil }

func Init(d *sql.DB) {
	db = d
}

func IncVisit() {
	atomic.AddInt64(&visits, 1)
	if useDB() {
		if _, err := db.Exec("UPDATE metrics_counters SET visits=visits+1, updated_at=NOW() WHERE id=1"); err != nil {
			log.Printf("metrics visit db err: %v", err)
		}
	}
}

func IncGenerate() {
	atomic.AddInt64(&generates, 1)
	if useDB() {
		if _, err := db.Exec("UPDATE metrics_counters SET generates=generates+1, updated_at=NOW() WHERE id=1"); err != nil {
			log.Printf("metrics generate db err: %v", err)
		}
	}
}

func Snapshot() (int64, int64) { return atomic.LoadInt64(&visits), atomic.LoadInt64(&generates) }

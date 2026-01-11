package metrics

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func SetupDB(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, nil
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS metrics_counters (
            id TINYINT PRIMARY KEY,
            visits BIGINT NOT NULL DEFAULT 0,
            generates BIGINT NOT NULL DEFAULT 0,
            updated_at TIMESTAMP NULL DEFAULT NULL
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
    `)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`INSERT IGNORE INTO metrics_counters (id, visits, generates, updated_at) VALUES (1, 0, 0, NULL)`)
	if err != nil {
		log.Printf("seed metrics row err: %v", err)
	}
	return db, nil
}

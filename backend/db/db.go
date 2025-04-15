// Package db provides database access functionality
package db

import (
	"database/sql"
	"log"
	"sync"

	_ "github.com/lib/pq"
)

// Queries provides all the database operations
type Queries struct {
	db *sql.DB
}

// Global database connection instance
var (
	globalDB *sql.DB
	dbMutex  sync.Mutex
)

// SetGlobalDB sets the global database connection
func SetGlobalDB(db *sql.DB) {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	globalDB = db
}

// NewQueries creates a new Queries instance
func NewQueries() *Queries {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	return &Queries{db: globalDB}
}

// SetDB sets the database connection
func (q *Queries) SetDB(db *sql.DB) {
	q.db = db
}

// GetDB returns the database connection
func (q *Queries) GetDB() *sql.DB {
	return q.db
}

// Close closes the database connection
func (q *Queries) Close() error {
	if q.db != nil {
		return q.db.Close()
	}
	return nil
}

// Ping checks the database connection
func (q *Queries) Ping() error {
	if q.db != nil {
		return q.db.Ping()
	}
	log.Println("Warning: No database connection set")
	return nil
}

// filepath: /home/sagar/Workspace/priya/TurpleMQ/internal/services/database.go
package services

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DatabaseService struct {
	db *sql.DB
}

// NewDatabaseService creates a new DatabaseService instance
func NewDatabaseService(dbPath string) (*DatabaseService, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create messages table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			topic TEXT NOT NULL,
			message TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, err
	}

	return &DatabaseService{
		db: db,
	}, nil
}

// SaveMessage saves a message to the database
func (ds *DatabaseService) SaveMessage(topic string, message string) error {
	_, err := ds.db.Exec(
		"INSERT INTO messages (topic, message, created_at) VALUES (?, ?, ?)",
		topic, message, time.Now(),
	)
	if err != nil {
		log.Printf("Error saving message to database: %v", err)
		return err
	}
	return nil
}

// GetMessagesByTopic retrieves messages for a specific topic
func (ds *DatabaseService) GetMessagesByTopic(topic string) ([]map[string]interface{}, error) {
	rows, err := ds.db.Query(
		"SELECT id, topic, message, created_at FROM messages WHERE topic = ? ORDER BY created_at DESC LIMIT 100",
		topic,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []map[string]interface{}
	for rows.Next() {
		var id int64
		var topic, message string
		var createdAt string
		if err := rows.Scan(&id, &topic, &message, &createdAt); err != nil {
			return nil, err
		}
		messages = append(messages, map[string]interface{}{
			"id":         id,
			"topic":      topic,
			"message":    message,
			"created_at": createdAt,
		})
	}

	return messages, nil
}

// Close closes the database connection
func (ds *DatabaseService) Close() error {
	return ds.db.Close()
}

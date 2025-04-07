package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/turplespace/msgqueue/internal/models"
	"github.com/turplespace/msgqueue/internal/services"
)

type WebSocketService struct {
	handler   *models.WebSocketHandler
	dbService *services.DatabaseService
}

func (wsh WebSocketService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := wsh.handler.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade Error:", err)
		return
	}

	defer conn.Close()
	var message models.Message
	newQueueService := services.NewQueueService(wsh.handler)
	newPublishService := services.NewPublishService(wsh.handler)
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Connection closed: %v\n", err)
				newPublishService.RemoveConnection(conn)
				return
			}
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Unexpected close error: %v\n", err)
				newPublishService.RemoveConnection(conn)
				return // Add return here to prevent further execution with a broken connection
			}
			log.Println("Read Error:", err)
			return
		}
		err = json.Unmarshal(msg, &message)
		if err != nil {
			log.Printf("JSON Unmarshal Error: %v in message: %s", err, string(msg))
			continue
		}
		if message.Role == "consumer" {
			if message.Subscribe {
				newPublishService.AddSubscribers(message.Topic, conn)
				log.Println("Subscribed to topic:", message.Topic)
			} else {
				msg, is_data := newQueueService.DeQueue(message.Topic)

				if is_data {
					conn.WriteJSON(msg)
					log.Printf("Data sent to consumer: %s\n", msg)
				} else {
					log.Println("No data in queue")
				}

			}
		} else if message.Role == "producer" {
			if message.TransmissionMode == "buffered" {
				newQueueService.EnQueue(message.Topic, message.Message)
				log.Printf("Message buffered: %s\n", message.Message)

			} else if message.TransmissionMode == "broadcast" {
				// Save broadcast message to database
				if wsh.dbService != nil {
					if err := wsh.dbService.SaveMessage(message.Topic, message.Message); err != nil {
						log.Printf("Error saving message to database: %v\n", err)
					} else {
						log.Printf("Message saved to database for topic '%s'\n", message.Topic)
					}
				}

				newPublishService.SendMessageToSubscribers(message)
				log.Printf("Message broadcasted: %s\n", message.Message)
			} else {
				log.Printf("Invalid TransmissionMode %s\n", message.TransmissionMode)
			}

		} else {
			log.Printf("Invalid Role %s\n", message.Role)
		}
	}
}

// CORS middleware to allow cross-origin requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleMessageHistory returns the message history for a given topic
func handleMessageHistory(dbService *services.DatabaseService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		topic := r.URL.Query().Get("topic")
		if topic == "" {
			http.Error(w, "Topic is required", http.StatusBadRequest)
			return
		}

		messages, err := dbService.GetMessagesByTopic(topic)
		if err != nil {
			log.Printf("Error retrieving messages: %v", err)
			http.Error(w, "Failed to retrieve messages", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"topic":    topic,
			"messages": messages,
		})
	}
}

// serveStaticFiles handles serving the React frontend static files
func serveStaticFiles() http.Handler {
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		// Default to "static" directory in the current working directory
		staticDir = "./static"
	}

	fileServer := http.FileServer(http.Dir(staticDir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For debugging
		log.Printf("Request for path: %s", r.URL.Path)

		// Special handling for asset files
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			fileServer.ServeHTTP(w, r)
			return
		}

		// Serve index.html for the root path
		if r.URL.Path == "/" {
			indexPath := filepath.Join(staticDir, "index.html")
			if _, err := os.Stat(indexPath); os.IsNotExist(err) {
				log.Printf("index.html not found at %s", indexPath)
				http.Error(w, "index.html not found", http.StatusNotFound)
				return
			}
			http.ServeFile(w, r, indexPath)
			return
		}

		// Check if the requested file exists
		path := filepath.Join(staticDir, r.URL.Path)
		_, err := os.Stat(path)

		// If the file doesn't exist or if it's a client-side route, serve index.html
		if os.IsNotExist(err) || strings.HasPrefix(r.URL.Path, "/room/") {
			indexPath := filepath.Join(staticDir, "index.html")
			if _, err := os.Stat(indexPath); os.IsNotExist(err) {
				log.Printf("index.html not found at %s", indexPath)
				http.Error(w, "index.html not found", http.StatusNotFound)
				return
			}
			http.ServeFile(w, r, indexPath)
			return
		}

		// Otherwise, serve the requested file
		fileServer.ServeHTTP(w, r)
	})
}

func main() {
	// Initialize database
	dbService, err := services.NewDatabaseService("./turplemq.db")
	if err != nil {
		log.Printf("Warning: Failed to initialize database: %v\n", err)
		log.Println("Continuing without database persistence...")
		dbService = nil
	} else {
		log.Println("Database initialized successfully")
		defer dbService.Close()
	}

	ws := models.WebSocketHandler{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		Subscribers: make(map[string][]*websocket.Conn),
		Queue:       make(map[string][]string),
	}
	handler := &WebSocketService{
		handler:   &ws,
		dbService: dbService,
	}

	// Create router
	mux := http.NewServeMux()

	// WebSocket handler
	mux.Handle("/ws", handler)

	// Message history API endpoint with CORS
	if dbService != nil {
		mux.HandleFunc("/messages", handleMessageHistory(dbService))
		log.Println("Message history API enabled at /messages?topic=<topic_name>")
	}

	// Serve static files for the React frontend
	staticHandler := serveStaticFiles()
	mux.Handle("/", staticHandler)
	mux.Handle("/assets/", staticHandler)

	// Print the current working directory for debugging
	cwd, err := os.Getwd()
	if err == nil {
		log.Printf("Current working directory: %s", cwd)
	}

	// Check if static directory exists
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./static"
	}
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		log.Printf("Warning: Static directory not found: %s", staticDir)
	} else {
		log.Printf("Static directory found: %s", staticDir)
		// List files in static directory
		if files, err := os.ReadDir(staticDir); err == nil {
			log.Printf("Files in static directory:")
			for _, file := range files {
				log.Printf("- %s", file.Name())
			}
		}
	}

	log.Println("Server listening on :8080")
	log.Println("Frontend available at http://localhost:8080")

	// Wrap with CORS middleware
	corsHandler := corsMiddleware(mux)

	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}

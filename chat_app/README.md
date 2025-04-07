# TurpleMQ Chat Application

A real-time chat application built with React that uses TurpleMQ for messaging.

## Features

- Real-time messaging via WebSockets
- Create or join chat rooms
- Message history for existing rooms
- User identification
- Full-duplex communication (simultaneous send/receive)
- Responsive UI with Tailwind CSS

## Getting Started

### Development Mode

To run this application in development mode:

1. Start the TurpleMQ backend:
   ```
   cd /path/to/TurpleMQ
   go run main.go
   ```

2. Start the React development server:
   ```
   cd /path/to/TurpleMQ/chat_app
   npm install
   npm run dev
   ```

3. Open your browser and navigate to `http://localhost:5173`

### Production Build

The application can be built and served from the TurpleMQ backend:

1. Build the Docker image:
   ```
   docker build -t turblemq-chat .
   ```

2. Run the container:
   ```
   docker run -p 8080:8080 -v turplemq_data:/data turblemq-chat
   ```

3. Open your browser and navigate to `http://localhost:8080`

## Using the Chat

1. Enter a username on the welcome screen
2. Choose to either:
   - Create a new room (generates a random room ID)
   - Join an existing room (requires room ID)
3. Start chatting!
4. Share the room ID with others so they can join your room

## Tech Stack

- Frontend: React with Tailwind CSS
- Backend: Go with WebSocket support
- Message persistence: SQLite

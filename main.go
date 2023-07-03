package main

import (
	"encoding/json"
	"fmt"
	"game"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var gameInstance *game.Game
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan *game.Game)

func main() {
	gameInstance = &game.Game{
		Width:   40,
		Height:  40,
		Players: make(map[string]*game.Player),
	}

	fmt.Println("Server started!")

	http.HandleFunc("/", handleRequest)
	go handlePlayerMovements()

	log.Fatal(http.ListenAndServe(":5555", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	// Add the new client connection to the clients map
	clients[conn] = true

	// Handle WebSocket communication for the client
	handleWebSocket(conn)
}

func handleWebSocket(conn *websocket.Conn) {
	// Close the connection and remove the client from the clients map when the function returns
	defer func() {
		delete(clients, conn)
		conn.Close()
	}()

	// Read and process incoming WebSocket messages
	for {
		// Read a message from the client
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		// Process the received message and update the game state accordingly
		var data map[string]interface{}
		err = json.Unmarshal(message, &data)
		if err != nil {
			log.Println("Error decoding WebSocket message:", err)
			continue
		}

		// Extract client ID from request headers
		clientID := conn.RemoteAddr().String()

		// Retrieve or create the player for the client
		player := getPlayer(clientID)

		// Move the player in the game
		direction, ok := data["direction"].(string)
		playerID, ok2 := data["player_id"].(string)
		if ok && ok2 {
			gameInstance.MovePlayer(playerID, direction, player)
		}

		// Broadcast the updated game state to all connected clients
		broadcast <- gameInstance
	}
}

func handlePlayerMovements() {
	for {
		// Wait for a game state update
		gameState := <-broadcast

		// Prepare response data
		responseData := map[string]interface{}{
			"gameState": gameState,
		}

		// Convert response data to JSON
		responseBody, err := json.Marshal(responseData)
		if err != nil {
			log.Println("Error preparing response:", err)
			continue
		}

		// Send the updated game state to all connected clients
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, responseBody)
			if err != nil {
				log.Println("WebSocket write error:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func getPlayer(clientID string) *game.Player {
	player, ok := gameInstance.Players[clientID]
	if !ok {
		// Create a new player for the client
		player = &game.Player{X: 0, Y: 0, Alive: true}
		gameInstance.Players[clientID] = player
	}
	return player
}

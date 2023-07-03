package main

import (
	"encoding/json"
	"fmt"
	"game"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

func main() {
	fmt.Println("Welcome to Nowhereland!\n\n")

	// Prompt the user for the client ID
	clientID := readInput("Enter your client ID: ")

	conn, err := establishWebSocketConnection(clientID)
	if err != nil {
		log.Fatalf("WebSocket connection error: %v", err)
	}
	defer conn.Close()

	// Create a goroutine to continuously listen for incoming messages
	go func() {
		for {

			gameState, err := receiveWebSocketMessage(conn)
			if err != nil {
				log.Fatalf("WebSocket receive error: %v", err)
			}

			// Render the game state
			renderGame(gameState)
		}
	}()

	for {
		dir := readInput("In what direction do you want to move (w,a,s,d) ?: ")

		err := sendWebSocketMessage(clientID, conn, dir)
		if err != nil {
			log.Fatalf("WebSocket send error: %v", err)
		}
	}
}

func establishWebSocketConnection(clientID string) (*websocket.Conn, error) {
	url := "ws://localhost:5555/ws"

	header := http.Header{}
	header.Add("X-Client-ID", clientID)

	conn, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func sendWebSocketMessage(playerID string, conn *websocket.Conn, direction string) error {
	message := map[string]interface{}{
		"player_id": playerID,
		"direction": direction,
	}

	err := conn.WriteJSON(message)
	if err != nil {
		return err
	}

	return nil
}

func receiveWebSocketMessage(conn *websocket.Conn) (*game.Game, error) {
	var message map[string]interface{}
	err := conn.ReadJSON(&message)
	if err != nil {
		return nil, err
	}

	gameStateJSON, err := json.Marshal(message["gameState"])
	if err != nil {
		return nil, err
	}

	var gameState game.Game
	err = json.Unmarshal(gameStateJSON, &gameState)
	if err != nil {
		return nil, err
	}

	return &gameState, nil
}

func renderGame(gameState *game.Game) {
	fmt.Println()

	// Create a two-dimensional grid with the field dimensions, filled with "."
	grid := make([][]string, gameState.Height)
	for i := range grid {
		grid[i] = make([]string, gameState.Width)
		for j := range grid[i] {
			grid[i][j] = "."
		}
	}

	// Add player symbols to their respective coordinates in the grid
	for _, player := range gameState.Players {
		if player.Alive {
			grid[player.Y][player.X] = "Њ"
		}
	}

	// Print the grid to the console
	for _, row := range grid {
		fmt.Println(strings.Join(row, " "))
	}
}

func readInput(prompt string) string {
	var input string
	fmt.Print(prompt)
	fmt.Scanln(&input)
	return strings.TrimSpace(input)
}

package example

import (
	"encoding/json"
	"log"
	"net/http"
)

// TransferRequest represents the request body for the /transfer endpoint
type TransferRequest struct {
	From   int `json:"from"`
	To     int `json:"to"`
	Amount int `json:"amount"`
}

// handleTransfer processes the transfer request
func handleTransfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TransferRequest

	// Decode the JSON request body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Here you would implement your logic for handling the transfer.
	// For now, we'll just log the details.
	log.Printf("Transfer request: From=%d, To=%d, Amount=%d", req.From, req.To, req.Amount)

	// Respond with a success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Transfer successful"))
}

func main() {
	http.HandleFunc("/transfer", handleTransfer)

	// Start the server on port 8080
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/bryangodson/detect-fraud-mvp/models"
	"github.com/bryangodson/detect-fraud-mvp/services"
)

func FraudCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the transaction from the request body
	var tx models.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set a timeout for fraud check
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Call the fraud check service ,runs rules+ML+logging
	score, decision, reasons, err := services.CheckFraud(ctx, tx)
	if err != nil {
		// Return a 202 Accepted with review if scoring had an internal error
		// but we still want to be conservative.
		resp := map[string]interface{}{
			"score":    score,
			"decision": decision,
			"reasons":  reasons,
			"error":    err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(resp)
		return
	}
	// Normal success path: return the score and decision
	resp := map[string]interface{}{
		"score":    score,
		"decision": decision,
		"reasons":  reasons,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

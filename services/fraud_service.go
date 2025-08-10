package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/bryangodson/detect-fraud-mvp/models"
	"github.com/bryangodson/detect-fraud-mvp/rules"
	_ "github.com/lib/pq"
)

//Db is a package-level varaible set in main.go if Postgress is available
//It is used to interact with the database; if not logs to file

var DB *sql.DB

// ModeServerUrl is the external model scoring endpoint. set via env var
var ModelServerURL string = os.Getenv("MODEL_SERVER_URL")

// modelResponse models the JSON returned by the model server
type modelResponse struct {
	Score float64 `json:"score"` //fraud score between 0.0 and 1.0
}

// CheckFraud runs rules,calls ML, decides action and logs the decision.
// Returns the fraud score,decision string,reasons if any and any error encountered
func CheckFraud(ctx context.Context, tx models.Transaction) (float64, string, []string, error) {
	//run quick rules
	fraudulentTransacitons := rules.BasicRules(tx)
	if len(fraudulentTransacitons) > 0 {
		//if fraud was detected, immediate action is taken
		//by assigning a high risk score (0.95) and "BLOCK" decision
		reasons := make([]string, len(fraudulentTransacitons))
		for i, r := range fraudulentTransacitons {
			reasons[i] = r.Reason
		}
		//attemp to log decision
		_ = logDecision(ctx, tx, 0.96, "BLOCK", reasons)
		return 0.95, "BLOCK", reasons, nil
	}

	//no rule blocked - call ML model
	mlScore, err := callModelServer(ctx, tx)
	if err != nil {
		// if model call fails, be conservative : place in REVIEW queue
		reasons := []string{"Model call failed: " + err.Error()}
		_ = logDecision(ctx, tx, 0.96, "REVIEW", reasons)
		return 0.5, "REVIEW", reasons, err
	}

	//decision threshholds
	decision := "ALLOW"
	if mlScore >= 0.7 {
		decision = "BLOCK"
	} else if mlScore >= 0.4 {
		decision = "REVIEW"
	}

	reasons := []string{} // no specific rule reasons
	if err := logDecision(ctx, tx, mlScore, decision, reasons); err != nil {
		// logging failure is non-fatal for the transaction flow, but we surface it.
		return mlScore, decision, reasons, fmt.Errorf("logging error: %w", err)
	}

	return mlScore, decision, reasons, nil
}

// callModelServer sends selected features to the ML model server and reads the score.
// This function uses the MODEL_SERVER_URL env var (set in start/config).
func callModelServer(ctx context.Context, tx models.Transaction) (float64, error) {
	// If no model server configured, fail fast so caller can handle fallback.
	if ModelServerURL == "" {
		return 0, errors.New("MODEL_SERVER_URL not set")
	}

	// Build a small payload with the features the model expects.
	// Keep payload compact: only the features the model needs.
	payload := map[string]interface{}{
		"amount":        tx.Amount,
		"account_age":   tx.AccountAgeDays,
		"is_new_device": tx.IsNewDevice,
		"country":       tx.Country,
		// add more features as your model requires
	}

	// Encode payload as JSON
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	// Use a context-aware HTTP client for timeouts
	req, err := http.NewRequestWithContext(ctx, "POST", ModelServerURL, bytesReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 3 * time.Second} // small timeout for real-time scoring
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("model server returned status %d", resp.StatusCode)
	}

	var mr modelResponse
	if err := json.NewDecoder(resp.Body).Decode(&mr); err != nil {
		return 0, err
	}

	// sanity-check score bounds
	if mr.Score < 0 {
		mr.Score = 0
	} else if mr.Score > 1 {
		mr.Score = 1
	}

	return mr.Score, nil
}

// bytesReader returns an io.Reader for a byte slice.
// We make this helper so the call is explicit (like Node Buffer).
func bytesReader(b []byte) *bytesReaderImpl { return &bytesReaderImpl{b: b} }

type bytesReaderImpl struct{ b []byte }

func (r *bytesReaderImpl) Read(p []byte) (n int, err error) {
	if len(r.b) == 0 {
		return 0, io.EOF
	}
	n = copy(p, r.b)
	r.b = r.b[n:]
	return n, nil
}

// logDecision stores a record of the decision for auditing. if DB is not available, it writes to a file
func logDecision(ctx context.Context, tx models.Transaction, score float64, decision string, reasons []string) error {

	//payload
	rec := map[string]interface{}{
		"transaction_id": tx.ID,
		"user_id":        tx.UserID,
		"amount":         tx.Amount,
		"score":          score,
		"decision":       decision,
		"reasons":        reasons,
		"timestamp":      time.Now().UTC(),
	}

	b, _ := json.Marshal(rec)
	if DB != nil {
		//best effort insert
		_, err := DB.ExecContext(ctx,
			`INSERT INTO fraud_decisions(transaction_id,user_id,amount,score,decision,resons,created_at) VALUES ($1,$2,$3,$4,$5,$6,NOW())`,
			tx.ID, tx.UserID, tx.Amount, score, decision, string(b))
		return err
	}

	// Fallback to file logger
	f, err := os.OpenFile("decisions.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(append(b, '\n')); err != nil {
		return err
	}
	return nil
}

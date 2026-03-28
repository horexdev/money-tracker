package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const initDataMaxAge = 24 * time.Hour

// ErrInvalidInitData is returned when the Telegram initData is missing, malformed, or has an invalid hash.
var ErrInvalidInitData = errors.New("invalid or expired Telegram initData")

// ValidateInitData validates the Telegram WebApp initData string using HMAC-SHA256.
// It returns the authenticated Telegram user ID on success.
//
// Algorithm (per Telegram docs):
//  1. secret_key = HMAC-SHA256("WebAppData", bot_token)
//  2. data_check_string = sorted "key=value" pairs (excluding "hash"), joined by "\n"
//  3. expected_hash = HMAC-SHA256(secret_key, data_check_string) encoded as lowercase hex
//  4. Compare expected_hash with the hash field in initData
func ValidateInitData(botToken, initData string) (int64, error) {
	if initData == "" {
		return 0, ErrInvalidInitData
	}

	vals, err := url.ParseQuery(initData)
	if err != nil {
		return 0, ErrInvalidInitData
	}

	hash := vals.Get("hash")
	if hash == "" {
		return 0, ErrInvalidInitData
	}

	// Build the data_check_string: sorted "key=value" pairs excluding "hash".
	var pairs []string
	for k, v := range vals {
		if k == "hash" {
			continue
		}
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v[0]))
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	// secret_key = HMAC-SHA256("WebAppData", bot_token)
	secretKeyMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretKeyMAC.Write([]byte(botToken))
	secretKey := secretKeyMAC.Sum(nil)

	// expected_hash = HMAC-SHA256(secret_key, data_check_string)
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expectedHash), []byte(hash)) {
		return 0, ErrInvalidInitData
	}

	// Validate auth_date freshness.
	authDateStr := vals.Get("auth_date")
	if authDateStr == "" {
		return 0, ErrInvalidInitData
	}
	authDateUnix, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return 0, ErrInvalidInitData
	}
	if time.Since(time.Unix(authDateUnix, 0)) > initDataMaxAge {
		return 0, ErrInvalidInitData
	}

	// Extract user ID from the "user" field, which is a JSON object.
	// We parse it minimally to avoid importing encoding/json for a single field.
	userJSON := vals.Get("user")
	userID, err := extractUserID(userJSON)
	if err != nil {
		return 0, ErrInvalidInitData
	}
	return userID, nil
}

// extractUserID extracts the "id" field from a Telegram user JSON string.
// Example input: {"id":123456789,"first_name":"John","username":"john"}
func extractUserID(userJSON string) (int64, error) {
	if userJSON == "" {
		return 0, errors.New("missing user field")
	}
	// Find `"id":` and parse the following integer.
	const prefix = `"id":`
	idx := strings.Index(userJSON, prefix)
	if idx == -1 {
		return 0, errors.New("id field not found in user JSON")
	}
	rest := strings.TrimSpace(userJSON[idx+len(prefix):])
	// Read digits until a non-digit character.
	end := 0
	for end < len(rest) && rest[end] >= '0' && rest[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0, errors.New("could not parse user id")
	}
	return strconv.ParseInt(rest[:end], 10, 64)
}

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

// TelegramUser holds the minimal user profile extracted from Telegram initData.
type TelegramUser struct {
	ID           int64
	Username     string
	FirstName    string
	LastName     string
	LanguageCode string
}

// ValidateInitData validates the Telegram WebApp initData string using HMAC-SHA256.
// It returns the authenticated Telegram user profile on success.
//
// Algorithm (per Telegram docs):
//  1. secret_key = HMAC-SHA256("WebAppData", bot_token)
//  2. data_check_string = sorted "key=value" pairs (excluding "hash"), joined by "\n"
//  3. expected_hash = HMAC-SHA256(secret_key, data_check_string) encoded as lowercase hex
//  4. Compare expected_hash with the hash field in initData
func ValidateInitData(botToken, initData string) (TelegramUser, error) {
	if initData == "" {
		return TelegramUser{}, ErrInvalidInitData
	}

	vals, err := url.ParseQuery(initData)
	if err != nil {
		return TelegramUser{}, ErrInvalidInitData
	}

	hash := vals.Get("hash")
	if hash == "" {
		return TelegramUser{}, ErrInvalidInitData
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
		return TelegramUser{}, ErrInvalidInitData
	}

	// Validate auth_date freshness.
	authDateStr := vals.Get("auth_date")
	if authDateStr == "" {
		return TelegramUser{}, ErrInvalidInitData
	}
	authDateUnix, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return TelegramUser{}, ErrInvalidInitData
	}
	if time.Since(time.Unix(authDateUnix, 0)) > initDataMaxAge {
		return TelegramUser{}, ErrInvalidInitData
	}

	// Extract user profile from the "user" field (JSON object).
	userJSON := vals.Get("user")
	tgUser, err := extractTelegramUser(userJSON)
	if err != nil {
		return TelegramUser{}, ErrInvalidInitData
	}
	return tgUser, nil
}

// extractTelegramUser parses user profile fields from a Telegram user JSON string.
// Example input: {"id":123456789,"first_name":"John","last_name":"Doe","username":"john"}
func extractTelegramUser(userJSON string) (TelegramUser, error) {
	if userJSON == "" {
		return TelegramUser{}, errors.New("missing user field")
	}

	// Parse id (integer).
	id, err := extractJSONInt(userJSON, "id")
	if err != nil {
		return TelegramUser{}, err
	}

	return TelegramUser{
		ID:           id,
		FirstName:    extractJSONString(userJSON, "first_name"),
		LastName:     extractJSONString(userJSON, "last_name"),
		Username:     extractJSONString(userJSON, "username"),
		LanguageCode: extractJSONString(userJSON, "language_code"),
	}, nil
}

// extractJSONInt extracts an integer field from a flat JSON string without importing encoding/json.
func extractJSONInt(json, key string) (int64, error) {
	prefix := `"` + key + `":`
	idx := strings.Index(json, prefix)
	if idx == -1 {
		return 0, fmt.Errorf("field %q not found", key)
	}
	rest := strings.TrimSpace(json[idx+len(prefix):])
	end := 0
	for end < len(rest) && rest[end] >= '0' && rest[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0, fmt.Errorf("could not parse integer field %q", key)
	}
	return strconv.ParseInt(rest[:end], 10, 64)
}

// extractJSONString extracts a string field from a flat JSON string without importing encoding/json.
// Returns an empty string if the field is absent.
func extractJSONString(json, key string) string {
	prefix := `"` + key + `":"`
	idx := strings.Index(json, prefix)
	if idx == -1 {
		return ""
	}
	rest := json[idx+len(prefix):]
	end := strings.IndexByte(rest, '"')
	if end == -1 {
		return ""
	}
	return rest[:end]
}

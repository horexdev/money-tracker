package testutil

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestLogger returns a logger that discards all output, suitable for tests.
func TestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// MustDecodeJSON decodes JSON from r into T, failing the test on error.
func MustDecodeJSON[T any](t *testing.T, r io.Reader) T {
	t.Helper()
	var v T
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	return v
}

// BuildValidInitData constructs a valid Telegram initData string signed with botToken.
// This mirrors the algorithm in internal/api/auth.go exactly.
func BuildValidInitData(botToken string, userID int64, firstName string, authDate time.Time) string {
	userJSON := fmt.Sprintf(`{"id":%d,"first_name":"%s"}`, userID, firstName)

	vals := url.Values{}
	vals.Set("auth_date", strconv.FormatInt(authDate.Unix(), 10))
	vals.Set("user", userJSON)

	var pairs []string
	for k, v := range vals {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v[0]))
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secretKeyMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretKeyMAC.Write([]byte(botToken))
	secretKey := secretKeyMAC.Sum(nil)

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(mac.Sum(nil))

	vals.Set("hash", hash)
	return vals.Encode()
}

package backup

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEmptyBackupMatchesAndroidV1FixtureShape(t *testing.T) {
	data, err := EncodeBackup(NewEmptyBackup(0))
	if err != nil {
		t.Fatalf("EncodeBackup() error = %v", err)
	}

	want := `{
  "format": "MoneyTrackerBackup",
  "version": 1,
  "created_at_epoch_millis": 0,
  "profiles": [],
  "exchange_rate_snapshots": []
}`
	if strings.TrimSpace(string(data)) != want {
		t.Fatalf("encoded empty backup mismatch\nwant:\n%s\n\ngot:\n%s", want, string(data))
	}
}

func TestValidateBackupRejectsInvalidRefs(t *testing.T) {
	backup := NewEmptyBackup(1)
	backup.Profiles = append(backup.Profiles, BackupProfile{Ref: "source:123"})

	err := ValidateBackup(backup)
	if err == nil {
		t.Fatal("ValidateBackup() error = nil, want forbidden ref error")
	}
	if !strings.Contains(err.Error(), "forbidden export ref part") {
		t.Fatalf("ValidateBackup() error = %q, want forbidden ref part", err.Error())
	}
}

func collectJSONKeys(t *testing.T, data []byte) map[string]struct{} {
	t.Helper()
	var root any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	keys := map[string]struct{}{}
	collectKeys(root, keys)
	return keys
}

func collectKeys(value any, keys map[string]struct{}) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			keys[key] = struct{}{}
			collectKeys(child, keys)
		}
	case []any:
		for _, child := range typed {
			collectKeys(child, keys)
		}
	}
}

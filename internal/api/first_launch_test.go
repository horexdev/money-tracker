package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/horexdev/money-tracker/internal/api"
	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/horexdev/money-tracker/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func jsonBody(s string) io.Reader { return bytes.NewBufferString(s) }

// allLangs is the full set of Telegram language codes supported by the app.
var allLangs = []string{
	"en", "ru", "uk", "be", "kk", "uz",
	"es", "de", "it", "fr", "pt", "nl",
	"ar", "tr", "ko", "ms", "id",
}

// expectedAccount maps each language to the expected (accountName, currencyCode)
// that ensureUser must produce on first login.
var expectedAccount = map[string][2]string{
	"en": {"Main account", "USD"},
	"ru": {"Основной счёт", "RUB"},
	"uk": {"Основний рахунок", "UAH"},
	"be": {"Асноўны рахунак", "BYN"},
	"kk": {"Негізгі шот", "KZT"},
	"uz": {"Asosiy hisob", "UZS"},
	"es": {"Cuenta principal", "EUR"},
	"de": {"Hauptkonto", "EUR"},
	"it": {"Conto principale", "EUR"},
	"fr": {"Compte principal", "EUR"},
	"pt": {"Conta principal", "BRL"},
	"nl": {"Hoofdrekening", "EUR"},
	"ar": {"الحساب الرئيسي", "SAR"},
	"tr": {"Ana hesap", "TRY"},
	"ko": {"주 계좌", "KRW"},
	"ms": {"Akaun utama", "MYR"},
	"id": {"Rekening utama", "IDR"},
}

// TestFirstLaunch_LocalizationTables verifies that the lookup tables embedded
// in server.go produce the correct name and currency for every supported language.
// This is a pure-logic test — no mocks, no HTTP.
func TestFirstLaunch_LocalizationTables(t *testing.T) {
	for _, lang := range allLangs {
		lang := lang
		t.Run(lang, func(t *testing.T) {
			want, ok := expectedAccount[lang]
			require.True(t, ok, "missing expected entry for lang %q", lang)

			assert.Equal(t, want[0], api.LocalizedAccountName(lang),
				"account name mismatch for lang %q", lang)
			assert.Equal(t, want[1], api.LocalizedAccountCurrency(lang),
				"currency mismatch for lang %q", lang)
		})
	}
}

// TestFirstLaunch_UnknownLang verifies the fallback behaviour for an unrecognised
// Telegram language code (e.g. "zh").
func TestFirstLaunch_UnknownLang(t *testing.T) {
	assert.Equal(t, "Main account", api.LocalizedAccountName("zh"))
	assert.Equal(t, "USD", api.LocalizedAccountCurrency("zh"))
	assert.Equal(t, "Main account", api.LocalizedAccountName(""))
	assert.Equal(t, "USD", api.LocalizedAccountCurrency(""))
}

// TestFirstLaunch_EnsureUser_CreatesDefaultAccount verifies that ensureUser
// creates a default account with the correct localised name and currency for
// each supported language when the user has no accounts yet.
func TestFirstLaunch_EnsureUser_CreatesDefaultAccount(t *testing.T) {
	for _, lang := range allLangs {
		lang := lang
		want := expectedAccount[lang]

		t.Run(lang, func(t *testing.T) {
			t.Parallel()

			userRepo := &mocks.MockUserStorer{}
			accountRepo := &mocks.MockAccountStorer{}

			user := &domain.User{ID: 1, CurrencyCode: "USD"}
			userRepo.On("Upsert", mock.Anything, mock.AnythingOfType("*domain.User")).
				Return(user, nil)
			// No existing accounts → triggers account creation.
			accountRepo.On("ListByUser", mock.Anything, int64(1)).
				Return([]*domain.Account{}, nil)
			accountRepo.On("Create", mock.Anything, mock.MatchedBy(func(a *domain.Account) bool {
				return a.Name == want[0] && a.CurrencyCode == want[1]
			})).Return(&domain.Account{ID: 1, Name: want[0], CurrencyCode: want[1]}, nil)

			userSvc := service.NewUserService(userRepo, testutil.TestLogger())
			accountSvc := service.NewAccountService(accountRepo, nil, testutil.TestLogger())

			ue := api.NewUserEnsurer(userSvc, accountSvc, testutil.TestLogger())
			err := api.EnsureUserForTest(ue, context.Background(), api.TelegramUserForTest(1, "Test", lang))
			require.NoError(t, err)

			accountRepo.AssertCalled(t, "Create", mock.Anything, mock.MatchedBy(func(a *domain.Account) bool {
				return a.Name == want[0] && a.CurrencyCode == want[1]
			}))
		})
	}
}

// TestFirstLaunch_EnsureUser_SkipsAccountIfExists verifies that ensureUser
// does NOT create a second account when the user already has one.
func TestFirstLaunch_EnsureUser_SkipsAccountIfExists(t *testing.T) {
	userRepo := &mocks.MockUserStorer{}
	accountRepo := &mocks.MockAccountStorer{}

	user := &domain.User{ID: 1, CurrencyCode: "USD"}
	userRepo.On("Upsert", mock.Anything, mock.AnythingOfType("*domain.User")).
		Return(user, nil)
	accountRepo.On("ListByUser", mock.Anything, int64(1)).
		Return([]*domain.Account{{ID: 99, Name: "Existing", CurrencyCode: "USD"}}, nil)
	accountRepo.On("GetBalanceInBase", mock.Anything, int64(99), int64(1)).Return(int64(0), nil)

	userSvc := service.NewUserService(userRepo, testutil.TestLogger())
	accountSvc := service.NewAccountService(accountRepo, nil, testutil.TestLogger())

	ue := api.NewUserEnsurer(userSvc, accountSvc, testutil.TestLogger())
	err := api.EnsureUserForTest(ue, context.Background(), api.TelegramUserForTest(1, "Test", "en"))
	require.NoError(t, err)

	accountRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// TestFirstLaunch_SettingsLanguage verifies that PATCH /api/v1/settings
// accepts every supported language and rejects unknown codes.
func TestFirstLaunch_SettingsLanguage(t *testing.T) {
	tests := []struct {
		lang     string
		wantOK   bool
		wantLang string
	}{
		{lang: "en", wantOK: true, wantLang: "en"},
		{lang: "ru", wantOK: true, wantLang: "ru"},
		{lang: "uk", wantOK: true, wantLang: "uk"},
		{lang: "be", wantOK: true, wantLang: "be"},
		{lang: "kk", wantOK: true, wantLang: "kk"},
		{lang: "uz", wantOK: true, wantLang: "uz"},
		{lang: "es", wantOK: true, wantLang: "es"},
		{lang: "de", wantOK: true, wantLang: "de"},
		{lang: "it", wantOK: true, wantLang: "it"},
		{lang: "fr", wantOK: true, wantLang: "fr"},
		{lang: "pt", wantOK: true, wantLang: "pt"},
		{lang: "nl", wantOK: true, wantLang: "nl"},
		{lang: "ar", wantOK: true, wantLang: "ar"},
		{lang: "tr", wantOK: true, wantLang: "tr"},
		{lang: "ko", wantOK: true, wantLang: "ko"},
		{lang: "ms", wantOK: true, wantLang: "ms"},
		{lang: "id", wantOK: true, wantLang: "id"},
		{lang: "zh", wantOK: false}, // unsupported
		{lang: "xx", wantOK: false}, // unknown
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.lang, func(t *testing.T) {
			userRepo := &mocks.MockUserStorer{}
			user := &domain.User{ID: 1, CurrencyCode: "USD", Language: "en"}
			userRepo.On("GetByID", mock.Anything, int64(1)).Return(user, nil)

			if tt.wantOK {
				updated := &domain.User{ID: 1, CurrencyCode: "USD", Language: domain.Language(tt.wantLang)}
				userRepo.On("UpdateLanguage", mock.Anything, int64(1), tt.lang).Return(updated, nil)
			}

			svc := service.NewUserService(userRepo, testutil.TestLogger())
			h := api.SettingsHandlerForTest(svc, 0, testutil.TestLogger())

			body := `{"language":"` + tt.lang + `"}`
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPatch, "/api/v1/settings", jsonBody(body))
			r = r.WithContext(api.WithUserID(r.Context(), 1))
			h.ServeHTTP(w, r)

			if tt.wantOK {
				assert.Equal(t, http.StatusOK, w.Code)
				var resp map[string]any
				require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				assert.Equal(t, tt.wantLang, resp["language"])
			} else {
				assert.Equal(t, http.StatusBadRequest, w.Code)
			}
		})
	}
}

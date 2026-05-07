package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/go-telegram/bot/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/horexdev/money-tracker/internal/testutil/mocks"
)

func TestGetString_KnownLanguage(t *testing.T) {
	en := getString(domain.LangEN)
	assert.NotEmpty(t, en.welcome, "english welcome must be defined")
	assert.NotEmpty(t, en.help)
	assert.NotEmpty(t, en.openButton)

	ru := getString(domain.LangRU)
	assert.NotEqual(t, en.welcome, ru.welcome, "russian welcome must differ from english")
}

func TestGetString_FallsBackToEnglish(t *testing.T) {
	en := getString(domain.LangEN)
	got := getString(domain.Language("xx"))
	assert.Equal(t, en, got)
}

func TestGetString_AllSupportedLanguagesPresent(t *testing.T) {
	supported := []domain.Language{
		domain.LangEN, domain.LangRU, domain.LangUK, domain.LangBE, domain.LangKK,
		domain.LangUZ, domain.LangES, domain.LangDE, domain.LangIT, domain.LangFR,
		domain.LangPT, domain.LangNL, domain.LangAR, domain.LangTR, domain.LangKO,
		domain.LangMS, domain.LangID,
	}
	english := getString(domain.LangEN)
	for _, lang := range supported {
		s := getString(lang)
		assert.NotEmptyf(t, s.welcome, "lang %q welcome empty", lang)
		assert.NotEmptyf(t, s.help, "lang %q help empty", lang)
		assert.NotEmptyf(t, s.openButton, "lang %q openButton empty", lang)
		// EN fallback path returns english struct exactly only for invalid lang;
		// every supported lang must override at least the welcome string.
		if lang != domain.LangEN {
			assert.NotEqualf(t, english.welcome, s.welcome, "lang %q must localise welcome", lang)
		}
	}
}

func TestExtractTelegramUser_Message(t *testing.T) {
	user := &models.User{ID: 42, FirstName: "Alice"}
	upd := &models.Update{Message: &models.Message{From: user}}
	got := extractTelegramUser(upd)
	assert.Equal(t, user, got)
}

func TestExtractTelegramUser_CallbackQuery(t *testing.T) {
	upd := &models.Update{
		CallbackQuery: &models.CallbackQuery{From: models.User{ID: 100, FirstName: "Bob"}},
	}
	got := extractTelegramUser(upd)
	assert.Equal(t, int64(100), got.ID)
	assert.Equal(t, "Bob", got.FirstName)
}

func TestExtractTelegramUser_NilCases(t *testing.T) {
	assert.Nil(t, extractTelegramUser(&models.Update{}))
	assert.Nil(t, extractTelegramUser(&models.Update{Message: &models.Message{}}))
	assert.Nil(t, extractTelegramUser(&models.Update{CallbackQuery: &models.CallbackQuery{}}))
}

func TestExtractUserID(t *testing.T) {
	assert.Equal(t, int64(0), extractUserID(&models.Update{}))
	assert.Equal(t, int64(7), extractUserID(&models.Update{
		Message: &models.Message{From: &models.User{ID: 7}},
	}))
	assert.Equal(t, int64(99), extractUserID(&models.Update{
		CallbackQuery: &models.CallbackQuery{From: models.User{ID: 99}},
	}))
}

func TestUserLang_FallsBackOnError(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	repo.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("not found"))
	svc := service.NewUserService(repo, testutil.TestLogger())

	got := userLang(context.Background(), svc, 1)
	assert.Equal(t, domain.LangEN, got)
}

func TestUserLang_FallsBackOnEmptyLanguage(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	repo.On("GetByID", mock.Anything, int64(2)).Return(&domain.User{ID: 2, Language: ""}, nil)
	svc := service.NewUserService(repo, testutil.TestLogger())

	got := userLang(context.Background(), svc, 2)
	assert.Equal(t, domain.LangEN, got)
}

func TestUserLang_ReturnsStoredLanguage(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	repo.On("GetByID", mock.Anything, int64(3)).Return(&domain.User{ID: 3, Language: domain.LangRU}, nil)
	svc := service.NewUserService(repo, testutil.TestLogger())

	got := userLang(context.Background(), svc, 3)
	assert.Equal(t, domain.LangRU, got)
}

package api_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/horexdev/money-tracker/internal/api"
	"github.com/horexdev/money-tracker/internal/domain"
)

func TestDefaultCategoriesFor_KnownLanguage(t *testing.T) {
	en := api.DefaultCategoriesFor("en")
	assert.NotEmpty(t, en)
	// One of the seeds must be the savings category.
	hasSavings := false
	for _, s := range en {
		if s.Type == domain.CategoryTypeSavings {
			hasSavings = true
			break
		}
	}
	assert.True(t, hasSavings, "every locale must include a savings seed")
}

func TestDefaultCategoriesFor_FallsBackToEnglish(t *testing.T) {
	en := api.DefaultCategoriesFor("en")
	zz := api.DefaultCategoriesFor("zz") // unknown locale
	assert.Equal(t, en, zz, "unknown locale should fall back to English seeds")
}

func TestDefaultCategoriesFor_AllLocalesHaveSameLength(t *testing.T) {
	enLen := len(api.DefaultCategoriesFor("en"))
	for _, lang := range []string{"ru", "uk", "be", "kk", "uz", "es", "de", "it", "fr", "pt", "nl", "ar", "tr", "ko", "ms", "id"} {
		got := api.DefaultCategoriesFor(lang)
		assert.Lenf(t, got, enLen, "locale %q must define same number of seeds as English", lang)
	}
}

func TestDefaultCategoriesFor_AllLocalesShareIcons(t *testing.T) {
	en := api.DefaultCategoriesFor("en")
	for _, lang := range []string{"ru", "uk", "es", "de", "fr"} {
		got := api.DefaultCategoriesFor(lang)
		require := len(en) == len(got)
		if !require {
			continue
		}
		for i := range en {
			assert.Equalf(t, en[i].Icon, got[i].Icon, "icon mismatch at %d in %q", i, lang)
			assert.Equalf(t, en[i].Type, got[i].Type, "type mismatch at %d in %q", i, lang)
			assert.Equalf(t, en[i].Color, got[i].Color, "color mismatch at %d in %q", i, lang)
		}
	}
}

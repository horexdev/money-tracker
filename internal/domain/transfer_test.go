package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/horexdev/money-tracker/internal/domain"
)

func TestTransfer_FromAndToTxIDs_NilByDefault(t *testing.T) {
	tr := &domain.Transfer{}
	assert.Nil(t, tr.FromTxID, "transfers created before migration 00020 leave linked tx IDs nil")
	assert.Nil(t, tr.ToTxID)
}

func TestTransfer_LinkedTransactionIDs_PointToOwnedRows(t *testing.T) {
	fromTx := int64(101)
	toTx := int64(102)
	tr := &domain.Transfer{
		FromAccountID:    1,
		ToAccountID:      2,
		FromCurrencyCode: "USD",
		ToCurrencyCode:   "EUR",
		ExchangeRate:     0.92,
		FromTxID:         &fromTx,
		ToTxID:           &toTx,
	}

	assert.NotEqual(t, tr.FromAccountID, tr.ToAccountID, "transfer requires distinct accounts")
	assert.Equal(t, "USD", tr.FromCurrencyCode)
	assert.Equal(t, "EUR", tr.ToCurrencyCode)
	require := tr.FromTxID
	require2 := tr.ToTxID
	assert.NotNil(t, require)
	assert.NotNil(t, require2)
	assert.Equal(t, int64(101), *require)
	assert.Equal(t, int64(102), *require2)
}

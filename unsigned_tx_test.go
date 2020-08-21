package iotapkg_test

import (
	"errors"
	"testing"

	"github.com/luca-moser/iotapkg"
	"github.com/stretchr/testify/assert"
)

func TestTransactionSelector(t *testing.T) {
	_, err := iotapkg.TransactionSelector(100)
	assert.True(t, errors.Is(err, iotapkg.ErrUnknownTransactionType))
}

func TestUnsignedTransaction_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iotapkg.Serializable
		err    error
	}
	tests := []test{
		func() test {
			unTx, unTxData := randUnsignedTransaction()
			return test{"ok", unTxData, unTx, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &iotapkg.UnsignedTransaction{}
			bytesRead, err := tx.Deserialize(tt.source)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, tx)
		})
	}
}

func TestUnsignedTransaction_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotapkg.UnsignedTransaction
		target []byte
	}
	tests := []test{
		func() test {
			unTx, unTxData := randUnsignedTransaction()
			return test{"ok", unTx, unTxData}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize()
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}
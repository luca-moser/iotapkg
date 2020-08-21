package iotapkg_test

import (
	"errors"
	"testing"

	"github.com/luca-moser/iotapkg"
	"github.com/stretchr/testify/assert"
)

func TestSignatureSelector(t *testing.T) {
	_, err := iotapkg.SignatureSelector(100)
	assert.True(t, errors.Is(err, iotapkg.ErrUnknownSignatureType))
}

func TestEd25519Signature_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iotapkg.Serializable
		err    error
	}
	tests := []test{
		func() test {
			edSig, edSigData := randEd25519Signature()
			return test{"ok", edSigData, edSig, nil}
		}(),
		func() test {
			edSig, edSigData := randEd25519Signature()
			return test{"not enough data", edSigData[:5], edSig, iotapkg.ErrInvalidBytes}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edSig := &iotapkg.Ed25519Signature{}
			bytesRead, err := edSig.Deserialize(tt.source)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, edSig)
		})
	}
}

func TestEd25519Signature_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotapkg.Ed25519Signature
		target []byte
	}
	tests := []test{
		func() test {
			edSig, edSigData := randEd25519Signature()
			return test{"ok", edSig, edSigData}
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

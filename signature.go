package iotapkg

import (
	"crypto/ed25519"
	"fmt"
)

// Defines the type of signature.
type SignatureType = byte

const (
	// Denotes a WOTS a signature.
	SignatureWOTS SignatureType = iota
	// Denotes an Ed25519 signature.
	SignatureEd25519

	// The size of a serialized Ed25519 signature with its type denoting byte and public key.
	Ed25519SignatureSerializedBytesSize = OneByte + ed25519.PublicKeySize + ed25519.SignatureSize
)

// SignatureSelector implements SerializableSelectorFunc for signature types.
func SignatureSelector(typeByte byte) (Serializable, error) {
	var seri Serializable
	switch typeByte {
	case SignatureWOTS:
		seri = &WOTSSignature{}
	case SignatureEd25519:
		seri = &Ed25519Signature{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownSignatureType, typeByte)
	}
	return seri, nil
}

type WOTSSignature struct{}

func (w *WOTSSignature) Serialize() ([]byte, error) {
	panic("implement me")
}

func (w *WOTSSignature) Deserialize(data []byte) (int, error) {
	panic("implement me")
}

type Ed25519Signature struct {
	PublicKey [ed25519.PublicKeySize]byte `json:"public_key"`
	Signature [ed25519.SignatureSize]byte `json:"signature"`
}

func (e Ed25519Signature) Deserialize(data []byte) (int, error) {
	// skip type byte
	data = data[OneByte:]
	if err := checkExactByteLength(Ed25519SignatureSerializedBytesSize, len(data)); err != nil {
		return 0, err
	}
	copy(e.PublicKey[:], data[:ed25519.PublicKeySize])
	copy(e.Signature[:], data[ed25519.PublicKeySize:])
	return Ed25519AddressSerializedBytesSize, nil
}

func (e Ed25519Signature) Serialize() ([]byte, error) {
	var b [Ed25519AddressSerializedBytesSize]byte
	b[0] = SignatureEd25519
	copy(b[OneByte:], e.PublicKey[:])
	copy(b[OneByte+ed25519.PublicKeySize:], e.Signature[:])
	return b[:], nil
}
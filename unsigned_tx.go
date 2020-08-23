package iota

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// Defines the type of transaction.
type TransactionType = byte

const (
	// Denotes an unsigned transaction.
	TransactionUnsigned TransactionType = iota

	TransactionIDLength = 32
)

var (
	ErrInputsOrderViolatesLexicalOrder   = errors.New("inputs must be in their lexical order (byte wise) when serialized")
	ErrOutputsOrderViolatesLexicalOrder  = errors.New("outputs must be in their lexical order (byte wise) when serialized")
	ErrInputUTXORefsNotUnique            = errors.New("inputs must each reference a unique UTXO")
	ErrOutputAddrNotUnique               = errors.New("outputs must each deposit to a unique address")
	ErrOutputsSumExceedsTotalSupply      = errors.New("accumulated output balance exceeds total supply")
	ErrOutputDepositsMoreThanTotalSupply = errors.New("an output can not deposit more than the total supply")
)

// TransactionSelector implements SerializableSelectorFunc for transaction types.
func TransactionSelector(typeByte byte) (Serializable, error) {
	var seri Serializable
	switch typeByte {
	case TransactionUnsigned:
		seri = &UnsignedTransaction{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownTransactionType, typeByte)
	}
	return seri, nil
}

// UnsignedTransaction is the unsigned part of a transaction.
type UnsignedTransaction struct {
	// The inputs of this transaction.
	Inputs Serializables `json:"inputs"`
	// The outputs of this transaction.
	Outputs Serializables `json:"outputs"`
	// The optional embedded payload.
	Payload Serializable `json:"payload"`
}

func (u *UnsignedTransaction) Deserialize(data []byte, skipValidation bool) (int, error) {
	if !skipValidation {
		if err := checkType(data, TransactionUnsigned); err != nil {
			return 0, fmt.Errorf("unable to deserialize unsigned transaction: %w", err)
		}
	}

	// skip type byte
	bytesReadTotal := OneByte
	data = data[OneByte:]

	inputs, inputBytesRead, err := DeserializeArrayOfObjects(data, skipValidation, InputSelector, &inputsArrayBound)
	if err != nil {
		return 0, err
	}
	bytesReadTotal += inputBytesRead

	if !skipValidation {
		if err := ValidateInputs(inputs, InputsUTXORefsUniqueValidator()); err != nil {
			return 0, err
		}
	}

	u.Inputs = inputs

	// advance to outputs
	data = data[inputBytesRead:]
	outputs, outputBytesRead, err := DeserializeArrayOfObjects(data, skipValidation, OutputSelector, &outputsArrayBound)
	if err != nil {
		return 0, err
	}
	bytesReadTotal += outputBytesRead

	if !skipValidation {
		if err := ValidateOutputs(outputs, OutputsAddrUniqueValidator()); err != nil {
			return 0, err
		}
	}

	u.Outputs = outputs

	// advance to payload
	// TODO: replace with payload deserializer
	data = data[outputBytesRead:]
	payloadLength, payloadLengthByteSize, err := ReadUvarint(bytes.NewReader(data[:binary.MaxVarintLen64]))
	if err != nil {
		return 0, err
	}
	bytesReadTotal += payloadLengthByteSize

	if payloadLength == 0 {
		return bytesReadTotal, nil
	}

	// TODO: payload extraction logic
	data = data[payloadLengthByteSize:]
	switch data[0] {

	}
	bytesReadTotal += int(payloadLength)

	return bytesReadTotal, nil
}

func (u *UnsignedTransaction) Serialize(skipValidation bool) (data []byte, err error) {
	if !skipValidation {
		if err := ValidateInputs(u.Inputs, InputsUTXORefsUniqueValidator()); err != nil {
			return nil, err
		}
		if err := ValidateOutputs(u.Outputs, OutputsAddrUniqueValidator()); err != nil {
			return nil, err
		}
	}

	var b bytes.Buffer
	if err := b.WriteByte(TransactionUnsigned); err != nil {
		return nil, err
	}

	varIntBuf := make([]byte, binary.MaxVarintLen64)
	bytesWritten := binary.PutUvarint(varIntBuf, uint64(len(u.Inputs)))

	if _, err := b.Write(varIntBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	for i := range u.Inputs {
		inputSer, err := u.Inputs[i].Serialize(skipValidation)
		if err != nil {
			return nil, fmt.Errorf("unable to serialize input at index %d: %w", i, err)
		}
		if _, err := b.Write(inputSer); err != nil {
			return nil, err
		}
	}

	// reuse varIntBuf (this is safe as b.Write() copies the bytes)
	bytesWritten = binary.PutUvarint(varIntBuf, uint64(len(u.Outputs)))
	if _, err := b.Write(varIntBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	for i := range u.Outputs {
		outputSer, err := u.Outputs[i].Serialize(skipValidation)
		if err != nil {
			return nil, fmt.Errorf("unable to serialize output at index %d: %w", i, err)
		}
		if _, err := b.Write(outputSer); err != nil {
			return nil, err
		}
	}

	// no payload
	if u.Payload == nil {
		if err := b.WriteByte(0); err != nil {
			return nil, err
		}
		return b.Bytes(), nil
	}

	payloadSer, err := u.Payload.Serialize(skipValidation)
	if _, err := b.Write(payloadSer); err != nil {
		return nil, err
	}

	bytesWritten = binary.PutUvarint(varIntBuf, uint64(len(payloadSer)))
	if _, err := b.Write(varIntBuf[:bytesWritten]); err != nil {
		return nil, err
	}

	if _, err := b.Write(payloadSer); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// SyntacticallyValid checks whether the unsigned transaction is syntactically valid by checking whether:
//	1. every input references a unique UTXO and has valid UTXO index bounds
//	2. every output deposits to a unique address and deposits more than zero
//	3. the accumulated deposit output is not over the total supply
// The function does not syntactically validate the input or outputs themselves.
func (u *UnsignedTransaction) SyntacticallyValid() error {
	if err := ValidateInputs(u.Inputs,
		InputsUTXORefIndexBoundsValidator(),
		InputsUTXORefsUniqueValidator(),
	); err != nil {
		return err
	}

	if err := ValidateOutputs(u.Outputs,
		OutputsAddrUniqueValidator(),
		OutputsDepositAmountValidator(),
	); err != nil {
		return err
	}

	return nil
}

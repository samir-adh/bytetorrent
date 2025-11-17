package message

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestRead(t *testing.T) {
	// Expected Payload Data
	testPayload := []byte{0x13, 0xAD, 0xBE, 0xEF}
	testMessageID := byte(10)

	bodyLength := 1 + len(testPayload)

	lengthPrefix := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthPrefix, uint32(bodyLength))

	validMessage := bytes.NewBuffer(lengthPrefix)
	validMessage.WriteByte(testMessageID)
	validMessage.Write(testPayload)

	readOutput, err := Read(validMessage)
	if err != nil {
		t.Fatalf("test case failed with : %s", err)
	}

	if readOutput.Id != messageId(testMessageID) {
		t.Errorf("expected message id %d but got %d", testMessageID, readOutput.Id)
	}

	if readOutput.Length != uint32(bodyLength) {
		t.Errorf("expected body length of %d but got %d", bodyLength, readOutput.Length)
	}

	validPayload := len(readOutput.Payload) == len(testPayload)
	if !validPayload {
		t.Errorf("mismatch in payload size, expected %d but got %d", len(testPayload), len(readOutput.Payload))
	}
	for i := range readOutput.Payload {
		if readOutput.Payload[i] != testPayload[i] {
			t.Errorf("incorrect payload")
		}
	}

}

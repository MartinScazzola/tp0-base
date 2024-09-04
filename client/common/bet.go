package common

import (
	"fmt"
	"strconv"
	"strings"
	"net"
	"bytes"
)

type Bet struct {
	Agency     uint8 
	FirstName string 
	LastName  string 
	Document   uint32 
	Birthdate  string 
	Number     uint32 
}

func (b *Bet)toBytes()  []byte {
	var data []byte

	agency := b.Agency
	data = append(data, byte(agency))

	firstNameBytes := []byte(b.FirstName)
	data = append(data, byte(len(firstNameBytes)))
	data = append(data, firstNameBytes...)

	lastNameBytes := []byte(b.LastName)
	data = append(data, byte(len(lastNameBytes)))
	data = append(data, lastNameBytes...)

	data = append(data, byte(b.Document>>24), byte(b.Document>>16), byte(b.Document>>8), byte(b.Document))

	dateParts := strings.Split(b.Birthdate, "-")
	year, _ := strconv.Atoi(dateParts[0])
	month, _ := strconv.Atoi(dateParts[1])
	day, _ := strconv.Atoi(dateParts[2])

	data = append(data, byte(year>>8), byte(year))
	data = append(data, byte(month))
	data = append(data, byte(day))

	data = append(data, byte(b.Number>>24), byte(b.Number>>16), byte(b.Number>>8), byte(b.Number))

	return data
}

func beginSendBets(conn net.Conn, id string) error {
	id_int, _ := strconv.Atoi(id)

	bytes := append([]byte(BEGIN_SEND_BETS), byte(id_int), BATCH_END_CHAR, BATCH_END_CHAR)
	return safeWrite(conn, bytes)
}
func endSendBets(conn net.Conn) error {
	bytes := append([]byte(END_SEND_BETS), BATCH_END_CHAR, BATCH_END_CHAR)
	return safeWrite(conn, bytes)
}

func parseDocumentList(data []byte) []uint32 {
	var documents []uint32

	for i := 0; i < len(data); i += 4 {
		document := uint32(data[i])<<24 | uint32(data[i+1])<<16 | uint32(data[i+2])<<8 | uint32(data[i+3])
		documents = append(documents, document)
	}

	return documents
}

func receiveWinners(conn net.Conn) ([]uint32, error) {
	data, err := safeRead(conn)

	if err != nil {
		return nil, fmt.Errorf("Error receiving winners: %v", err)
	}

	return parseDocumentList(bytes.TrimRight(data, "||")), nil
}

func receiveConfirm(conn net.Conn) (string, error) {
	data, err := safeRead(conn)

	if err != nil {
		return "", fmt.Errorf("Error receiving confirmation: %v", err)
	}
	status := bytes.TrimRight(data, "||")

	return string(status), err
}

func batchToBytes(bets []Bet) []byte {
	var data []byte

	for _, bet := range bets {
		data = append(data, bet.toBytes()...)
	}

	return append(data, BATCH_END_CHAR, BATCH_END_CHAR)
}

func safeRead(conn net.Conn) ([]byte, error) {
	totalBytesRead := 0

	data := make([]byte, READ_BUFFER_SIZE)

	for totalBytesRead < READ_BUFFER_SIZE {
		bytesRead, err := conn.Read(data[totalBytesRead:])

		if err != nil || bytesRead == 0 {
			return nil, fmt.Errorf("Error reading from the server: %v", err)
		}

		totalBytesRead += bytesRead

		if data[totalBytesRead - 1] == byte(BATCH_END_CHAR) && data[totalBytesRead - 2] == byte(BATCH_END_CHAR) {
			break
		}
	}
	
	data = data[:totalBytesRead]

	return data, nil
}

func safeWrite(conn net.Conn, bytes []byte) error {
	totalBytesWritten := 0
	for totalBytesWritten < len(bytes) {
		bytesWritten, err := conn.Write(bytes[totalBytesWritten:])

		if err != nil || bytesWritten == 0 {
			return fmt.Errorf("Error sending the batch: %v", err)
		}

		totalBytesWritten += bytesWritten
	}
	return nil
}

func sendBetsBatch(conn net.Conn, bets []Bet) error {
	batchBytes := batchToBytes(bets)

	if len(batchBytes) > MAX_BATCH_SIZE {
		return fmt.Errorf("Batch too long; exceeds 8 kB\n")
	}

	err := safeWrite(conn, batchBytes)

	if err != nil {
		return fmt.Errorf("Error sending the batch: %v", err)
	}

	return nil
}


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

func endSendBets(conn net.Conn) error {
	msg := "END"
	bytes := append([]byte(msg), []byte("||")...)
	return safeWrite(conn, bytes)
}

func receiveConfirm(conn net.Conn) (string, error) {
	status, err := safeRead(conn)
	return string(status), err
}

func batchToBytes(bets []Bet) []byte {
	var data []byte

	for _, bet := range bets {
		data = append(data, bet.toBytes()...)
	}

	return append(data, '|', '|')
}

func safeRead(conn net.Conn) ([]byte, error) {
	totalBytesRead := 0

	data := make([]byte, 1024)

	for totalBytesRead < 1024 {
		bytesRead, err := conn.Read(data[totalBytesRead:])

		if err != nil || bytesRead == 0 {
			return nil, fmt.Errorf("Error reading from the server: %v", err)
		}

		totalBytesRead += bytesRead

		if data[totalBytesRead - 1] == byte('|') && data[totalBytesRead - 2] == byte('|') {
			break
		}
	}

	//fmt.Println("Total bytes read: %v", data)
	
	data = data[:totalBytesRead]

	return bytes.TrimRight(data, "||"), nil
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

	if len(batchBytes) > 8192 {
		return fmt.Errorf("Batch too long; exceeds 8 kB\n")
	}

	err := safeWrite(conn, batchBytes)

	if err != nil {
		return fmt.Errorf("Error sending the batch: %v", err)
	}

	return nil
}


package common

import (
	"fmt"
	"strconv"
	"strings"
	"net"
)

type Bet struct {
	Agency     uint8
	FirstName string
	LastName  string
	Document   uint32
	Birthdate  string
	Number     uint32
}

func sendBet(b Bet, conn net.Conn)  error {
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

	err := safeWrite(conn, append(data, byte('|')))

	if err != nil {
		return fmt.Errorf("failed to send the bet: %v", err)
	}

	return nil
}

func receiveConfirm(conn net.Conn, bet Bet) error {
	data, err := safeRead(conn, 2)

	if err != nil {
		return fmt.Errorf("failed to receive the server response: %v", err)
	}

	msg := string(data)

	if msg != "OK" {
		return fmt.Errorf("failed to receive the server response: %v", err)
	}
	return nil
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

func safeRead(conn net.Conn, sizeToRead int) ([]byte, error) {
	totalBytesRead := 0
	data := make([]byte, sizeToRead)

	for totalBytesRead < sizeToRead {
		bytesRead, err := conn.Read(data[totalBytesRead:])

		if err != nil || bytesRead == 0 {
			return nil, fmt.Errorf("Error reading from the server: %v", err)
		}

		totalBytesRead += bytesRead
	}

	return data, nil
}
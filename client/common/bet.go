package common

import (
	"fmt"
	"strconv"
	"strings"
	"net"
	"bufio"
	"bytes"
	"encoding/binary"
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

	data = append([]byte{byte(len(data) >> 8), byte(len(data))}, data...)

	return data
}

func receiveConfirm(conn net.Conn, bet Bet) error {
	msg, err := bufio.NewReader(conn).ReadString('\n')

	if err != nil && msg != "OK" {
		return fmt.Errorf("failed to receive the server response: %v", err)
	}
	log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v", bet.Document, bet.Number)
	return nil
}

func batchToBytes(bets []Bet) []byte {
	var data []byte

	for _, bet := range bets {
		data = append(data, bet.toBytes()...)
	}

	return data
}


func sendBetsBatchs(conn net.Conn, bets []Bet, batchSize int) error {
	lastBetSent := 0

	for lastBetSent < len(bets) {
		if lastBetSent+batchSize > len(bets) {
			batchSize = len(bets) - lastBetSent
		}

		batchBytes := batchToBytes(bets[lastBetSent : lastBetSent+batchSize])

		fmt.Println("Batch size: ", len(batchBytes))
		if len(batchBytes) > 8192 {
			return fmt.Errorf("Batch too long; exceeds 8 kB\n")
		}

		var sizeBuffer bytes.Buffer
		if err := binary.Write(&sizeBuffer, binary.BigEndian, uint16(len(batchBytes))); err != nil {
			return fmt.Errorf("Error converting batch size to bytes: %v", err)
		}

		_, err := conn.Write(append(sizeBuffer.Bytes(), batchBytes...))
		if err != nil {
			return fmt.Errorf("Error sending the batch: %v", err)
		}

		msg, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			return fmt.Errorf("Error reading server response: %v", err)
		}


		if strings.TrimSpace(msg) != "OK" {
			return fmt.Errorf("Batch failed: %s", msg)
		}

		log.Infof("action: apuestas_enviadas | result: success | batch_size: %v", batchSize)

		lastBetSent += batchSize
	}

	return nil
}
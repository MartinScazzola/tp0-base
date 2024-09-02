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

func endSendBets(conn net.Conn) error {
	msg := "END"
	bytes := []byte(msg)
	data := append([]byte{byte(len(bytes))}, bytes...)
	_ , err := conn.Write(data)
	if err != nil {
		return fmt.Errorf("Error sending end message: %v", err)
	}

	return nil
}

func receiveConfirm(conn net.Conn) (string, error) {
	return bufio.NewReader(conn).ReadString('\n')
}

func batchToBytes(bets []Bet) []byte {
	var data []byte

	for _, bet := range bets {
		data = append(data, bet.toBytes()...)
	}

	return data
}


func sendBetsBatch(conn net.Conn, bets []Bet) error {
	batchBytes := batchToBytes(bets)

	if len(batchBytes) > 8192 {
		return fmt.Errorf("Batch too long; exceeds 8 kB\n")
	}

	var sizeBuffer bytes.Buffer
	if err := binary.Write(&sizeBuffer, binary.BigEndian, uint16(len(batchBytes))); err != nil {
		return fmt.Errorf("Error converting batch size to bytes: %v", err)
	}


	msg := "BATCH"
	msgBytes := []byte(msg)

	dataMsg := append([]byte{byte(len(msgBytes))}, msgBytes...)

	dataBatch := append(sizeBuffer.Bytes(), batchBytes...)

	fmt

	_, err := conn.Write(append(dataMsg, dataBatch...))
	if err != nil {
		return fmt.Errorf("Error sending the batch: %v", err)
	}

	status, err := receiveConfirm(conn)

	if err != nil && status != "OK" {
		return fmt.Errorf("Error receiving confirmation: %v", err)
	}

	return nil
}


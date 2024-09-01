package common

import (
	"fmt"
	"strconv"
	"strings"
	"net"
	"bufio"
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

	data = append([]byte{byte(len(data) >> 8), byte(len(data))}, data...)

	if _, err := conn.Write(data); err != nil {
		return err
	}

	return nil
}

func receiveConfirm(conn net.Conn, bet Bet) error {
	msg, err := bufio.NewReader(conn).ReadString('\n')

	if err != nil && msg != "OK" {
		return fmt.Errorf("failed to receive the server response: %v", err)
	}
	log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v", bet.Document, bet.Number)
	return nil
}
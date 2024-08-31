package common

import (
	"fmt"
	"strconv"
	"strings"
)

type Bet struct {
	Agency     uint8
	FirstName string
	LastName  string
	Document   uint32
	Birthdate  string
	Number     uint32
}

func (b *Bet) toBytes() []byte {
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

	return append([]byte{byte(len(data) >> 8), byte(len(data))}, data...)
}

func fromBytes(data []byte) *Bet {
	index := 0

	agency := uint8(data[index])
	index++

	firstNameLength := int(data[index])
	index++
	firstName := string(data[index : index+firstNameLength])
	index += firstNameLength

	lastNameLength := int(data[index])
	index++
	lastName := string(data[index : index+lastNameLength])
	index += lastNameLength

	document := uint32(data[index])<<24 | uint32(data[index+1])<<16 | uint32(data[index+2])<<8 | uint32(data[index+3])
	index += 4

	year := int(data[index])<<8 | int(data[index+1])
	index += 2
	month := int(data[index])
	index++
	day := int(data[index])
	index++

	birthDate := fmt.Sprintf("%04d-%02d-%02d", year, month, day)

	number := uint32(data[index])<<24 | uint32(data[index+1])<<16 | uint32(data[index+2])<<8 | uint32(data[index+3])

	return &Bet{
		Agency:     agency,
		FirstName: firstName,
		LastName:  lastName,
		Document:   document,
		Birthdate:  birthDate,
		Number:     number,
	}
}

func main() {
	bet := Bet{123,"Juan", "Perez", 12345678, "1990-01-01", 42}
	fmt.Println("bet original", bet)

	bytes := bet.toBytes()
	fmt.Printf("bytes %v\n", bytes)

	length := int(bytes[0])<<8 | int(bytes[1])
	betRecv := fromBytes(bytes[2:length+2])

	fmt.Println("bet_recv", *betRecv)
}
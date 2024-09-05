package common

import (
	"encoding/csv"
	"fmt"
	"github.com/op/go-logging"
	"io"
	"net"
	"os"
	"strconv"
	"time"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	BetsFile      string
	BatchSize     int
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

func (c *Client) CleanUp() {
	if c.conn != nil {
		time.Sleep(5 * time.Second)
		c.conn.Close()
		log.Infof("Client connection closed")
	}
}

func (c *Client) getBetsFromFile(batchSize int, lastBetSent int) ([]Bet, error) {
	/*
	   Retrieves a batch of bets from a file.

	   Opens a CSV file specified in the client's configuration and reads bet data, starting
	   from the specified position (`lastBetSent`). It reads up to the specified `batchSize`
	   number of bets and returns them. If an error occurs while opening or reading the file,
	   the function returns an error.
	*/
	file, err := os.Open(c.config.BetsFile)
	if err != nil {
		return nil, fmt.Errorf("could not open file %s: %v", c.config.BetsFile, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	var data []Bet
	lineNumber := 0

	for {
		line, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading line: %v", err)
		}

		if len(line) == 0 {
			continue
		}

		if lineNumber < lastBetSent {
			lineNumber++
			continue
		}

		if len(data) < batchSize {
			document, err := strconv.ParseUint(line[2], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("could not parse document number: %v", err)
			}

			number, err := strconv.ParseUint(line[4], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("could not parse number: %v", err)
			}

			bet := Bet{
				FirstName: line[0],
				LastName:  line[1],
				Document:  uint32(document),
				Birthdate: line[3],
				Number:    uint32(number),
			}
			data = append(data, bet)
		}

		lineNumber++

		if len(data) >= batchSize {
			break
		}
	}

	return data, nil
}

func (c *Client) StartClientSendBetsLoop(stopChan chan os.Signal) error {
	/*
	   Main loop to send bets from a client to a server.

	   Initializes the client socket and starts sending bets in batches read from a file.
	   The loop continues until all bets are sent or a stop signal is received. Each batch
	   of bets is sent, and the client waits for a confirmation response before sending the
	   next batch. If an error occurs during sending or receiving, the function exits with an error.
	*/
	c.createClientSocket()

	lastBetSent := 0
	beginSendBets(c.conn, c.config.ID)

loop:
	for {
		select {
		case <-stopChan:
			log.Infof("action: loop_stopped | result: success | client_id: %v", c.config.ID)
			c.conn.Close()
			os.Exit(0)
		default:

			betsBatch, err := c.getBetsFromFile(c.config.BatchSize, lastBetSent)

			if err != nil {
				return fmt.Errorf("Error getting bets from file: %v", err)
			}

			if len(betsBatch) == 0 {
				break loop
			}

			err = sendBetsBatch(c.conn, betsBatch)

			if err != nil {
				return fmt.Errorf("Error sending bets: %v", err)
			}

			status, err := receiveConfirm(c.conn)

			if err != nil {
				return fmt.Errorf("Error receiving confirmation: %v", err)
			}

			if status == BATCH_SENT_OK {
				log.Infof("action: apuestas_enviadas | result: success | cantidad: %v", len(betsBatch))
			} else if status == BATCH_SENT_FAIL {
				log.Infof("action: apuestas_enviadas | result: fail")
			}

			lastBetSent += c.config.BatchSize

			time.Sleep(c.config.LoopPeriod)
		}
	}

	err := endSendBets(c.conn)

	if err != nil {
		return fmt.Errorf("Error ending send bets: %v", err)
	}

	winners, err := receiveWinners(c.conn)

	if err != nil {
		return fmt.Errorf("Error receiving winners: %v", err)
	}

	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", len(winners))

	c.CleanUp()
	return nil
}

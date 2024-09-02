package common

import (
	"net"
	"time"
	"os"
	"github.com/op/go-logging"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	BetsFile	  string
	BatchSize 	  int
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

func (c *Client) getBetsFromFile(batchSize int, lastBetSent int) ([]Bet, error) {

	id, err := strconv.ParseUint(c.config.ID, 10, 8)
	if err != nil {
		return nil, fmt.Errorf("Could not parse client ID: %s", err)
	}
	
	
	fmt.Printf("Getting bets from file: %s\n", c.config.BetsFile)

	file, err := os.Open(c.config.BetsFile)
	if err != nil {
		log.Fatalf("Could not open file %s: %s", c.config.BetsFile, err)
	}
	defer file.Close()
	fmt.Println("File opened successfully")

	reader := csv.NewReader(file)

	var data []Bet

	i := 0

	for {

		if i < lastBetSent {
			i++
			continue
		}

		line, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				fmt.Println("End of file reached")
				break
			}
			log.Fatalf("Error reading line: %s", err)
		}

		if len(line) == 0 {
			fmt.Println("Encountered an empty line.")
			continue
		}

		document, err := strconv.ParseUint(line[2], 10, 32)

		if err != nil {
			return nil, fmt.Errorf("Could not parse document number: %s", err)
		}
		
		number, err := strconv.ParseUint(line[4], 10, 32)

		if err != nil {
			return nil, fmt.Errorf("Could not parse number: %s", err)
		}

		bet := Bet{uint8(id),line[0], line[1], uint32(document), line[3], uint32(number)}

		data = append(data, bet)

		i++
	}

	return data, nil
}


func (c *Client) StartClientLoop(stopChan chan os.Signal) error {
	c.createClientSocket()

	lastBetSent := 0

	for {
		select {
		case <-stopChan:
			return nil
		default:
			betsBatch, err := c.getBetsFromFile(c.config.BatchSize, lastBetSent)

			if err != nil {
				return fmt.Errorf("Error getting bets from file: %v", err)
			}

			if len(betsBatch) == 0 {
				log.Infof("No more bets to send")
				break
			}

			err = sendBetsBatch(c.conn, betsBatch)

			if err == nil {
				log.Infof("action: apuestas_enviadas | result: success ")
			} else {
				log.Infof("action: apuestas_enviadas | result: fail")
			}
			
			lastBetSent += c.config.BatchSize

			time.Sleep(c.config.LoopPeriod)
		}
	}

	endSendBets(c.conn)
	c.CleanUp()
	return nil
}

func (c *Client) CleanUp() {
	if c.conn != nil {
		c.conn.Close()
		log.Infof("Client connection closed")
	}
}
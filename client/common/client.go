package common

import (
	"fmt"
	"net"
	"time"
	"os"
	"strconv"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
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

func (c *Client) getBetFromEnv() (Bet, error) {
	document, err := strconv.ParseUint(os.Getenv("DOCUMENTO"), 10, 32)
	if err != nil {
		return Bet{}, fmt.Errorf("error al convertir DOCUMENTO a uint32: %v", err)
	}

	number, err := strconv.ParseUint(os.Getenv("NUMERO"), 10, 32)
	if err != nil {
		return Bet{}, fmt.Errorf("error al convertir NUMERO a uint32: %v", err)
	}

	agency, err := strconv.ParseUint(os.Getenv("CLI_ID"), 10, 8)
	if err != nil {
		return Bet{}, fmt.Errorf("error al convertir NUMERO a uint32: %v", err)
	}

	bet := Bet{
		Agency:    uint8(agency),
		FirstName:    os.Getenv("NOMBRE"),
		LastName:  os.Getenv("APELLIDO"),
		Document:       uint32(document),
		Birthdate: os.Getenv("NACIMIENTO"),
		Number:    uint32(number),
	}

	return bet, nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(stopChan chan os.Signal) error {

	bet, err := c.getBetFromEnv()

	if err != nil {
		return fmt.Errorf("Failed to get bet from environment variables: %v", err)
	}

	if err := c.createClientSocket(); err != nil {
		return fmt.Errorf("Failed to connect to the server: %v", err)
	}

	if err := sendBet(bet, c.conn); err != nil {
		return fmt.Errorf("Failed to send the bet: %v", err)
	}

	if err := receiveConfirm(c.conn, bet); err != nil {
		return fmt.Errorf("Failed to receive the server response: %v", err)
	}

	log.Infof("action: apuesta_enviada | result: success | dni: %d | numero: %d", bet.Document, bet.Number)

	c.CleanUp()

	return nil
}

func (c *Client) CleanUp() {
	if c.conn != nil {
		time.Sleep(3 * time.Second)
		c.conn.Close()
		log.Infof("Client connection closed")
	}
}
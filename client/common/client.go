package common

import (
	"bufio"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopLapse     time.Duration
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
		log.Fatalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// Send a message to the server.
// This function avoids the short write problem
func (c *Client) sendMessage(msg string) {
	writer := bufio.NewWriter(c.conn)
	_, err := writer.WriteString(msg)
	if err != nil {
		log.Fatalf("action: send_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	err = writer.Flush()
	if err != nil {
		log.Fatalf("action: flush_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
}

func (_c *Client) initializeSignalReceiver() chan os.Signal {
	signalReceiver := make(chan os.Signal, 1)
	signal.Notify(signalReceiver, syscall.SIGTERM)
	return signalReceiver
}

func (c *Client) handleShutdown(signalReceiver chan os.Signal) {
	c.conn.Close()
	log.Infof("action: socket_closing | result: success | client_id: %v",
		c.config.ID,
	)

	close(signalReceiver)
	log.Infof("action: signal_receiver_channel_shutdown | result: success | client_id: %v",
		c.config.ID,
	)
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(testBet Bet) {
	signalReceiver := c.initializeSignalReceiver()

	// autoincremental msgToSendID to identify every message sent
	msgToSendID := 1

loop:
	// Send messages if the loopLapse threshold has not been surpassed
	for timeout := time.After(c.config.LoopLapse); ; {
		select {
		case <-timeout:
			log.Infof("action: timeout_detected | result: success | client_id: %v",
				c.config.ID,
			)
			break loop
		case <-signalReceiver:
			c.handleShutdown(signalReceiver)
			break loop

		default:
		}

		// Create the connection the server in every loop iteration. Send an
		c.createClientSocket()

		msgToSend := testBet.ToString() + DELIMITER
		c.sendMessage(msgToSend)

		receivedMsg, err := bufio.NewReader(c.conn).ReadString(DELIMITER[0])
		msgToSendID++
		c.conn.Close()

		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		if receivedMsg == msgToSend {
			log.Infof("action: apuesta_enviada | result: success | dni: %d | numero: %d",
				testBet.PlayerDocID,
				testBet.WageredNumber,
			)
			return
		} else {
			log.Errorf("action: message_mismatch | result: fail | client_id: %v | sent_message: %v | received_message: %v",
				c.config.ID,
				msgToSend,
				receivedMsg,
			)
			return
		}

		// Wait a time between sending one message and the next one
		// time.Sleep(c.config.LoopPeriod)
	}

	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

package common

import (
	"bufio"
	"encoding/csv"
	"io"
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
	BetsPerChunk  int
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
// This function avoids the short-write problem
func (c *Client) sendMessage(msg string) error {
	writer := bufio.NewWriter(c.conn)
	writer.WriteString(msg)
	err := writer.Flush()
	if err != nil {
		log.Fatalf("action: flush_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	return nil
}

// Receives a message from the server.
// This function avoids the short-read problem
func (c *Client) receiveMessage() (string, error) {
	msg, err := bufio.NewReader(c.conn).ReadString(DELIMITER[0])
	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	return msg, err
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
func (c *Client) StartClientLoop() {
	signalReceiver := c.initializeSignalReceiver()

	agencyFile, err := os.Open("agency-" + c.config.ID + ".csv")
	if err != nil {
		log.Fatalf("action: open_agency_file | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	defer agencyFile.Close()
	agencyFileReader := csv.NewReader(agencyFile)

	// autoincremental msgToSendID to identify every message sent
	msgToSendID := 1

	for {
		// Create the connection the server
		c.createClientSocket() // TO BE MOVED UP ON EXERCISE 8
		defer c.conn.Close()   // TO BE MOVED UP ON EXERCISE 8

		select {
		case <-signalReceiver:
			c.handleShutdown(signalReceiver)
			return
		default:
		}

		currentChunkSize := 0
		chunkStringsToConcat := []string{}

		amountOfBetsRead := 0
		reachedEOF := false

		for (amountOfBetsRead < c.config.BetsPerChunk) && (currentChunkSize < MAX_CHUNK_SIZE_PER_MSG) && !reachedEOF {
			recordAsBet, err := ReadBetFromCsvRecord(agencyFileReader, c.config.ID)
			if err != io.EOF && err != nil {
				log.Fatalf("action: read_agency_file | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err,
				)
			}
			if err == io.EOF {
				reachedEOF = true
				continue
			}
			amountOfBetsRead++

			chunkStringsToConcat = append(chunkStringsToConcat, recordAsBet.ToString())
			currentChunkSize += len(recordAsBet.ToString()) + 1 // Due to incoming separation commas
		}
		if reachedEOF && amountOfBetsRead == 0 {
			return
		} else if reachedEOF && amountOfBetsRead > 0 {
			c.config.BetsPerChunk = amountOfBetsRead
		}

		if !(currentChunkSize < MAX_CHUNK_SIZE_PER_MSG) {

			betsPerChunkCorrectlyAdjusted := false
			if amountOfBetsRead%DEFAULT_BETS_PER_CHUNK != 0 {
				amountOfBetsInFirstChunk := amountOfBetsRead % DEFAULT_BETS_PER_CHUNK
				c.config.BetsPerChunk = amountOfBetsInFirstChunk
			} else {
				c.config.BetsPerChunk = DEFAULT_BETS_PER_CHUNK
			}

			for len(chunkStringsToConcat) > 0 {
				msgToSend := ""
				for i := 0; i < c.config.BetsPerChunk; i++ {
					if i == c.config.BetsPerChunk-1 {
						msgToSend += chunkStringsToConcat[i] + DELIMITER
					} else {
						msgToSend += chunkStringsToConcat[i] + ","
					}
				}
				err := c.sendMessage(msgToSend)
				if err != nil {
					return
				}

				receivedMsg, err := c.receiveMessage()
				if err != nil {
					return
				}
				if receivedMsg == CHUNK_ACK_MSG_FORMAT_FROM_SV {
					log.Infof("action: chunk_ack_received | result: success | client_id: %v | chunk_id: %v",
						c.config.ID,
						msgToSendID,
					)
				} else {
					log.Errorf("action: message_mismatch | result: fail | client_id: %v | chunk_id: %v | received_message: %v",
						c.config.ID,
						msgToSendID,
						receivedMsg,
					)
					return
				}
				c.conn.Close() // TO BE REMOVED ON EXERCISE 8

				chunkStringsToConcat = chunkStringsToConcat[c.config.BetsPerChunk:]
				if !betsPerChunkCorrectlyAdjusted {
					betsPerChunkCorrectlyAdjusted = true
					c.config.BetsPerChunk = DEFAULT_BETS_PER_CHUNK
				}

				c.createClientSocket() // TO BE REMOVED ON EXERCISE 8
				msgToSendID++
			}
			c.conn.Close() // TO BE REMOVED ON EXERCISE 8

		} else {

			msgToSend := ""
			for i := 0; i < c.config.BetsPerChunk; i++ {
				if i == c.config.BetsPerChunk-1 {
					msgToSend += chunkStringsToConcat[i] + DELIMITER
				} else {
					msgToSend += chunkStringsToConcat[i] + ","
				}
			}
			err := c.sendMessage(msgToSend)
			if err != nil {
				return
			}

			receivedMsg, err := c.receiveMessage()
			if err != nil {
				return
			}
			if receivedMsg == CHUNK_ACK_MSG_FORMAT_FROM_SV {
				log.Infof("action: chunk_ack_received | result: success | client_id: %v | chunk_id: %v",
					c.config.ID,
					msgToSendID,
				)
			} else {
				log.Errorf("action: message_mismatch | result: fail | client_id: %v | chunk_id: %v | received_message: %v",
					c.config.ID,
					msgToSendID,
					receivedMsg,
				)
				return
			}
			c.conn.Close() // TO BE REMOVED ON EXERCISE 8

			msgToSendID++
		}

	}

}

// Contains the implementation of a LSP client.

package lsp

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cmu440/lspnet"
	"sync"
	"time"
)

type client struct {
	// TODO: implement this!
	msg           []bool
	conn		  *lspnet.UDPConn
	seqNum        int
	connId        int
	left          int
	right         int
	start         int
	lastHB		  uint64
	isClosed 	  bool
	mu sync.Mutex
}

// NewClient creates, initiates, and returns a new client. This function
// should return after a connection with the server has been established
// (i.e., the client has received an Ack message from the server in response
// to its connection request), and should return a non-nil error if a
// connection could not be made (i.e., if after K epochs, the client still
// hasn't received an Ack message from the server in response to its K
// connection requests).
//
// initialSeqNum is an int representing the Initial Sequence Number (ISN) this
// client must use. You may assume that sequence numbers do not wrap around.
//
// hostport is a colon-separated string identifying the server's host address
// and port number (i.e., "localhost:9999").
func NewClient(hostport string, initialSeqNum int, params *Params) (Client, error) {
	for {
		go func() {
			conn, ok := lspnet.DialUDP(network, laddr, raddr)

			if conn != nil {
				for {
					res := []byte
					if _, _, err := conn.ReadFromUDP(res); err != nil {
						return &client{msg: []Message{}, mapMsgtoArray: make(map[int]int), seqNum: 1, connId: conn, left: 0, right: 0, start: 0}, nil

					}

				}
			}
		}()
	}
	return nil, errors.New("not yet implemented")
}

func (c *client) ConnID() int {
	return -1
}

// 2. Read and checksum
// 3. Accept the msg and revise the array unAck -> Ack
// 4. Move the window if possible
func (c *client) Read() ([]byte, error) {
	if c.isClosed {
		return nil, errors.New("connection closed")
	}
	var res []byte
	if _,_,err := c.conn.ReadFromUDP(res); err != nil {
		return nil, errors.New("connection issue")
	}
	var data Message
	if err := json.Unmarshal(res, &data); err != nil {
		return nil, errors.New("wrong unmarshal")
	}
	// Always be Ack
	seq := data.SeqNum
	if seq == 0 {
		return data.Payload, nil
	}
	if !c.msg[seq] {
		c.msg[seq] = true
		for c.left == seq {
			c.left++
		}
	}
	return data.Payload, nil
}

func (c *client) write(payload []byte, currentBackoff int) {
	if c.isClosed {
		return
	}
	time.Sleep(time.Duration(currentBackoff) * DefaultMaxBackOffInterval)
	c.mu.Lock()
	curD := c.seqNum
	if c.msg[curD] {
		return
	}
	c.msg = append(c.msg, false)
	if c.connId == -1 {
		msg := NewConnect(0)
		var data []byte
		data, err := json.Marshal(msg)
		if err != nil {
			fmt.Println(err.Error())
		}

		if _, err := c.conn.Write(data); err != nil {
			fmt.Println(err.Error())
		}
		c.seqNum++
	}
	c.mu.Unlock()
	if currentBackoff == 0 {
		go c.write(payload, 1)
	} else {
		go c.write(payload, currentBackoff * 2)
	}
}

// 1. New thread
// 2. Write, if fails, retry it after backoff exponential
// 3. append it to back of the array with unAck
func (c *client) Write(payload []byte) error {
	if c.isClosed {
		return errors.New("connection closed")
	}
	go c.write(payload, DefaultMaxBackOffInterval)
	return errors.New("not yet implemented")
}

func (c *client) Close() error {
	return errors.New("not yet implemented")
}

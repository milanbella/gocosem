package gocosem

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

var (
	ErrFailSNRM = errors.New("SNRM fail")
	ErrFailAARQ = errors.New("AARQ fail")
	ErrFailDISC = errors.New("DISC fail")
	ErrTimeout  = errors.New("timeout")
)

const (
	cosemPort      = 4059
	defaultTimeout = time.Second * 10
)

var Debug bool

func logf(format string, arg ...interface{}) {
	if Debug {
		msg := fmt.Sprintf(format, arg...)
		log.Printf("[ gocosem ] %s", msg)
	}
}

type Cosem struct {
	ip         string
	timeoutDur time.Duration
	conn       net.Conn
	readC      chan []byte
	errC       chan error
}

func NewCosem(ip string) *Cosem {
	return &Cosem{
		ip:         ip,
		timeoutDur: defaultTimeout,
		readC:      make(chan []byte, 1),
		errC:       make(chan error, 1),
	}
}

func (c *Cosem) Connect() (err error) {
	addr := fmt.Sprintf("%s:%d", c.ip, cosemPort)
	logf("connecting to %s", addr)

	if c.conn, err = net.Dial("tcp", addr); err != nil {
		return err
	}
	go c.readResponses()

	if err = c.sendSNRM(); err != nil {
		return err
	}
	if err = c.sendAARQ(); err != nil {
		return err
	}
	logf("connected %s", c.ip)
	return nil
}

func (c *Cosem) Disconnect() error {
	logf("disconnecting %s", c.ip)

	err := c.sendDISC()
	c.conn.Close()
	<-c.errC

	logf("disconnected %s", c.ip)
	return err
}

func (c *Cosem) RetrieveSerialNumber() (string, error) {
	logf("sending serial number request")

	rep, err := c.sendRequest(reqSerialNum)
	if err != nil {
		return "", err
	} else if len(rep) < len(repSerialNum)+11 || !isSameData(repSerialNum, rep[:len(repSerialNum)]) {
		return "", fmt.Errorf("request failed")
	}
	return string(rep[len(repSerialNum) : len(repSerialNum)+8]), nil
}

func (c *Cosem) sendSNRM() error {
	logf("sending SNRM")

	if rep, err := c.sendRequest(reqSNRM); err != nil {
		return err
	} else if !isSameData(rep, repSNRM) {
		logf("invalid SNRM reply")
		return ErrFailSNRM
	}
	return nil
}

func (c *Cosem) sendAARQ() error {
	logf("sending AARQ")

	if rep, err := c.sendRequest(reqAARQ); err != nil {
		return err
	} else if !isSameData(rep, repAARQ) {
		logf("invalid AARQ reply")
		return ErrFailAARQ
	}
	return nil
}

func (c *Cosem) sendDISC() error {
	logf("sending DISC")

	if rep, err := c.sendRequest(reqDISC); err != nil {
		return err
	} else if !isSameData(rep, repDISC) {
		logf("invalid DISC reply")
		return ErrFailDISC
	}
	return nil
}

func (c *Cosem) sendRequest(req []byte) ([]byte, error) {
	logf("request data \"% 02X\"\n", req)

	if _, err := c.conn.Write(req); err != nil {
		return nil, err
	}

	select {
	case resp := <-c.readC:
		logf("response data \"% 02X\"\n", resp)
		return resp, nil

	case <-time.After(c.timeoutDur):
		logf("request timed out")
		return nil, ErrTimeout

	case err := <-c.errC:
		logf("connection error %s", err)
		return nil, err
	}
}

func (c *Cosem) readResponses() {
	buffC := make(chan []byte, 1)

	go func() {
		defer close(buffC)
		buf := make([]byte, 1024)
		for {
			if n, err := c.conn.Read(buf); err != nil {
				c.errC <- err
				return
			} else {
				d := make([]byte, n)
				copy(d, buf[:n])
				buffC <- d
			}
		}
	}()

	for d := range buffC {
		logf("received %d bytes", len(d))
		func() {
			for d[len(d)-1] != 0x7E {
				logf("waiting for rest of data")
				select {
				case ex := <-buffC:
					d = append(d, ex...)

				case <-time.After(time.Millisecond * 500):
					return
				}
			}
		}()
		c.readC <- d
	}
}

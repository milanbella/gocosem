package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

const (
	cosemPort = 4059
)

var (
	reqSNRM = []byte{
		0x7E, 0xA0, 0x07, 0x03, 0x03, 0x93, 0x8C, 0x11, 0x7E,
	}
	repSNRM = []byte{
		0x7E, 0xA0, 0x20, 0x03, 0x03, 0x73, 0xF0, 0x2E, 0x81,
		0x80, 0x14, 0x05, 0x02, 0x00, 0xEF, 0x06, 0x02, 0x00,
		0xEF, 0x07, 0x04, 0x00, 0x00, 0x00, 0x01, 0x08, 0x04,
		0x00, 0x00, 0x00, 0x01, 0x07, 0x4B, 0x7E,
	}
	reqAARQ = []byte{
		0x7E, 0xA0, 0x44, 0x03, 0x03, 0x10, 0x65, 0x94, 0xE6, 0xE6,
		0x00, 0x60, 0x36, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74,
		0x05, 0x08, 0x01, 0x01, 0x8A, 0x02, 0x07, 0x80, 0x8B, 0x07,
		0x60, 0x85, 0x74, 0x05, 0x08, 0x02, 0x01, 0xAC, 0x0A, 0x80,
		0x08, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0xBE,
		0x10, 0x04, 0x0E, 0x01, 0x00, 0x00, 0x00, 0x06, 0x5F, 0x1F,
		0x04, 0x00, 0xFF, 0xFF, 0xFF, 0x02, 0x00, 0xA5, 0xD9, 0x7E,
	}
	repAARQ = []byte{
		0x7E, 0xA0, 0x37, 0x03, 0x03, 0x30, 0xEF, 0xCA, 0xE6, 0xE7,
		0x00, 0x61, 0x29, 0xA1, 0x09, 0x06, 0x07, 0x60, 0x85, 0x74,
		0x05, 0x08, 0x01, 0x01, 0xA2, 0x03, 0x02, 0x01, 0x00, 0xA3,
		0x05, 0xA1, 0x03, 0x02, 0x01, 0x00, 0xBE, 0x10, 0x04, 0x0E,
		0x08, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0xFE, 0x1D,
		0x00, 0xEF, 0x00, 0x07, 0x79, 0x15, 0x7E,
	}
	reqUtility = []byte{
		0x7E, 0xA0, 0x19, 0x03, 0x03, 0x32, 0xEC, 0xC8, 0xE6, 0xE6,
		0x00, 0xC0, 0x01, 0xC1, 0x00, 0x01, 0x00, 0x00, 0x60, 0x01,
		0x00, 0xFF, 0x02, 0x00, 0x89, 0xA0, 0x7E,
	}
	repUtility = []byte{
		0x7E, 0xA0, 0x1A, 0x03, 0x03, 0x52, 0x27, 0x8E, 0xE6, 0xE7,
		0x00, 0xC4, 0x01, 0xC1, 0x00, 0x09, 0x08,
	}
	reqDISC = []byte{
		0x7E, 0xA0, 0x07, 0x03, 0x03, 0x53, 0x80, 0xD7, 0x7E,
	}
)

type Cosem struct {
	ip     string
	conn   net.Conn
	readC  chan []byte
	errC   chan error
	writeC chan []byte
}

func NewCosem(ip string) *Cosem {
	return &Cosem{
		ip:     ip,
		readC:  make(chan []byte, 1),
		errC:   make(chan error, 1),
		writeC: make(chan []byte, 1),
	}
}

func (c *Cosem) Connect() (err error) {
	log.Println("connecting")
	c.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", c.ip, cosemPort))
	if err != nil {
		return err
	}
	c.processCom()
	log.Println("connected")

	log.Println("sending SNRM")
	if rep, err := c.sendReq(reqSNRM); err != nil {
		return err
	} else if !sameData(rep, repSNRM) {
		return fmt.Errorf("SNRM failed")
	}

	log.Println("sending AARQ")
	if rep, err := c.sendReq(reqAARQ); err != nil {
		return err
	} else if !sameData(rep, repAARQ) {
		return fmt.Errorf("AARQ failed")
	}

	return nil
}

func (c *Cosem) sendReq(req []byte) (data []byte, err error) {
	if verbose {
		log.Printf("sending data: % 02X \n", req)
	}
	select {
	case c.writeC <- req:
	default:
		return data, fmt.Errorf("send failed")
	}
	for {
		select {
		case data = <-c.readC:
			if verbose {
				log.Printf("received data: % 02X \n", data)
			}
			return data, nil
		case <-time.After(time.Second * 10):
			return data, fmt.Errorf("timed out")
		case err := <-c.errC:
			return data, err
		}
	}
}

func (c *Cosem) processCom() {
	buffC := make(chan []byte, 1)
	go func() {
		if verbose {
			log.Println("buffer goroutine started")
			defer log.Println("buffer goroutine finished")
		}
		for d := range buffC {
			func() {
				for d[len(d)-1] != 0x7E {
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
	}()
	go func() {
		if verbose {
			log.Println("read goroutine started")
			defer log.Println("read goroutine finished")
		}
		defer close(buffC)
		defer close(c.writeC)
		buf := make([]byte, 1024)
		for {
			n, err := c.conn.Read(buf)
			if err != nil {
				c.errC <- err
				return
			}
			if verbose {
				log.Println("read", n, "bytes")
			}
			d := make([]byte, n)
			copy(d, buf[:n])
			buffC <- d
		}
	}()
	go func() {
		if verbose {
			log.Println("write goroutine started")
			defer log.Println("write goroutine finished")
		}
		for data := range c.writeC {
			if _, err := c.conn.Write(data); err != nil {
				c.errC <- err
				return
			}
		}
	}()
}

func (c *Cosem) Disconnect() {
	log.Println("disconnecting")
	if _, err := c.sendReq(reqDISC); err != nil {
		log.Println("disconnect failed", err)
	}
	c.conn.Close()
	<-c.errC
}

func (c *Cosem) GetUtility() (string, error) {
	log.Println("sending utility request")
	rep, err := c.sendReq(reqUtility)
	if err != nil {
		return "", err
	} else if len(rep) < len(repUtility)+11 || !sameData(repUtility, rep[:len(repUtility)]) {
		return "", fmt.Errorf("request failed")
	}
	return string(rep[len(repUtility) : len(repUtility)+8]), nil
}

func sameData(data, comp []byte) bool {
	if len(data) != len(comp) {
		return false
	}
	for i, b := range data {
		if b != comp[i] {
			return false
		}
	}
	return true
}

package main

import (
	"flag"
	"log"
)

var (
	ip      string
	verbose bool
)

func init() {
	flag.StringVar(&ip, "ip", "192.168.3.119", "target ip")
	flag.BoolVar(&verbose, "v", false, "verbose")
}

func main() {
	flag.Parse()

	cosem := NewCosem(ip)
	if err := cosem.Connect(); err != nil {
		log.Fatal(err)
	}
	if utility, err := cosem.GetUtility(); err != nil {
		log.Println("utility failed", err)
	} else {
		log.Println("utility:", utility)
	}
	cosem.Disconnect()
}

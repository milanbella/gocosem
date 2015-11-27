package main

import (
	"flag"
	"log"
	cosem "x4/gocosem"
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
	cosem.Debug = verbose

	log.Println("initialising connection")
	cos := cosem.NewCosem(ip)
	if err := cos.Connect(); err != nil {
		log.Println("connection failed:", err)
		return
	}
	defer cos.Disconnect()

	log.Println("retrieving serial number")
	if sn, err := cos.RetrieveSerialNumber(); err != nil {
		log.Println("retrieving failed:", err)
	} else {
		log.Printf("successfully retrieved \"%s\"", sn)
	}
}

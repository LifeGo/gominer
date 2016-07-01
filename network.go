package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

//HeaderReporter defines the required method a SIA client or pool client should implement for miners to be able to report solved headers
type HeaderReporter interface {
	//SubmitHeader reports a solved header
	SubmitHeader(header []byte) (err error)
}

// SiadClient is used to connect to siad
type SiadClient struct {
	siadurl string
}

// NewSiadClient creates a new SiadClient given a 'host:port' connectionstring
func NewSiadClient(connectionstring string, querystring string) *SiadClient {
	s := SiadClient{}
	s.siadurl = "http://" + connectionstring + "/miner/header?" + querystring
	return &s
}

//GetHeaderForWork fetches new work from the SIA daemon
func (sc *SiadClient) GetHeaderForWork() (target, header []byte, err error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", sc.siadurl, nil)
	if err != nil {
		return
	}

	req.Header.Add("User-Agent", "Sia-Agent")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
	case 400:
		err = fmt.Errorf("Invalid siad response, status code %d, is your wallet initialized and unlocked?", resp.StatusCode)
		return
	default:
		err = fmt.Errorf("Invalid siad response, status code %d", resp.StatusCode)
		return
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if len(buf) < 112 {
		err = fmt.Errorf("Invalid siad response, only received %d bytes, is your wallet initialized and unlocked?", len(buf))
		return
	}

	target = buf[:32]
	header = buf[32:112]

	return
}

//SubmitHeader reports a solved header to the SIA daemon
func (sc *SiadClient) SubmitHeader(header []byte) (err error) {
	req, err := http.NewRequest("POST", sc.siadurl, bytes.NewReader(header))
	if err != nil {
		return
	}

	req.Header.Add("User-Agent", "Sia-Agent")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	switch resp.StatusCode {
	case 204:
	default:
		err = fmt.Errorf("Invalid siad response, status code %d", resp.StatusCode)
		return
	}
	return
}

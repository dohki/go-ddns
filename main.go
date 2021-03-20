package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiURL = "https://domains.google.com/nic/update"

type config struct {
	Username, Password, Hostname string
}

var conf = new(config)
var reqData = url.Values{}

func init() {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data, conf); err != nil {
		panic(err)
	}
	reqData.Set("hostname", conf.Hostname)
}

func handleErr(err error) {
	// TODO: Alert via slack
	if err != nil {
		log.Println(err)
	}
}

func makeReq() *http.Request {
	reader := strings.NewReader(reqData.Encode())
	req, err := http.NewRequest("POST", apiURL, reader)
	if err != nil {
		handleErr(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(conf.Username, conf.Password)
	return req
}

func updateIP(client *http.Client) (string, error) {
	req := makeReq()
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func handleResult(result string) {
	// FIXME: Use regexp
	if strings.HasPrefix(result, "good") || strings.HasPrefix(result, "nochg") {
		log.Println(result)
	} else {
		handleErr(errors.New(result))
	}
}

func main() {
	client := http.DefaultClient
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		result, err := updateIP(client)
		handleErr(err)
		handleResult(result)
	}
}

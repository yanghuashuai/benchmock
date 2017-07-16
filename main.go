package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"time"
)

type Latency struct {
	Average int `json:"average"`
	Delta   int `json:"delta"`
}

func (l Latency) CalcDuration() time.Duration {
	latency := l.Average + rand.Intn(l.Delta)
	return time.Duration(latency) * time.Millisecond
}

type InterfaceDesc struct {
	URI        string            `json:"uri"`
	StatusCode int               `json:"statusCode"`
	Header     map[string]string `json:"header"`
	Body       interface{}       `json:"body"`
	Latency    Latency           `json:"latency"`
}

const pattern = `
==
uri=%s
statusCode=%d
body=%s
latency=%s
`

var file string
var host string

func init() {
	flag.StringVar(&file, "f", "", "configuation file for mock interfaces")
	flag.StringVar(&host, "h", "127.0.0.1:9527", "host address of the mock server")
}

func main() {
	flag.Parse()
	if len(file) == 0 {
		flag.PrintDefaults()
		exit(nil)
	}

	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		exit(err)
	}
	data, err := ioutil.ReadFile(path.Join(wd, file))
	if err != nil {
		exit(err)
	}

	var interfaces []InterfaceDesc
	err = json.Unmarshal(data, &interfaces)
	if err != nil {
		exit(err)
	}
	for _, desc := range interfaces {
		var body []byte
		if desc.Body != nil {
			body, err = json.Marshal(desc.Body)
			if err != nil {
				exit(err)
			}
		}
		http.HandleFunc(desc.URI, func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(desc.Latency.CalcDuration())
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			for k, v := range desc.Header {
				w.Header().Set(k, v)
			}
			w.WriteHeader(desc.StatusCode)
			if len(body) > 0 {
				w.Write(body)
			}
		})
		fmt.Printf(pattern,
			desc.URI,
			desc.StatusCode,
			string(body),
			fmt.Sprintf("[latency:%d,delta:%d]", desc.Latency.Average, desc.Latency.Delta),
		)
	}

	err = http.ListenAndServe(host, nil)
	if err != nil {
		err = fmt.Errorf("ListenAndServe on(%s) error: %v", host, err)
		exit(err)
	}
}

func exit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
	os.Exit(1)
}

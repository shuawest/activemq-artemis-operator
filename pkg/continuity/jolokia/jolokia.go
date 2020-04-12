package jolokia

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("jolokia")

type IJolokia interface {
	Read(_path string) (string, error)
	Exec(_path string, _postJsonString string) (string, error)
}

type Jolokia struct {
	ip         string
	port       string
	jolokiaURL string
	user       string
	pass       string
	userAgent  string
	origin     string
}

type JolokiaError struct {
	Operation string `json:"operation"`
	Detail    string `json:"detail"`
	Status    string `json:"status"`
	Cause     error  `json:"cause"`
}

func (je JolokiaError) Error() string {
	msg, _ := json.Marshal(je)
	return string(msg)
}

func NewJolokia(_ip string, _port string, _path string, _user string, _pass string) *Jolokia {
	j := Jolokia{
		ip:         _ip,
		port:       _port,
		jolokiaURL: "http://" + _user + ":" + _pass + "@" + _ip + ":" + _port + _path,
		user:       _user,
		pass:       _pass,
		userAgent:  "activemq-artemis-management",
		origin:     "http://localhost/",
	}
	return &j
}

func (j *Jolokia) Read(_path string) (string, error) {
	url := j.jolokiaURL + "/read/" + _path

	jolokiaClient := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 seconds
	}

	var bodyString string

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return bodyString, JolokiaError{
			Operation: "failed creating read request",
			Detail:    url,
			Cause:     err,
		}
	}
	req.Header.Set("User-Agent", j.userAgent)
	req.Header.Set("Origin", j.origin)

	res, err := jolokiaClient.Do(req)
	if err != nil {
		if res != nil {
			return bodyString, JolokiaError{
				Operation: "Fault received while calling jolokia read endpoint",
				Detail:    url,
				Cause:     err,
				Status:    res.Status,
			}
		} else {
			return bodyString, JolokiaError{
				Operation: "Failed calling jolokia read endpoint",
				Detail:    url,
				Cause:     err,
			}
		}
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return bodyString, JolokiaError{
			Operation: "Failed reading read result body",
			Cause:     err,
			Status:    res.Status,
		}
	}

	bodyString = string(body)

	return bodyString, nil
}

func (j *Jolokia) Exec(_path string, _postJsonString string) (string, error) {
	url := j.jolokiaURL + "/exec/" + _path

	jolokiaClient := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 seconds
	}

	var bodyString string

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(_postJsonString)))
	if err != nil {
		return bodyString, JolokiaError{
			Operation: "failed creating exec request",
			Detail:    url,
			Cause:     err,
		}
	}
	req.Header.Set("User-Agent", j.userAgent)
	req.Header.Set("Origin", j.origin)
	req.Header.Set("Content-Type", "application/json")

	res, err := jolokiaClient.Do(req)
	if err != nil {
		if res != nil {
			return bodyString, JolokiaError{
				Operation: "Fault received while calling jolokia exec endpoint",
				Detail:    url,
				Cause:     err,
				Status:    res.Status,
			}
		} else {
			return bodyString, JolokiaError{
				Operation: "Failed calling jolokia exec endpoint",
				Detail:    url,
				Cause:     err,
			}
		}
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return bodyString, JolokiaError{
			Operation: "Failed reading exec result body",
			Cause:     err,
			Status:    res.Status,
		}
	}

	bodyString = string(body)

	return bodyString, err
}

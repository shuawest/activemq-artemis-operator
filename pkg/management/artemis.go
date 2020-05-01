package management

import (
	"encoding/json"
	"strconv"

	"strings"

	"github.com/rh-messaging/activemq-artemis-operator/pkg/management/jolokia"
)

type IArtemis interface {
	NewArtemis(_name string, _jolokia jolokia.IJolokia) *Artemis
	CreateQueue(addressName string, queueName string) (string, error)
	DeleteQueue(queueName string) (string, error)
	ListBindingsForAddress(addressName string) (string, error)
	DeleteAddress(addressName string) (string, error)
}

type Artemis struct {
	name    string
	jolokia jolokia.IJolokia
}

type ArtemisError struct {
	Operation string `json:"operation"`
	Detail    string `json:"detail"`
	Status    string `json:"status"`
	Cause     error  `json:"cause"`
}

func (ae ArtemisError) Error() string {
	msg, _ := json.Marshal(ae)
	return string(msg)
}

type ExecResult struct {
	Timestamp int    `json:"timestamp"`
	Status    int    `json:"status"`
	ErrorType string `json:"error_type"`
	Error     string `json:"error"`
}

func NewArtemis(_name string, _jolokia jolokia.IJolokia) *Artemis {
	artemis := Artemis{
		name:    _name,
		jolokia: _jolokia,
	}
	return &artemis
}

func (artemis *Artemis) CreateQueue(addressName string, queueName string, routingType string) error {

	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\""
	routingType = strings.ToUpper(routingType)
	parameters := `"` + addressName + `","` + queueName + `",` + `"` + routingType + `"`
	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"createQueue(java.lang.String,java.lang.String,java.lang.String)","arguments":[` + parameters + `]` + ` }`
	data, err := artemis.jolokia.Exec(url, jsonStr)

	return evaluateExecResult(data, err)
}

func (artemis *Artemis) DeleteQueue(queueName string) error {

	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\""
	parameters := `"` + queueName + `"`
	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"destroyQueue(java.lang.String)","arguments":[` + parameters + `]` + ` }`
	data, err := artemis.jolokia.Exec(url, jsonStr)

	return evaluateExecResult(data, err)
}

func (artemis *Artemis) ListBindingsForAddress(addressName string) (string, error) {

	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\""
	parameters := `"` + addressName + `"`
	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"listBindingsForAddress(java.lang.String)","arguments":[` + parameters + `]` + ` }`
	data, err := artemis.jolokia.Exec(url, jsonStr)

	execErr := evaluateExecResult(data, err)
	if execErr != nil {
		return "", execErr
	}

	return data, err
}

func (artemis *Artemis) DeleteAddress(addressName string) error {

	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\""
	parameters := `"` + addressName + `"`
	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"deleteAddress(java.lang.String)","arguments":[` + parameters + `]` + ` }`
	data, err := artemis.jolokia.Exec(url, jsonStr)

	return evaluateExecResult(data, err)
}

func evaluateExecResult(body string, err error) error {
	if err != nil {
		return ArtemisError{
			Operation: "calling jolokia exec service",
			Detail:    err.Error(),
			Cause:     err,
		}
	}

	// Unmarshal result
	result := &ExecResult{}
	err = json.Unmarshal([]byte(body), result)
	if err != nil {
		return ArtemisError{
			Operation: "failed to unmarshal result",
			Detail:    "body: " + body,
			Cause:     err,
		}
	}

	if result.Status != 200 {
		return ArtemisError{
			Operation: "jolokia returned fault",
			Detail:    result.Error,
			Status:    strconv.Itoa(result.Status),
			Cause:     err,
		}
	}

	return err
}

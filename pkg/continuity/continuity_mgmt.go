package continuity

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/rh-messaging/activemq-artemis-operator/pkg/management/jolokia"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("continuity_mgmt")

type ArtemisContinuity struct {
	name    string
	jolokia jolokia.IJolokia
}

type ContinuityError struct {
	Operation string `json:"operation"`
	Detail    string `json:"detail"`
	Status    string `json:"status"`
	Cause     error  `json:"cause"`
}

func (cme ContinuityError) Error() string {
	msg, _ := json.Marshal(cme)
	return string(msg)
}

type BootedStatus struct {
	Booted bool `json:"Booted,omitempty"`
}

type BootedResult struct {
	Value     BootedStatus `json:"value"`
	Timestamp int          `json:"timestamp"`
	Status    int          `json:"status"`
	ErrorType string       `json:"error_type"`
	Error     string       `json:"error"`
}

type ExecResult struct {
	Timestamp int    `json:"timestamp"`
	Status    int    `json:"status"`
	ErrorType string `json:"error_type"`
	Error     string `json:"error"`
}

func NewArtemisContinuity(_name string, _jolokia jolokia.IJolokia) *ArtemisContinuity {
	ac := ArtemisContinuity{
		name:    _name,
		jolokia: _jolokia,
	}
	return &ac
}

func (artemis *ArtemisContinuity) IsBooted() (bool, error) {
	url := fmt.Sprintf(`org.apache.activemq.artemis:broker="%s",component=continuity,name="continuity.bootstrap"`, artemis.name)

	// Call the service
	body, err := artemis.jolokia.Read(url)
	if err != nil {
		return false, ContinuityError{
			Operation: "calling jolokia read service",
			Detail:    err.Error(),
			Cause:     err,
		}
	}

	// Unmarshal result
	result := &BootedResult{}
	err = json.Unmarshal([]byte(body), result)
	if err != nil {
		return false, ContinuityError{
			Operation: "failed to unmarshal result",
			Detail:    "body: " + body,
			Cause:     err,
		}
	}

	if result.Status != 200 {
		return false, ContinuityError{
			Operation: "jolokia returned fault",
			Detail:    body,
			Status:    strconv.Itoa(result.Status),
			Cause:     err,
		}
	}

	return result.Value.Booted, nil
}

func (artemis *ArtemisContinuity) Configure(siteId string, activeOnStart bool, servingAcceptors string, localConnectorRef string, remoteConnectorRefs string, reorgManagement bool) error {
	url := fmt.Sprintf("org.apache.activemq.artemis:broker=\\\"%s\\\",component=continuity,name=\\\"continuity.bootstrap\\\"", artemis.name)
	operation := "configure(java.lang.String,java.lang.Boolean,java.lang.String,java.lang.String,java.lang.String,java.lang.Boolean)"
	parameters := fmt.Sprintf(`"%s",%t,"%s","%s","%s",%t`, siteId, activeOnStart, servingAcceptors, localConnectorRef, remoteConnectorRefs, reorgManagement)
	jsonStr := fmt.Sprintf(`{ "type":"EXEC","mbean":"%s","operation":"%s","arguments":[%s] }`, url, operation, parameters)

	body, err := artemis.jolokia.Exec(url, jsonStr)
	return evaluateExecResult(body, err)
}

func (artemis *ArtemisContinuity) SetSecrets(localContinuityUser string, localContinuityPass string, remoteContinuityUser string, remoteContinuityPass string) error {
	url := fmt.Sprintf("org.apache.activemq.artemis:broker=\\\"%s\\\",component=continuity,name=\\\"continuity.bootstrap\\\"", artemis.name)
	operation := "setSecrets(java.lang.String,java.lang.String,java.lang.String,java.lang.String)"
	parameters := fmt.Sprintf(`"%s","%s","%s","%s"`, localContinuityUser, localContinuityPass, remoteContinuityUser, remoteContinuityPass)
	jsonStr := fmt.Sprintf(`{ "type":"EXEC","mbean":"%s","operation":"%s","arguments":[%s] }`, url, operation, parameters)

	body, err := artemis.jolokia.Exec(url, jsonStr)
	return evaluateExecResult(body, err)
}

func (artemis *ArtemisContinuity) Tune(activationTimeout int, inflowStagingDelay int, bridgeInterval int, bridgeIntervalMultiplier float32, pollDuration int) error {
	url := fmt.Sprintf("org.apache.activemq.artemis:broker=\\\"%s\\\",component=continuity,name=\\\"continuity.bootstrap\\\"", artemis.name)
	operation := "tune(java.lang.Long,java.lang.Long,java.lang.Long,java.lang.Double,java.lang.Long)"
	parameters := fmt.Sprintf(`%d,%d,%d,%f,%d`, activationTimeout, inflowStagingDelay, bridgeInterval, bridgeIntervalMultiplier, pollDuration)
	jsonStr := fmt.Sprintf(`{ "type":"EXEC","mbean":"%s","operation":"%s","arguments":[%s] }`, url, operation, parameters)

	body, err := artemis.jolokia.Exec(url, jsonStr)
	return evaluateExecResult(body, err)
}

func (artemis *ArtemisContinuity) Boot() error {
	url := fmt.Sprintf("org.apache.activemq.artemis:broker=\\\"%s\\\",component=continuity,name=\\\"continuity.bootstrap\\\"", artemis.name)
	operation := "boot"
	jsonStr := fmt.Sprintf(`{ "type":"EXEC","mbean":"%s","operation":"%s" }`, url, operation)

	body, err := artemis.jolokia.Exec(url, jsonStr)
	return evaluateExecResult(body, err)
}

func evaluateExecResult(body string, err error) error {
	if err != nil {
		return ContinuityError{
			Operation: "calling jolokia exec service",
			Detail:    err.Error(),
			Cause:     err,
		}
	}

	// Unmarshal result
	result := &ExecResult{}
	err = json.Unmarshal([]byte(body), result)
	if err != nil {
		return ContinuityError{
			Operation: "failed to unmarshal result",
			Detail:    "body: " + body,
			Cause:     err,
		}
	}

	if result.Status != 200 {
		return ContinuityError{
			Operation: "jolokia returned fault",
			Detail:    result.Error,
			Status:    strconv.Itoa(result.Status),
			Cause:     err,
		}
	}

	return err
}

// func (artemis *ArtemisContinuity) RebootContinuity() (*jolokia.ExecData, error) {
// 	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\",component=continuity,name=\\\"continuity.bootstrap\\\""
// 	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"reboot" }`

// 	data, err := artemis.jolokia.Exec(url, jsonStr)

// 	return data, err
// }

// func (artemis *ArtemisContinuity) DestroyContinuity() (*jolokia.ExecData, error) {
// 	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\",component=continuity,name=\\\"continuity.bootstrap\\\""
// 	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"destory" }`

// 	data, err := artemis.jolokia.Exec(url, jsonStr)

// 	return data, err
// }

// func (artemis *ArtemisContinuity) ActivateSite(siteId string, activationTimeout int64) (*jolokia.ExecData, error) {
// 	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\",component=continuity,name=\\\"continuity.service\\\""
// 	parameters := strconv.FormatInt(activationTimeout, 10)
// 	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"activateSite(java.lang.Long)","arguments":[` + parameters + `]` + ` }`
// 	data, err := artemis.jolokia.Exec(url, jsonStr)

// 	return data, err
// }

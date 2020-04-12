package continuity

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/continuity/jolokia"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func init() {
	logf.SetLogger(zap.LoggerTo(os.Stdout))
}

type ErrorMock struct {
	msg string
}

func (em ErrorMock) Error() string {
	return em.msg
}

type JolokiaMock struct {
	body string
	err  error
}

func (jm JolokiaMock) Read(_path string) (string, error) {
	return jm.body, jm.err
}

func (jm JolokiaMock) Exec(_path string, _postJsonString string) (string, error) {
	return jm.body, jm.err
}

/** Integration Tests
 *  Need to run an AMQ Broker with bootstrap installed (https://github.com/shuawest/amq-server-continuity/blob/master/bootstrap.sh)
 *  then set runIntegrationTests to true
**/
var runIntegrationTests bool = false

func TestInt_IsBooted(t *testing.T) {
	if !runIntegrationTests {
		t.Skip()
	}

	jolokia := jolokia.NewJolokia("localhost", "8181", "/console/jolokia", "continuity1", "pass1")
	artemis := NewArtemisContinuity("amqbroker-site1", jolokia)

	_, err := artemis.IsBooted()
	assertNoError(t, err, "")
}

func TestInt_Configure(t *testing.T) {
	if !runIntegrationTests {
		t.Skip()
	}

	jolokia := jolokia.NewJolokia("localhost", "8181", "/console/jolokia", "continuity1", "pass1")
	artemis := NewArtemisContinuity("amqbroker-site1", jolokia)

	err := artemis.Configure("site1", true, "amqp", "continuity-local", "ctyconn", true)
	assertNoError(t, err, "")
}

func TestInt_SetSecrets(t *testing.T) {
	if !runIntegrationTests {
		t.Skip()
	}

	jolokia := jolokia.NewJolokia("localhost", "8181", "/console/jolokia", "continuity1", "pass1")
	artemis := NewArtemisContinuity("amqbroker-site1", jolokia)

	err := artemis.SetSecrets("continuity1", "pass1", "continuity2", "pass2")
	assertNoError(t, err, "")
}

func TestInt_Tune(t *testing.T) {
	if !runIntegrationTests {
		t.Skip()
	}

	jolokia := jolokia.NewJolokia("localhost", "8181", "/console/jolokia", "continuity1", "pass1")
	artemis := NewArtemisContinuity("amqbroker-site1", jolokia)

	err := artemis.Tune(1000000, 1000000, 1000, 0.5, 1000)
	assertNoError(t, err, "")
}

func TestInt_Boot(t *testing.T) {
	if !runIntegrationTests {
		t.Skip()
	}

	jolokia := jolokia.NewJolokia("localhost", "8181", "/console/jolokia", "continuity1", "pass1")
	artemis := NewArtemisContinuity("amqbroker-site1", jolokia)

	err := artemis.Boot()
	assertNoError(t, err, "")
}

/** Mock Sucess Tests **/

func TestMock_IsBooted(t *testing.T) {
	jmock := JolokiaMock{
		body: `{"value":{"Booted":true},"timestamp":1586367954,"status":200}`,
		err:  nil,
	}
	artemis := NewArtemisContinuity("amqbroker-site1", jmock)

	isBooted, err := artemis.IsBooted()
	assertEqual(t, isBooted, true, "")
	assertNoError(t, err, "")
}

func TestMock_NotBooted(t *testing.T) {
	jmock := JolokiaMock{
		body: `{"value":{"Booted":false},"timestamp":1586367954,"status":200}`,
		err:  nil,
	}
	artemis := NewArtemisContinuity("amqbroker-site1", jmock)

	isBooted, err := artemis.IsBooted()
	assertEqual(t, isBooted, false, "")
	assertNoError(t, err, "")
}

/** Mock Fail Tests **/

func TestMock_FailCannotConnect(t *testing.T) {
	jmock := JolokiaMock{
		err: ErrorMock{
			msg: "Get http://admin:***@localhost:8161/console/jolokia/read/org.apache.activemq.artemis:broker=%22amqbroker-site1%22,component=continuity,name=%22continuity.bootstrap%22: dial tcp [::1]:8161: connect: connection refused",
		},
	}
	artemis := NewArtemisContinuity("amqbroker-site1", jmock)

	isBooted, err := artemis.IsBooted()
	assertEqual(t, isBooted, false, "isBooted false on failure")
	assertError(t, err, "connect: connection refused")
}

func TestMock_FailNonExistentBroker(t *testing.T) {
	jmock := JolokiaMock{
		body: "{\"request\":{\"mbean\":\"org.apache.activemq.artemis:broker=\\\"NonExistentBroker\\\",component=continuity,name=\\\"continuity.bootstrap\\\"\",\"type\":\"read\"},\"error_type\":\"javax.management.InstanceNotFoundException\",\"error\":\"javax.management.InstanceNotFoundException : org.apache.activemq.artemis:broker=\\\"NonExistentBroker\\\",component=continuity,name=\\\"continuity.bootstrap\\\"\",\"status\":404}",
		err:  nil,
	}
	artemis := NewArtemisContinuity("NonExistentBroker", jmock)

	isBooted, err := artemis.IsBooted()
	assertEqual(t, isBooted, false, "isBooted false on failure")
	assertError(t, err, "javax.management.InstanceNotFoundException")
}

/** Test Helper Functions **/

func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

func assertNoError(t *testing.T, err error, message string) {
	if err == nil {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("Expected no error, but got: %s", err)
	}
	log.Error(err, message)
	t.Fatal(message)
}

func assertError(t *testing.T, err error, expectedMessage string) {
	var msg string
	if err != nil && strings.Contains(err.Error(), expectedMessage) {
		return
	}
	if err == nil {
		msg = fmt.Sprintf("Expected error containing '%s', but error was nil", expectedMessage)
	} else {
		msg = fmt.Sprintf("Expected error containing '%s', but got: %s", expectedMessage, err)
	}
	log.Error(err, msg)
	t.Fatal(msg)
}

// func evalReadResults(callName string, data *jolokia.ReadData, err error) {
// 	if data != nil {
// 		reqLogger := log.WithValues("data", data, "data.Value", data.Value)
// 		reqLogger.Info(callName + " result")
// 	}
// 	if err != nil {
// 		log.Error(err, callName+" err")
// 	}
// 	//fmt.Println("\n")
// }

// func evalExecResults(callName string, data *jolokia.ExecData, err error) {
// 	if data != nil {
// 		reqLogger := log.WithValues("data", data, "data.Value", data.Value)
// 		reqLogger.Info(callName + " result")
// 	}
// 	if err != nil {
// 		log.Error(err, callName+" err")
// 	}
// 	//fmt.Println("\n")
// }

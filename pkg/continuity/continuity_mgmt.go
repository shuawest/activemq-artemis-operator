package continuity

import (
	"fmt"
	"os"
	"strconv"

	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/continuity/jolokia"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("continuity_mgmt")

// TODO: remove after testing
func init() {
	logf.SetLogger(zap.LoggerTo(os.Stdout))
}

// type IArtemisContinuity interface {
// 	NewArtemisContinuity(_ip string, _jolokiaPort string, _name string, _user string, _pass string) *ArtemisContinuity
// 	ConfigureContinuity(siteId string, peerSiteUrl string, localUser string) (*jolokia.ExecData, error)
// 	RemoveContinuity(siteId string) (*jolokia.ExecData, error)
// 	FetchContinuity(siteId string) (*jolokia.ExecData, error)
// 	ActivateSite(siteId string) (*jolokia.ExecData, error)
// }

type ArtemisContinuity struct {
	ip          string
	jolokiaPort string
	name        string
	user        string
	pass        string
	jolokia     *jolokia.Jolokia
}

func NewArtemisContinuity(_ip string, _jolokiaPort string, _name string, _user string, _pass string) *ArtemisContinuity {
	reqLogger := log.WithValues("ip", _ip, "port", _jolokiaPort, "name", _name, "user", _user, "pass", "***")
	reqLogger.Info("New artemis")

	artemiscty := ArtemisContinuity{
		ip:          _ip,
		jolokiaPort: _jolokiaPort,
		name:        _name,
		user:        _user,
		pass:        _pass,
		jolokia:     jolokia.NewJolokia(_ip, _jolokiaPort, "/console/jolokia", _user, _pass),
	}
	return &artemiscty
}

type BootedResult struct {
	Booted bool `json:"Booted,omitempty"`
}

func (artemis *ArtemisContinuity) IsBooted() (bool, error) {
	url := fmt.Sprintf(`org.apache.activemq.artemis:broker="%s",component=continuity,name="continuity.bootstrap"`, artemis.name)
	data, err := artemis.jolokia.Read(url)

	if data != nil {
		reqLogger := log.WithValues("artemis.name", artemis.name, "result", data.Value, "status", data.Status, "data.Value[Booted]", data.Value["Booted"])
		reqLogger.Info("IsBooted received data")

		return data.Value["Booted"].(bool), err
		// data.Value["Booted"]
		// bootedResult, ok := data.Value.(string)
		// if ok {
		// 	log.Info("IsBooted unmarshalled IS ok")

		// 	// bootedResult := &BootedResult{}
		// 	// err = json.Unmarshal([]byte(resBootedString), bootedResult)

		// 	// reqLogger2 := log.WithValues("bootedResult", bootedResult, "err", err)
		// 	// reqLogger2.Info("IsBooted unmarshalled data")

		// 	//return bootedResult.Booted, err
		// 	return false, err
		// } else {
		// 	log.Info("IsBooted unmarshalled NOT ok")
		// }
	}

	if err != nil {
		reqLogger := log.WithValues("artemis.name", artemis.name, "err", err)
		reqLogger.Info("IsBooted received data")
	}

	return false, err
}

func (artemis *ArtemisContinuity) Configure(siteId string, activeOnStart bool, servingAcceptors string, localConnectorRef string, remoteConnectorRefs string, reorgManagement bool) (*jolokia.ExecData, error) {
	reqLogger := log.WithValues("artemis.name", artemis.name)

	url := fmt.Sprintf(`org.apache.activemq.artemis:broker="%s",component=continuity,name=\"continuity.bootstrap\"`, artemis.name)
	operation := "configure(java.lang.String,java.lang.Boolean,java.lang.String,java.lang.String,java.lang.String,java.lang.Boolean)"
	parameters := fmt.Sprintf(`"%s",%t,"%s","%s","%s",%t`, siteId, activeOnStart, servingAcceptors, localConnectorRef, remoteConnectorRefs, reorgManagement)
	jsonStr := fmt.Sprintf(`{ "type":"EXEC","mbean":"%s","operation":"%s","arguments":[%s] }`, url, operation, parameters)

	reqLogger.Info("Sending exec url '" + url + "' json '" + jsonStr + "'")

	data, err := artemis.jolokia.Exec(url, jsonStr)

	if data != nil {
		reqLogger.Info("Configure received exec result '" + data.Value + "' status '" + strconv.Itoa(data.Status) + "', ip:port '" + artemis.ip + ":" + artemis.jolokiaPort)
	}

	if err != nil {
		reqLogger.Info("Configure received exec err '" + err.Error() + "'")
	}

	return data, err
}

func (artemis *ArtemisContinuity) SetSecrets(localContinuityUser string, localContinuityPass string, remoteContinuityUser string, remoteContinuityPass string) (*jolokia.ExecData, error) {
	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\",component=continuity,name=\\\"continuity.bootstrap\\\""
	parameters := `"` + localContinuityUser + `","` + localContinuityPass + `",` + `"` + remoteContinuityUser + `",` + `"` + remoteContinuityPass + `"`
	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"setSecrets(java.lang.String,java.lang.String,java.lang.String,java.lang.String)","arguments":[` + parameters + `] }`

	data, err := artemis.jolokia.Exec(url, jsonStr)

	return data, err
}

func (artemis *ArtemisContinuity) Tune(activationTimeout int, inflowStagingDelay int, bridgeInterval int, bridgeIntervalMultiplier float32, pollDuration int) (*jolokia.ExecData, error) {
	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\",component=continuity,name=\\\"continuity.bootstrap\\\""
	parameters := strconv.Itoa(activationTimeout) + `,` + strconv.Itoa(inflowStagingDelay) + `,` + strconv.Itoa(bridgeInterval) + `,` + fmt.Sprintf("%f", bridgeIntervalMultiplier) + `,` + strconv.Itoa(pollDuration)
	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"tune(java.lang.Long,java.lang.Long,java.lang.Long,java.lang.Double,java.lang.Long)","arguments":[` + parameters + `] }`

	data, err := artemis.jolokia.Exec(url, jsonStr)

	return data, err
}

func (artemis *ArtemisContinuity) Boot() (*jolokia.ExecData, error) {
	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\",component=continuity,name=\\\"continuity.bootstrap\\\""
	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"boot" }`

	data, err := artemis.jolokia.Exec(url, jsonStr)

	return data, err
}

func (artemis *ArtemisContinuity) RebootContinuity() (*jolokia.ExecData, error) {
	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\",component=continuity,name=\\\"continuity.bootstrap\\\""
	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"reboot" }`

	data, err := artemis.jolokia.Exec(url, jsonStr)

	return data, err
}

func (artemis *ArtemisContinuity) DestroyContinuity() (*jolokia.ExecData, error) {
	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\",component=continuity,name=\\\"continuity.bootstrap\\\""
	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"destory" }`

	data, err := artemis.jolokia.Exec(url, jsonStr)

	return data, err
}

func (artemis *ArtemisContinuity) ActivateSite(siteId string, activationTimeout int64) (*jolokia.ExecData, error) {
	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\",component=continuity,name=\\\"continuity.service\\\""
	parameters := strconv.FormatInt(activationTimeout, 10)
	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"activateSite(java.lang.Long)","arguments":[` + parameters + `]` + ` }`
	data, err := artemis.jolokia.Exec(url, jsonStr)

	return data, err
}

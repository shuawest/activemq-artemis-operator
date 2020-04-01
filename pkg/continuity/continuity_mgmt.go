package continuity

import (
	"fmt"
	"strconv"

	"github.com/roddiekieley/activemq-artemis-management/jolokia"
)

type IArtemisContinuity interface {
	NewArtemisContinuity(_ip string, _jolokiaPort string, _name string) *ArtemisContinuity
	ConfigureContinuity(siteId string, peerSiteUrl string, localUser string) (*jolokia.ExecData, error)
	RemoveContinuity(siteId string) (*jolokia.ExecData, error)
	FetchContinuity(siteId string) (*jolokia.ExecData, error)
	ActivateSite(siteId string) (*jolokia.ExecData, error)
}

type ArtemisContinuity struct {
	ip          string
	jolokiaPort string
	name        string
	jolokia     *jolokia.Jolokia
}

func NewArtemisContinuity(_ip string, _jolokiaPort string, _name string) *ArtemisContinuity {
	artemiscty := ArtemisContinuity{
		ip:          _ip,
		jolokiaPort: _jolokiaPort,
		name:        _name,
		jolokia:     jolokia.NewJolokia(_ip, _jolokiaPort, "/console/jolokia"),
	}
	return &artemiscty
}

func (artemis *ArtemisContinuity) IsBooted() (*jolokia.ExecData, error) {
	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\",component=continuity,name=\\\"continuity.bootstrap\\\""
	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"isBooted" }`

	data, err := artemis.jolokia.Exec(url, jsonStr)

	return data, err
}

func (artemis *ArtemisContinuity) Configure(siteId string, activeOnStart bool, servingAcceptors string, localConnectorRef string, remoteConnectorRefs string, reorgManagement bool) (*jolokia.ExecData, error) {
	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\",component=continuity,name=\\\"continuity.bootstrap\\\""
	parameters := `"` + siteId + `",` + strconv.FormatBool(activeOnStart) + `,` + `"` + servingAcceptors + `",` + `"` + localConnectorRef + `",` + `"` + remoteConnectorRefs + `",` + `` + strconv.FormatBool(reorgManagement)
	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"configure(java.lang.String,java.lang.Boolean,java.lang.String,java.lang.String,java.lang.String,java.lang.Boolean)","arguments":[` + parameters + `] }`

	data, err := artemis.jolokia.Exec(url, jsonStr)

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

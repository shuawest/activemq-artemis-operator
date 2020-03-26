package continuitymgmt

import (
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

func (artemis *ArtemisContinuity) ConfigureContinuity(siteId string, peerSiteUrl string, localUser string, localPass string) (*jolokia.ExecData, error) {
	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\",component=continuity,name=\\\"continuity.service\\\""
	parameters := `"` + siteId + `","` + peerSiteUrl + `",` + `"` + localUser + `",` + `"` + localPass + `"`
	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"configureContinuity(java.lang.String,java.lang.String,java.lang.String,java.lang.String)","arguments":[` + parameters + `]` + ` }`
	data, err := artemis.jolokia.Exec(url, jsonStr)

	return data, err
}

func (artemis *ArtemisContinuity) RemoveContinuity(siteId string) (*jolokia.ExecData, error) {
	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\",component=continuity,name=\\\"continuity.service\\\""
	parameters := `"` + siteId + `"`
	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"removeContinuity(java.lang.String)","arguments":[` + parameters + `]` + ` }`
	data, err := artemis.jolokia.Exec(url, jsonStr)

	return data, err
}

func (artemis *ArtemisContinuity) FetchContinuity(siteId string) (*jolokia.ExecData, error) {
	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\",component=continuity,name=\\\"continuity.service\\\""
	parameters := `"` + siteId + `"`
	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"getContinuityConfig(java.lang.String)","arguments":[` + parameters + `]` + ` }`
	data, err := artemis.jolokia.Exec(url, jsonStr)

	return data, err
}

func (artemis *ArtemisContinuity) ActivateSite(siteId string, activationTimeout int64) (*jolokia.ExecData, error) {
	url := "org.apache.activemq.artemis:broker=\\\"" + artemis.name + "\\\",component=continuity,name=\\\"continuity.service\\\""
	parameters := activationTimeout
	jsonStr := `{ "type":"EXEC","mbean":"` + url + `","operation":"activateSite(java.lang.Long)","arguments":[` + parameters + `]` + ` }`
	data, err := artemis.jolokia.Exec(url, jsonStr)

	return data, err
}

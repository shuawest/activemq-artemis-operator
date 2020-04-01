// +build !ignore_autogenerated

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v2alpha2

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha2.ActiveMQArtemis":       schema_pkg_apis_broker_v2alpha2_ActiveMQArtemis(ref),
		"github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha2.ActiveMQArtemisSpec":   schema_pkg_apis_broker_v2alpha2_ActiveMQArtemisSpec(ref),
		"github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha2.ActiveMQArtemisStatus": schema_pkg_apis_broker_v2alpha2_ActiveMQArtemisStatus(ref),
	}
}

func schema_pkg_apis_broker_v2alpha2_ActiveMQArtemis(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ActiveMQArtemis is the Schema for the activemqartemis API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha2.ActiveMQArtemisSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha2.ActiveMQArtemisStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha2.ActiveMQArtemisSpec", "github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha2.ActiveMQArtemisStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_broker_v2alpha2_ActiveMQArtemisSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ActiveMQArtemisSpec defines the desired state of ActiveMQArtemis",
				Properties: map[string]spec.Schema{
					"adminUser": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"adminPassword": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"deploymentPlan": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha2.DeploymentPlanType"),
						},
					},
					"acceptors": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha2.AcceptorType"),
									},
								},
							},
						},
					},
					"connectors": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha2.ConnectorType"),
									},
								},
							},
						},
					},
					"console": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha2.ConsoleType"),
						},
					},
					"version": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"upgrades": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha2.ActiveMQArtemisUpgrades"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha2.AcceptorType", "github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha2.ActiveMQArtemisUpgrades", "github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha2.ConnectorType", "github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha2.ConsoleType", "github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha2.DeploymentPlanType"},
	}
}

func schema_pkg_apis_broker_v2alpha2_ActiveMQArtemisStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ActiveMQArtemisStatus defines the observed state of ActiveMQArtemis",
				Properties: map[string]spec.Schema{
					"podStatus": {
						SchemaProps: spec.SchemaProps{
							Description: "INSERT ADDITIONAL STATUS FIELD - define observed state of cluster Important: Run \"operator-sdk generate k8s\" to regenerate code after modifying this file Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html",
							Ref:         ref("github.com/RHsyseng/operator-utils/pkg/olm.DeploymentStatus"),
						},
					},
				},
				Required: []string{"podStatus"},
			},
		},
		Dependencies: []string{
			"github.com/RHsyseng/operator-utils/pkg/olm.DeploymentStatus"},
	}
}
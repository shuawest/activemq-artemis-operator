// +build !ignore_autogenerated

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v2alpha2

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"github.com/rh-messaging/activemq-artemis-operator/pkg/apis/continuity/v2alpha2.ActiveMQArtemisContinuity":       schema_pkg_apis_continuity_v2alpha2_ActiveMQArtemisContinuity(ref),
		"github.com/rh-messaging/activemq-artemis-operator/pkg/apis/continuity/v2alpha2.ActiveMQArtemisContinuitySpec":   schema_pkg_apis_continuity_v2alpha2_ActiveMQArtemisContinuitySpec(ref),
		"github.com/rh-messaging/activemq-artemis-operator/pkg/apis/continuity/v2alpha2.ActiveMQArtemisContinuityStatus": schema_pkg_apis_continuity_v2alpha2_ActiveMQArtemisContinuityStatus(ref),
	}
}

func schema_pkg_apis_continuity_v2alpha2_ActiveMQArtemisContinuity(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ActiveMQArtemisContinuity is the Schema for the ActiveMQArtemisContinuity API",
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
							Ref: ref("github.com/rh-messaging/activemq-artemis-operator/pkg/apis/continuity/v2alpha2.ActiveMQArtemisContinuitySpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/rh-messaging/activemq-artemis-operator/pkg/apis/continuity/v2alpha2.ActiveMQArtemisContinuityStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/rh-messaging/activemq-artemis-operator/pkg/apis/continuity/v2alpha2.ActiveMQArtemisContinuitySpec", "github.com/rh-messaging/activemq-artemis-operator/pkg/apis/continuity/v2alpha2.ActiveMQArtemisContinuityStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_continuity_v2alpha2_ActiveMQArtemisContinuitySpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ActiveMQArtemisContinuitySpec defines the desired state of ActiveMQArtemisContinuity",
				Properties: map[string]spec.Schema{
					"siteId": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"localContinuityUser": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"localContinuityPass": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"remoteContinuityPass": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"remoteContinuityUser": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"activeOnStart": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"inflowStagingDelay": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"bridgeInterval": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"bridgeIntervalMultiplier": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"number"},
							Format: "float",
						},
					},
					"pollDuration": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"activationTimeout": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"reorgManagement": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
				},
			},
		},
		Dependencies: []string{},
	}
}

func schema_pkg_apis_continuity_v2alpha2_ActiveMQArtemisContinuityStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ActiveMQArtemisContinuityStatus defines the observed state of ActiveMQArtemisContinuity",
				Properties:  map[string]spec.Schema{},
			},
		},
		Dependencies: []string{},
	}
}
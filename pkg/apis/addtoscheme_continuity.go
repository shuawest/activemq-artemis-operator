package apis

import (
	routev1 "github.com/openshift/api/route/v1"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/apis/continuity/v2alpha2"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v2alpha2.SchemeBuilder.AddToScheme, routev1.SchemeBuilder.AddToScheme)
}

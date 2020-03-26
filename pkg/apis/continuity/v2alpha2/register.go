// NOTE: Boilerplate only.  Ignore this file.

// Package v2alpha2 contains API Schema definitions for the continuity v2alpha2 API group
// +k8s:deepcopy-gen=package,register
// +groupName=continuity.amq.io
package v2alpha2

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/runtime/scheme"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "continuity.amq.io", Version: "v2alpha2"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

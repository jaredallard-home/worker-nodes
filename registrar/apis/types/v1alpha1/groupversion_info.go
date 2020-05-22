// Package v1 contains API Schema definitions for the registrar v1 API group
// +kubebuilder:object:generate=true
// +groupName=registrar.jaredallard.me
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

//go:generate controller-gen crd output:dir=../../../config paths=./...
//go:generate controller-gen object paths=./...

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "registrar.jaredallard.me", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

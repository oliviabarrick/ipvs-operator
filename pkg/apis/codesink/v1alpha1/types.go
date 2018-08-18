package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WeightedServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []WeightedService `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WeightedService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              WeightedServiceSpec   `json:"spec"`
	Status            WeightedServiceStatus `json:"status,omitempty"`
}

type WeightedServiceSpec struct {
	*corev1.ServiceSpec
	Weights []ServiceWeight `json:"weights,omitempty"`
}

type ServiceWeight struct {
	Weight   int               `json:"weight"`
	Selector map[string]string `json:"selector"`
}

type WeightedServiceStatus struct {
	// Fill me
}

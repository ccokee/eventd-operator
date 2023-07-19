// publisher_types.go

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// PublisherSpec defines the desired state of Publisher
type PublisherSpec struct {
	// GCP service account key in base64-encoded format
	GCPSAKey string `json:"gcpServiceAccountKey,omitempty"`

	// GCP project ID
	ProjectID string `json:"projectID,omitempty"`

	// Pub/Sub topic to publish the Kubernetes events
	Topic string `json:"topic,omitempty"`

	// List of allowed Kubernetes event types to watch
	AllowedMessageTypes []string `json:"allowedMessageTypes,omitempty"`
}

// PublisherStatus defines the observed state of Publisher
type PublisherStatus struct {
	// Conditions represent the latest available observations of an object's state
	Conditions []PublisherCondition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Publisher is the Schema for the publishers API
type Publisher struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PublisherSpec   `json:"spec,omitempty"`
	Status PublisherStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PublisherList contains a list of Publisher
type PublisherList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Publisher `json:"items"`
}

type PublisherConditionType string

const (
	// ConditionConfigured represents that the publisher is properly configured.
	ConditionConfigured PublisherConditionType = "Configured"
)

// PublisherCondition defines condition for the publisher
type PublisherCondition struct {
	Type    PublisherConditionType `json:"type"`
	Status  metav1.ConditionStatus `json:"status"`
	Reason  string                 `json:"reason,omitempty"`
	Message string                 `json:"message,omitempty"`
}

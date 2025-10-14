package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EventRecyclerSpec defines the desired state of EventRecycler
type EventRecyclerSpec struct {
	// WatchInterval specifies how often to check for error events
	// +kubebuilder:default="60s"
	WatchInterval metav1.Duration `json:"watchInterval,omitempty"`

	// SearchTerms are the error messages to watch for in events
	// +kubebuilder:validation:MinItems=1
	SearchTerms []string `json:"searchTerms"`

	// Threshold is the number of matching events before triggering a recycle
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=5
	Threshold int `json:"threshold,omitempty"`

	// DryRun when true will log actions without actually recycling node groups
	// +kubebuilder:default=false
	DryRun bool `json:"dryRun,omitempty"`

	// AWSRegion specifies the AWS region for node group operations
	// +optional
	AWSRegion string `json:"awsRegion,omitempty"`

	// PollInterval specifies the interval for polling instance states during recycle
	// +kubebuilder:default="15s"
	PollInterval metav1.Duration `json:"pollInterval,omitempty"`

	// RecycleTimeout specifies the maximum time to wait for a recycle operation
	// +kubebuilder:default="20m"
	RecycleTimeout metav1.Duration `json:"recycleTimeout,omitempty"`
}

// EventRecyclerStatus defines the observed state of EventRecycler
type EventRecyclerStatus struct {
	// LastCheckTime is the last time events were checked
	LastCheckTime metav1.Time `json:"lastCheckTime,omitempty"`

	// ActiveRecycles lists node groups currently being recycled
	ActiveRecycles []string `json:"activeRecycles,omitempty"`

	// RecycleHistory tracks recent recycle operations
	RecycleHistory []RecycleHistoryEntry `json:"recycleHistory,omitempty"`

	// EventCounts tracks event counts per node group
	EventCounts map[string]int `json:"eventCounts,omitempty"`
}

// RecycleHistoryEntry represents a single recycle operation
type RecycleHistoryEntry struct {
	// NodeGroup is the name of the recycled node group
	NodeGroup string `json:"nodeGroup"`

	// Timestamp when the recycle was triggered
	Timestamp metav1.Time `json:"timestamp"`

	// EventCount is the number of events that triggered the recycle
	EventCount int `json:"eventCount"`

	// Status of the recycle operation
	Status string `json:"status"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// EventRecycler is the Schema for the eventrecyclers API
type EventRecycler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EventRecyclerSpec   `json:"spec,omitempty"`
	Status EventRecyclerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// EventRecyclerList contains a list of EventRecycler
type EventRecyclerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EventRecycler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EventRecycler{}, &EventRecyclerList{})
}

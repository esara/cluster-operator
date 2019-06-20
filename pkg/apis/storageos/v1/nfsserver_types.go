/*
Copyright 2018 The Rook Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DefaultNFSContainerImage is the name of the Ganesha container to run.
	// TODO: change to an image we maintain.
	DefaultNFSContainerImage = "storageos/nfs-ganesha"

	// DefaultSize is used when no Size is
	DefaultSize = "5Gi"

	PhasePending = "Pending"
	PhaseRunning = "Running"
	// PhaseSucceeded = "Succeeded"
	// PhaseFailed    = "Failed"
	PhaseUnknown = "Unknown"
)

// NFSServer is the Schema for the nfsservers API.
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type NFSServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NFSServerSpec   `json:"spec,omitempty"`
	Status NFSServerStatus `json:"status,omitempty"`
}

// NFSServerList contains a list of NFSServer
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type NFSServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NFSServer `json:"items"`
}

// NFSServerSpec defines the desired state of NFSServer
type NFSServerSpec struct {
	Size string

	// NFSContainer is the container image to use for the NFS server.
	NFSContainer string `json:"nfsContainer"`

	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

	// The annotations-related configuration to add/set on each Pod related object.
	Annotations map[string]string `json:"annotations,omitempty"`

	// Replicas of the NFS daemon
	// TODO(simon): Don't think we can have multiple servers?
	// Replicas int `json:"replicas,omitempty"`

	// The parameters to configure the NFS export
	Exports []ExportsSpec `json:"exports,omitempty"`

	// Reclamation policy for the persistent volume shared to the user's pod.
	PersistentVolumeReclaimPolicy v1.PersistentVolumeReclaimPolicy

	// PV mount options. Not validated - mount of the PVs will simply fail if
	// one is invalid.
	MountOptions []string
}

// GetSize returns the requested volume size.
func (s NFSServerSpec) GetSize() string {
	if s.Size != "" {
		return s.Size
	}
	return DefaultSize
}

// GetContainerImage returns the NFS server container image.
func (s NFSServerSpec) GetContainerImage() string {
	if s.NFSContainer != "" {
		return s.NFSContainer
	}
	return DefaultNFSContainerImage
}

// ExportsSpec represents the spec of NFS exports
type ExportsSpec struct {
	// Name of the export
	Name string `json:"name,omitempty"`

	// The NFS server configuration
	Server ServerSpec `json:"server,omitempty"`

	// PVC from which the NFS daemon gets storage for sharing
	PersistentVolumeClaim v1.PersistentVolumeClaimVolumeSource `json:"persistentVolumeClaim,omitempty"`
}

// ServerSpec represents the spec for configuring the NFS server
type ServerSpec struct {
	// Reading and Writing permissions on the export
	// Valid values are "ReadOnly", "ReadWrite" and "none"
	AccessMode string `json:"accessMode,omitempty"`

	// This prevents the root users connected remotely from having root privileges
	// Valid values are "none", "rootid", "root", and "all"
	Squash string `json:"squash,omitempty"`

	// The clients allowed to access the NFS export
	AllowedClients []AllowedClientsSpec `json:"allowedClients,omitempty"`
}

// AllowedClientsSpec represents the client specs for accessing the NFS export
type AllowedClientsSpec struct {

	// Name of the clients group
	Name string `json:"name,omitempty"`

	// The clients that can access the share
	// Values can be hostname, ip address, netgroup, CIDR network address, or all
	Clients []string `json:"clients,omitempty"`

	// Reading and Writing permissions for the client to access the NFS export
	// Valid values are "ReadOnly", "ReadWrite" and "none"
	// Gets overridden when ServerSpec.accessMode is specified
	AccessMode string `json:"accessMode,omitempty"`

	// Squash options for clients
	// Valid values are "none", "rootid", "root", and "all"
	// Gets overridden when ServerSpec.squash is specified
	Squash string `json:"squash,omitempty"`
}

// NFSServerStatus defines the observed state of the NFSServer.
type NFSServerStatus struct {

	// RemoteTarget is the connection string that clients can use to access the
	// shared filesystem.
	RemoteTarget string `json:"remoteTarget"`

	// Phase is a simple, high-level summary of where the NFS Server is in its
	// lifecycle. Phase will be set to Ready when the NFS Server is ready for
	// use.  It is intended to be similar to the PodStatus Phase described at:
	// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#podstatus-v1-core
	//
	// There are five possible phase values:
	//   - Pending: The NFS Server has been accepted by the Kubernetes system,
	//     but one or more of the components has not been created. This includes
	//     time before being scheduled as well as time spent downloading images
	//     over the network, which could take a while.
	//   - Running: The NFS Server has been bound to a node, and all of the
	//     dependencies have been created.
	//   - Succeeded: All NFS Server dependencies have terminated in success,
	//     and will not be restarted.
	//   - Failed: All NFS Server dependencies in the pod have terminated, and
	//     at least one container has terminated in failure. The container
	//     either exited with non-zero status or was terminated by the system.
	//   - Unknown: For some reason the state of the NFS Server could not be
	//     obtained, typically due to an error in communicating with the host of
	//     the pod.
	//
	// TODO(sc): do we want to add more info, e.g. Condition, messages or
	// StartTime?
	Phase string `json:"phase"`
}

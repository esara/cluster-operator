package nfsserver

import (
	storageosv1alpha1 "github.com/storageos/cluster-operator/pkg/apis/storageos/v1alpha1"
	"github.com/storageos/cluster-operator/pkg/nfs"
)

// Server stores the current NFS server's information. It binds the
// server and the deployment together, ensuring deployment interacts with the
// right server resource.
type Server struct {
	cached *storageosv1alpha1.NFSServer
	// deployment implements the Deployment interface. This is
	// cached for a cluserverster to avoid recreating it without any change to
	// the server cached. Every new server will create its unique deployment.
	deployment Deployment
}

// NewServer creates a new NFS Server.  It caches the passed NFSServer object.
func NewServer(obj *storageosv1alpha1.NFSServer) *Server {
	return &Server{cached: obj}
}

// SetDeployment creates a new Server Deployment and sets it for the current
// NFSServer.
func (s *Server) SetDeployment(r *ReconcileNFSServer) {
	s.deployment = nfs.NewDeployment(r.client, s.cached, r.recorder, r.scheme)
}

// IsCurrentServer compares the server attributes to check if the given
// server is the same as the current server.
func (s *Server) IsCurrentServer(server *storageosv1alpha1.NFSServer) bool {
	if (s.cached.GetName() == server.GetName()) &&
		(s.cached.GetNamespace() == server.GetNamespace()) {
		return true
	}
	return false
}

// Deploy deploys the StorageOS cluster.
func (s *Server) Deploy(r *ReconcileNFSServer) error {

	if s.deployment == nil {
		s.SetDeployment(r)
	}
	return s.deployment.Deploy()
}

// DeleteDeployment deletes the StorageOS Cluster deployment.
func (s *Server) DeleteDeployment() error {
	return s.deployment.Delete()
}

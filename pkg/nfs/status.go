package nfs

import (
	"context"
	"fmt"
	"reflect"

	storageosv1 "github.com/storageos/cluster-operator/pkg/apis/storageos/v1"
	v1 "k8s.io/api/core/v1"
)

func (s *Deployment) updateStatus(status *storageosv1.NFSServerStatus) error {

	if reflect.DeepEqual(s.nfsServer.Status, *status) {
		return nil
	}

	// When there's a difference in node ready count, broadcast the health change event.
	if s.nfsServer.Status.RemoteTarget != status.RemoteTarget {

		if s.nfsServer.Status.RemoteTarget != "" {
			if s.recorder != nil {
				s.recorder.Event(s.nfsServer, v1.EventTypeNormal, "ChangedStatus", fmt.Sprintf("NFS server is now functional", status.RemoteTarget))
			}
		} else {
			if s.recorder != nil {
				s.recorder.Event(s.nfsServer, v1.EventTypeWarning, "ChangedStatus", "NFS server not listening")
			}
		}
	}
	s.nfsServer.Status = *status
	return s.client.Update(context.Background(), s.nfsServer)
}

// getStatus determines the status of the NFS Server deployment.
func (s *Deployment) getStatus() (*storageosv1.NFSServerStatus, error) {

	status := &storageosv1.NFSServerStatus{
		Phase:        storageosv1.PhaseUnknown,
		RemoteTarget: "",
	}

	ss, err := s.getStatefulSet(s.nfsServer.Name, s.nfsServer.Namespace)
	if err != nil {
		return status, err
	}

	svc, err := s.getService(s.nfsServer.Name, s.nfsServer.Namespace)
	if err != nil {
		return status, err
	}

	// We got both without error, so upgrade to Pending.
	status.Phase = storageosv1.PhasePending

	// If the service is created, set the cluster address as the endpoint.
	if svc.Spec.ClusterIP != "" {
		status.RemoteTarget = svc.Spec.ClusterIP

		// If the NFS pod is also ready, then we can mark the NFS Server as
		// online.
		if ss.Status.ReadyReplicas > 0 {
			status.Phase = storageosv1.PhaseRunning
		}
	}

	return status, nil
}

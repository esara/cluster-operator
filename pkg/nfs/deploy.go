package nfs

import (
	"errors"

	storageosv1 "github.com/storageos/cluster-operator/pkg/apis/storageos/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	appName         = "storageos"
	statefulsetKind = "statefulset"

	DefaultNFSPort     = 2049
	DefaultRPCPort     = 111
	DefaultMetricsPort = 9587
)

// Deploy creates all the resources required to provision an NFS PV on top of
// a StorageOS block device.
func (d *Deployment) Deploy() error {

	// Only set size if given, nil otherwise.  Will use block volume default.
	var size *resource.Quantity
	requestedCapacity, ok := d.nfsServer.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	if ok {
		size = &requestedCapacity
	}

	_, err := d.ensureService(DefaultNFSPort, DefaultRPCPort, DefaultMetricsPort)
	if err != nil {
		return err
	}
	if err := d.createNFSConfigMap(); err != nil {
		return err
	}
	if err := d.createStatefulSet(size, DefaultNFSPort, DefaultRPCPort, DefaultMetricsPort); err != nil {
		return err
	}
	// if err := d.createPV(svc.Spec.ClusterIP, "/export", size); err != nil {
	// 	return err
	// }

	status, err := d.getStatus()
	if err != nil {
		return err
	}

	d.logger.WithValues("status", status).V(4).Info("Updating status")

	if err := d.updateStatus(status); err != nil {
		return err
	}

	if status.Phase != storageosv1.PhaseRunning {
		return errors.New("NFS server not ready")
	}

	return nil
}

func labelsForStatefulSet(name string) map[string]string {
	return map[string]string{"app": appName, "nfsserver": name}
}

// func labelsForApp(nfsServer *storageosv1.NFSServer) map[string]string {
// 	return map[string]string{
// 		"app": nfsServer.Name,
// 	}
// }

func addOwnerRefToObject(obj metav1.Object, ownerRef metav1.OwnerReference) {
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

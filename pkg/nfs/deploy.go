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

// Due to https://github.com/kubernetes/kubernetes/issues/74916 fixed in
// 1.15, labels intended for the PVC must be set on the Pod template.
// In 1.15 and later we can just set the "app" and "nfsserver" labels here.  For
// now, pass all labels rather than check k8s versions.  The only downside is
// that the nfs pod gets storageos.com labels that don't do anything directly.
func labelsForStatefulSet(name string, labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}

	labels["app"] = appName
	labels["nfsserver"] = name

	// TODO: setting fenced should only be done if we _know_ that fencing hasn't been disabled else provisioning will fail
	labels["storageos.com/fenced"] = "true"
	return labels
}

func addOwnerRefToObject(obj metav1.Object, ownerRef metav1.OwnerReference) {
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

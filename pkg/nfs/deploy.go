package nfs

import (
	"log"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	appName         = "storageos"
	statefulsetKind = "statefulset"

	DefaultNFSPort = 2049
	DefaultRPCPort = 111
)

// Deploy creates all the resources required to provision an NFS PV on top of
// a StorageOS block device.
func (d *Deployment) Deploy() error {

	log.Printf("Deploy: %#v\n", d.nfsServer)

	size, err := resource.ParseQuantity(d.nfsServer.Spec.GetSize())
	if err != nil {
		return err
	}

	_, err = d.ensureService(DefaultNFSPort, DefaultRPCPort)
	if err != nil {
		return err
	}
	if err := d.createNFSConfigMap(); err != nil {
		return err
	}
	if err := d.createStatefulSet(size, DefaultNFSPort, DefaultRPCPort); err != nil {
		return err
	}
	// if err := d.createPV(svc.Spec.ClusterIP, "/export", size); err != nil {
	// 	return err
	// }

	status, err := d.getStatus()
	if err != nil {
		return err
	}

	log.Printf("Updating status: %v", status)

	if err := d.updateStatus(status); err != nil {
		return err
	}

	return nil
}

func labelsForStatefulSet(name string) map[string]string {
	return map[string]string{"app": appName, "storageos_cr": name}
}

// func labelsForApp(nfsServer *storageosv1.NFSServer) map[string]string {
// 	return map[string]string{
// 		"app": nfsServer.Name,
// 	}
// }

func addOwnerRefToObject(obj metav1.Object, ownerRef metav1.OwnerReference) {
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

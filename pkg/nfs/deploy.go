package nfs

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	appName         = "storageos"
	statefulsetKind = "statefulset"
)

// Deploy creates all the resources required to provision an NFS PV on top of
// a StorageOS block device.
func (d *Deployment) Deploy() error {

	if err := d.createStatefulSet(); err != nil {
		return err
	}
	return nil
}

// func (d *Deployment) createStatefulSet() error {

// 	ss := appsv1.StatefulSet{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      d.nfsServer.Name,
// 			Namespace: d.nfsServer.Namespace,
// 			// Labels:          createAppLabels(d.nfsServer),
// 			OwnerReferences: []metav1.OwnerReference{d.nfsServer.ownerRef},
// 		},
// 		Spec: appsv1.StatefulSetSpec{
// 			Replicas: &replicas,
// 			Selector: &metav1.LabelSelector{
// 				// MatchLabels: createAppLabels(nfsServer),
// 			},
// 			Template:    nfsPodSpec,
// 			ServiceName: d.nfsServer.name,
// 		},
// 	}
// }

// createPVC returns a busybox pod with the same name/namespace as the cr
// func (d *Deployment) createPVC() error {

// 	scName := "fast"

// 	pvc := &corev1.PersistentVolumeClaim{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      d.nfsServer.Name,
// 			Namespace: d.nfsServer.Namespace,
// 		},
// 		Spec: corev1.PersistentVolumeClaimSpec{
// 			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
// 			StorageClassName: &scName,
// 		},
// 	}
// 	return d.createOrUpdateObject(pvc)
// }

func labelsForStatefulSet(name string) map[string]string {
	return map[string]string{"app": appName, "storageos_cr": name, "kind": statefulsetKind}
}

func addOwnerRefToObject(obj metav1.Object, ownerRef metav1.OwnerReference) {
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

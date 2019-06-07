package nfs

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *Deployment) createPV(server string, path string, size resource.Quantity) error {

	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:            d.nfsServer.Name,
			OwnerReferences: d.nfsServer.ObjectMeta.OwnerReferences,
		},
		Spec: corev1.PersistentVolumeSpec{
			// TODO: not sure we can pass in the reclaim policy from anywhere?
			// PersistentVolumeReclaimPolicy: d.nfsServer.Spec.PersistentVolumeReclaimPolicy,
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			MountOptions: d.nfsServer.Spec.MountOptions,
			Capacity: corev1.ResourceList{
				corev1.ResourceName(corev1.ResourceStorage): size,
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server:   server,
					Path:     path,
					ReadOnly: false,
				},
			},
		},
	}

	return d.createOrUpdateObject(pv)
}

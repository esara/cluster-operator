package nfsserver

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deploy creates all the resources required to provision an NFS PV on top of
// a StorageOS block device.
func (s *Deployment) Deploy() error {

	if err := s.createPVC(); err != nil {
		return err
	}
	return nil
}

// createPVC returns a busybox pod with the same name/namespace as the cr
func (s *Deployment) createPVC() error {

	scName := "fast"

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.nfsServer.Name,
			Namespace: s.nfsServer.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			StorageClassName: &scName,
		},
	}
	return s.createOrUpdateObject(pvc)
}

package storageos

// import (
// 	storagev1 "k8s.io/api/storage/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// )

// func (s *Deployment) createNFSStorageClass() error {

// 	sc := s.getNFSStorageClass("nfs")
// 	sc.Parameters = map[string]string{
// 		"pool": "default",
// 	}

// 	// Add CSI creds secrets in parameters.
// 	if CSIV1Supported(s.k8sVersion) {
// 		// New CSI secret parameter keys were introduced in CSI v1.
// 		sc.Parameters[csiV1FSType] = defaultFSType
// 		if s.stos.Spec.CSI.EnableProvisionCreds {
// 			sc.Parameters[csiV1ProvisionerSecretNameKey] = csiProvisionerSecretName
// 			sc.Parameters[csiV1ProvisionerSecretNamespaceKey] = s.stos.Spec.GetResourceNS()
// 		}
// 		if s.stos.Spec.CSI.EnableControllerPublishCreds {
// 			sc.Parameters[csiV1ControllerPublishSecretNameKey] = csiControllerPublishSecretName
// 			sc.Parameters[csiV1ControllerPublishSecretNamespaceKey] = s.stos.Spec.GetResourceNS()
// 		}
// 		if s.stos.Spec.CSI.EnableNodePublishCreds {
// 			sc.Parameters[csiV1NodePublishSecretNameKey] = csiNodePublishSecretName
// 			sc.Parameters[csiV1NodePublishSecretNamespaceKey] = s.stos.Spec.GetResourceNS()
// 		}
// 	} else {
// 		sc.Parameters[fsType] = defaultFSType
// 		if s.stos.Spec.CSI.EnableProvisionCreds {
// 			sc.Parameters[csiV0ProvisionerSecretNameKey] = csiProvisionerSecretName
// 			sc.Parameters[csiV0ProvisionerSecretNamespaceKey] = s.stos.Spec.GetResourceNS()
// 		}
// 		if s.stos.Spec.CSI.EnableControllerPublishCreds {
// 			sc.Parameters[csiV0ControllerPublishSecretNameKey] = csiControllerPublishSecretName
// 			sc.Parameters[csiV0ControllerPublishSecretNamespaceKey] = s.stos.Spec.GetResourceNS()
// 		}
// 		if s.stos.Spec.CSI.EnableNodePublishCreds {
// 			sc.Parameters[csiV0NodePublishSecretNameKey] = csiNodePublishSecretName
// 			sc.Parameters[csiV0NodePublishSecretNamespaceKey] = s.stos.Spec.GetResourceNS()
// 		}
// 	}

// 	return s.createOrUpdateObject(sc)
// }

// func (s *Deployment) getNFSStorageClass(name string) *storagev1.StorageClass {
// 	return &storagev1.StorageClass{
// 		TypeMeta: metav1.TypeMeta{
// 			APIVersion: "storage.k8s.io/v1",
// 			Kind:       "StorageClass",
// 		},
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name: name,
// 			Labels: map[string]string{
// 				"app": appName,
// 			},
// 		},
// 		Provisioner: nfsProvisionerName,
// 	}
// }

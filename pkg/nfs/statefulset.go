package nfs

import (
	"log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *Deployment) createStatefulSet() error {

	// ss := &appsv1.StatefulSet{}
	replicas := int32(1)

	ss := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            d.nfsServer.Name,
			Namespace:       d.nfsServer.Namespace,
			Labels:          labelsForStatefulSet(d.nfsServer.Name),
			OwnerReferences: d.nfsServer.ObjectMeta.OwnerReferences,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: "storageos",
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsForStatefulSet(d.nfsServer.Name),
			},
			Template:             d.createPodTemplateSpec(),
			VolumeClaimTemplates: d.createVolumeClaimTemplateSpecs(),
		},
	}

	log.Printf("ss: %#v", ss)

	// podSpec := &sset.Spec.Template.Spec

	// s.addPodPriorityClass(podSpec)

	// s.addNodeAffinity(podSpec)

	// if err := s.addTolerations(podSpec); err != nil {
	// 	return err
	// }

	return d.createOrUpdateObject(ss)
}

func (d *Deployment) createVolumeClaimTemplateSpecs() []corev1.PersistentVolumeClaim {

	scName := "fast"

	return []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      d.nfsServer.Name,
				Namespace: d.nfsServer.Namespace,
				Labels:    labelsForStatefulSet(d.nfsServer.Name),
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				VolumeName:       d.nfsServer.Name,
				AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				StorageClassName: &scName,
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceName(v1.ResourceStorage): d.nfsServer.Size,
					},
				},
			},
		},
	}
}

func (d *Deployment) createPodTemplateSpec() corev1.PodTemplateSpec {

	// TODO pass in as params
	nfsPort := 2049
	rpcPort := 111

	return v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.nfsServer.Name,
			Namespace: d.nfsServer.Namespace,
			Labels:    labelsForStatefulSet(d.nfsServer.Name),
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					ImagePullPolicy: "IfNotPresent",
					Name:            d.nfsServer.Name,
					Image:           d.nfsServer.Spec.GetContainerImage(),
					// Args: []string{"nfs", "server", "--ganeshaConfigPath=" + NFSConfigMapPath + "/" + nfsServer.name},
					Ports: []v1.ContainerPort{
						{
							Name:          "nfs-port",
							ContainerPort: int32(nfsPort),
						},
						{
							Name:          "rpc-port",
							ContainerPort: int32(rpcPort),
						},
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      d.nfsServer.Name,
							MountPath: "/export",
						},
					},
					SecurityContext: &v1.SecurityContext{
						Capabilities: &v1.Capabilities{
							Add: []v1.Capability{
								"SYS_ADMIN",
								"DAC_READ_SEARCH",
							},
						},
					},
				},
			},
		},
	}
}

func (d *Deployment) deleteStatefulSet(name string) error {
	return d.deleteObject(d.getStatefulSet(name))
}

func (d *Deployment) getStatefulSet(name string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: d.nfsServer.Namespace,
			Labels: map[string]string{
				"app": "storageos",
			},
		},
	}
}

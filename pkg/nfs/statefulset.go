package nfs

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *Deployment) createStatefulSet() error {

	// ss := &appsv1.StatefulSet{}
	ls := labelsForStatefulSet(d.nfsServer.Name)
	replicas := int32(1)

	nfsPodSpec := d.createNfsPodSpec()

	ss := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.nfsServer.Name,
			Namespace: d.nfsServer.Namespace,
			Labels: map[string]string{
				"app": "storageos",
			},
			OwnerReferences: d.nfsServer.ObjectMeta.OwnerReferences,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: "storageos",
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: nfsPodSpec,
		},
	}

	// podSpec := &sset.Spec.Template.Spec

	// s.addPodPriorityClass(podSpec)

	// s.addNodeAffinity(podSpec)

	// if err := s.addTolerations(podSpec); err != nil {
	// 	return err
	// }

	return d.createOrUpdateObject(ss)
}

func (d *Deployment) createNfsPodSpec() corev1.PodTemplateSpec {
	return v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.nfsServer.Name,
			Namespace: d.nfsServer.Namespace,
			// Labels:    createAppLabels(nfsServer),
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					ImagePullPolicy: "IfNotPresent",
					Name:            d.nfsServer.Name,
					// Image:           d.containerImage,
					// Args: []string{"nfs", "server", "--ganeshaConfigPath=" + NFSConfigMapPath + "/" + nfsServer.name},
					Ports: []v1.ContainerPort{
						// {
						// 	Name:          "nfs-port",
						// 	ContainerPort: int32(nfsPort),
						// },
						// {
						// 	Name:          "rpc-port",
						// 	ContainerPort: int32(rpcPort),
						// },
					},
					// VolumeMounts: createVolumeMountList(nfsServer),
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
			// Volumes: createPVCSpecList(nfsServer),
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

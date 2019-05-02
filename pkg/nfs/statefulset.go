package nfs

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Deployment) createStatefulSet() error {
	ls := labelsForStatefulSet(s.nfsServer.Name)
	replicas := int32(1)

	sset := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.nfsServer.Name,
			Namespace: s.nfsServer.Namespace,
			Labels: map[string]string{
				"app": "storageos",
			},
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: "storageos",
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: StatefulsetSA,
					Containers: []corev1.Container{
						{
							Image:           s.stos.Spec.GetCSIExternalProvisionerImage(CSIV1Supported(s.k8sVersion)),
							Name:            "csi-external-provisioner",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Args: []string{
								"--v=5",
								"--provisioner=storageos",
								"--csi-address=$(ADDRESS)",
							},
							Env: []corev1.EnvVar{
								{
									Name:  addressEnvVar,
									Value: "/csi/csi.sock",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "nfs-data",
									MountPath: "/csi",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "nfs-data",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: s.stos.Spec.GetCSIPluginDir(CSIV1Supported(s.k8sVersion)),
									Type: &hostpathDirOrCreate,
								},
							},
						},
					},
				},
			},
		},
	}

	podSpec := &sset.Spec.Template.Spec

	s.addPodPriorityClass(podSpec)

	s.addNodeAffinity(podSpec)

	if err := s.addTolerations(podSpec); err != nil {
		return err
	}

	return s.createOrUpdateObject(sset)
}

func (s *Deployment) deleteStatefulSet(name string) error {
	return s.deleteObject(s.getStatefulSet(name))
}

func (s *Deployment) getStatefulSet(name string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: s.stos.Spec.GetResourceNS(),
			Labels: map[string]string{
				"app": "storageos",
			},
		},
	}
}

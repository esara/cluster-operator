package nfs

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (d *Deployment) createStatefulSet(size *resource.Quantity, nfsPort int, rpcPort int, metricsPort int) error {

	replicas := int32(1)

	ss := &appsv1.StatefulSet{
		// TODO - TypeMeta not needed?
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.nfsServer.Name,
			Namespace: d.nfsServer.Namespace,
			// Labels:          labelsForStatefulSet(d.nfsServer.Name),
			OwnerReferences: d.nfsServer.ObjectMeta.OwnerReferences,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: d.nfsServer.Name,
			Replicas:    &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsForStatefulSet(d.nfsServer.Name),
			},
			Template:             d.createPodTemplateSpec(nfsPort, rpcPort, metricsPort),
			VolumeClaimTemplates: d.createVolumeClaimTemplateSpecs(size),
		},
	}

	// log.Printf("ss: %#v", ss)

	// podSpec := &sset.Spec.Template.Spec

	// s.addPodPriorityClass(podSpec)

	// s.addNodeAffinity(podSpec)

	// if err := s.addTolerations(podSpec); err != nil {
	// 	return err
	// }

	return d.createOrUpdateObject(ss)
}

func (d *Deployment) createVolumeClaimTemplateSpecs(size *resource.Quantity) []corev1.PersistentVolumeClaim {

	// TODO: constant/lookup
	scName := "fast"

	claim := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			// Name:      d.nfsServer.Name,
			Name:      "nfs-data",
			Namespace: d.nfsServer.Namespace,
			Labels:    labelsForStatefulSet(d.nfsServer.Name),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			StorageClassName: &scName,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{},
			},
		},
	}

	if size != nil {
		claim.Spec.Resources.Requests = corev1.ResourceList{
			corev1.ResourceName(corev1.ResourceStorage): *size,
		}
	}

	return []corev1.PersistentVolumeClaim{claim}
}

func (d *Deployment) createPodTemplateSpec(nfsPort int, rpcPort int, metricsPort int) corev1.PodTemplateSpec {

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			// Name:      d.nfsServer.Name,
			// Namespace: d.nfsServer.Namespace,
			Labels: labelsForStatefulSet(d.nfsServer.Name),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					ImagePullPolicy: "IfNotPresent",
					Name:            "ganesha",
					// Name:            d.nfsServer.Name,
					Image: d.nfsServer.Spec.GetContainerImage(),
					// Args: []string{"nfs", "server", "--ganeshaConfigPath=" + NFSConfigMapPath + "/" + nfsServer.name},
					Env: []corev1.EnvVar{
						{
							Name:  "GANESHA_CONFIGFILE",
							Value: "/config/" + d.nfsServer.Name,
						},
					},
					Ports: []corev1.ContainerPort{
						{
							Name:          "nfs-port",
							ContainerPort: int32(nfsPort),
						},
						{
							Name:          "rpc-port",
							ContainerPort: int32(rpcPort),
						},
						{
							Name:          "metrics-port",
							ContainerPort: int32(metricsPort),
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "nfs-config",
							MountPath: "/config",
						},
						{
							Name:      "nfs-data",
							MountPath: "/export",
						},
					},
					SecurityContext: &corev1.SecurityContext{
						Capabilities: &corev1.Capabilities{
							Add: []corev1.Capability{
								"SYS_ADMIN",
								"DAC_READ_SEARCH",
							},
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "nfs-config",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: d.nfsServer.Name,
							},
						},
					},
				},
			},
		},
	}
}

func (d *Deployment) deleteStatefulSet(name string, namespace string) error {

	obj, err := d.getStatefulSet(name, namespace)
	if err != nil {
		return err
	}
	return d.deleteObject(obj)
}

func (d *Deployment) getStatefulSet(name string, namespace string) (*appsv1.StatefulSet, error) {

	instance := &appsv1.StatefulSet{}
	nn := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	if err := d.client.Get(context.TODO(), nn, instance); err != nil {
		return nil, err
	}

	return instance, nil
}

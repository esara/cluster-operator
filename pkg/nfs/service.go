package nfs

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (d *Deployment) ensureService(nfsPort int, rpcPort int, metricsPort int) (*corev1.Service, error) {

	svc, err := d.getService(d.nfsServer.Name, d.nfsServer.Namespace)
	if err == nil {
		return svc, err
	}

	if err := d.createService(nfsPort, rpcPort, metricsPort); err != nil {
		return nil, err
	}
	return d.getService(d.nfsServer.Name, d.nfsServer.Namespace)
}

func (d *Deployment) createService(nfsPort int, rpcPort int, metricsPort int) error {

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            d.nfsServer.Name,
			Namespace:       d.nfsServer.Namespace,
			Labels:          labelsForStatefulSet(d.nfsServer.Name),
			OwnerReferences: d.nfsServer.ObjectMeta.OwnerReferences,
		},
		Spec: corev1.ServiceSpec{
			Selector: labelsForStatefulSet(d.nfsServer.Name),
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "nfs",
					Port:       int32(nfsPort),
					TargetPort: intstr.FromInt(int(nfsPort)),
				},
				{
					Name:       "rpc",
					Port:       int32(rpcPort),
					TargetPort: intstr.FromInt(int(rpcPort)),
				},
				{
					Name:       "metrics",
					Port:       int32(metricsPort),
					TargetPort: intstr.FromInt(int(metricsPort)),
				},
			},
		},
	}

	return d.createOrUpdateObject(svc)
}

func (d *Deployment) getService(name string, namespace string) (*corev1.Service, error) {

	service := &corev1.Service{}

	namespacedService := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	if err := d.client.Get(context.TODO(), namespacedService, service); err != nil {
		return nil, err
	}
	return service, nil
}

func (d *Deployment) deleteService() error {

	svc, err := d.getService(d.nfsServer.Name, d.nfsServer.Namespace)
	if err != nil {
		return err
	}
	return d.deleteObject(svc)
}

package nfs

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (d *Deployment) ensureService(nfsPort int, rpcPort int) (*corev1.Service, error) {

	svc, err := d.getService()
	if err == nil {
		return svc, err
	}

	if err := d.createService(nfsPort, rpcPort); err != nil {
		return nil, err
	}
	return d.getService()
}

func (d *Deployment) createService(nfsPort int, rpcPort int) error {

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
			},
		},
	}

	return d.createOrUpdateObject(svc)
}

func (d *Deployment) getService() (*corev1.Service, error) {

	service := &corev1.Service{}

	namespacedService := types.NamespacedName{
		Namespace: d.nfsServer.Namespace,
		Name:      d.nfsServer.Name,
	}
	if err := d.client.Get(context.TODO(), namespacedService, service); err != nil {
		return nil, err
	}
	return service, nil
}

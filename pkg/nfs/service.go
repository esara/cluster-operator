package nfs

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (d *Deployment) createService(nfsPort int, rpcPort int) error {

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            d.nfsServer.Name,
			Namespace:       d.nfsServer.Namespace,
			Labels:          labelsForStatefulSet(d.nfsServer.Name),
			OwnerReferences: d.nfsServer.ObjectMeta.OwnerReferences,
		},
		Spec: corev1.ServiceSpec{
			Selector: labelsForStatefulSet(nfsServer),
			Type:     v1.ServiceTypeClusterIP,
			Ports:    []v1.ServicePort{
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


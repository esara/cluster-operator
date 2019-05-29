package nfsserver

import (
	"context"
	"log"
	"strings"

	storageosv1alpha1 "github.com/storageos/cluster-operator/pkg/apis/storageos/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// var log = logf.Log.WithName("nfsserver")

// AddController creates a new NFSServer Controller and adds it to the Manager.
// The Manager will set fields on the Controller and Start it when the Manager
// is Started.
func AddController(mgr manager.Manager) error {
	return addController(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNFSServer{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		recorder: mgr.GetRecorder("storageos-nfsserver"),
	}
}

// addController adds a new NFSServer Controller to mgr with r as the
// reconcile.Reconciler.
func addController(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("nfsserver-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource NFSServer
	err = c.Watch(&source.Kind{Type: &storageosv1alpha1.NFSServer{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner NFSServer
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &storageosv1alpha1.NFSServer{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileNFSServer{}

// ReconcileNFSServer reconciles a NFSServer object
type ReconcileNFSServer struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client        client.Client
	scheme        *runtime.Scheme
	recorder      record.EventRecorder
	currentServer *Server
}

// Reconcile reads that state of the cluster for a NFSServer object and makes
// changes based on the state read and what is in the NFSServer.Spec
func (r *ReconcileNFSServer) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	// reqLogger.Info("Reconciling NFSServer")
	log.Print("Reconciling NFSServer")

	// Fetch the NFSServer instance
	instance := &storageosv1alpha1.NFSServer{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile
			// request. Owned objects are automatically garbage collected.
			// Return and don't requeue.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// // Define a new St object
	// pod := newPodForCR(instance)

	// // Set NFSServer instance as the owner and controller
	// if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
	// 	log.Printf("Skip reconcile: failed to set controller reference for %s/%s: %s", instance.Namespace, instance.Name, err.Error())
	// 	return reconcile.Result{}, err
	// }

	// // Check if the NFSServer StatefulSet already exists
	// found := &appsv1.StatefulSet{}
	// err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	// if err != nil && errors.IsNotFound(err) {
	// 	// reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
	// 	err = r.client.Create(context.TODO(), pod)
	// 	if err != nil {
	// 		return reconcile.Result{}, err
	// 	}

	// 	// Pod created successfully - don't requeue
	// 	return reconcile.Result{}, nil
	// } else if err != nil {
	// 	return reconcile.Result{}, err
	// }

	// Pod already exists - don't requeue
	// reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)

	if err := r.reconcile(instance); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileNFSServer) reconcile(m *storageosv1alpha1.NFSServer) error {

	// Finalizers are set when an object should be deleted. Apply deploy only
	// when finalizers is empty.
	if len(m.GetFinalizers()) == 0 {

		if err := r.currentServer.Deploy(r); err != nil {
			// Ignore "Operation cannot be fulfilled" error. It happens when the
			// actual state of object is different from what is known to the operator.
			// Operator would resync and retry the failed operation on its own.
			if !strings.HasPrefix(err.Error(), "Operation cannot be fulfilled") {
				r.recorder.Event(m, corev1.EventTypeWarning, "FailedCreation", err.Error())
			}
			return err
		}
	} else {
		// Delete the deployment once the finalizers are set on the cluster
		// resource.
		r.recorder.Event(m, corev1.EventTypeNormal, "Terminating", "Deleting all the resources...")

		if err := r.currentServer.DeleteDeployment(); err != nil {
			return err
		}

		// r.ResetCurrentCluster()

		// Reset finalizers and let k8s delete the object.
		// When finalizers are set on an object, metadata.deletionTimestamp is
		// also set. deletionTimestamp helps the garbage collector identify
		// when to delete an object. k8s deletes the object only once the
		// list of finalizers is empty.
		m.SetFinalizers([]string{})
		return r.client.Update(context.Background(), m)
	}

	return nil

	return nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPVCForCR(cr *storageosv1alpha1.NFSServer) *corev1.PersistentVolumeClaim {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pvc",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{},
	}
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *storageosv1alpha1.NFSServer) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}

// newPVForCR returns a busybox pod with the same name/namespace as the cr
func newPVForCR(cr *storageosv1alpha1.NFSServer) *corev1.PersistentVolumeClaim {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pvc",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{},
	}
}

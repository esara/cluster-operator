package nfsserver

import (
	"context"
	"strings"

	storageosv1 "github.com/storageos/cluster-operator/pkg/apis/storageos/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/storageos/cluster-operator/pkg/nfs"
)

var log = ctrl.Log.WithName("nfsserver")

const finalizer = "finalizer.nfsserver.storageos.com"

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

	// Watch for PVC request for StorageOS shared volumes.
	// TODO: re-enable for non-CSI
	// err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForObject{})
	// if err != nil {
	// 	return err
	// }

	// Watch for changes to primary resource NFSServer.
	err = c.Watch(&source.Kind{Type: &storageosv1.NFSServer{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource StatefulSet and requeue the owner
	// NFSServer.
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &storageosv1.NFSServer{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Service and requeue the owner
	// NFSServer.
	//
	// This is used to update the NFSServer Status with the connection endpoint
	// once it comes online.
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &storageosv1.NFSServer{},
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
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a NFSServer object and makes
// changes based on the state read and what is in the NFSServer.Spec.
func (r *ReconcileNFSServer) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	log := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	// log.Info("Reconciling NFS Server")

	// Fetch the NFSServer instance
	instance := &storageosv1.NFSServer{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile
			// request. Owned objects are automatically garbage collected.
			// Return and don't requeue.
			log.Info("Creating NFS Server")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "failed to retrieve NFS Server, will retry")
		return reconcile.Result{}, err
	}

	// // Check if the Memcached instance is marked to be deleted, which is
	// // indicated by the deletion timestamp being set.
	// if instance.GetDeletionTimestamp() != nil {

	// 	if contains(instance.GetFinalizers(), finalizer) {
	// 		// Run finalization logic for memcachedFinalizer. If the
	// 		// finalization logic fails, don't remove the finalizer so
	// 		// that we can retry during the next reconciliation.
	// 		if err := r. d.Delete()(reqLogger, instance); err != nil {
	// 			return reconcile.Result{}, err
	// 		}

	// 		// Remove instanceFinalizer. Once all finalizers have been
	// 		// removed, the object will be deleted.
	// 		instance.SetFinalizers(remove(instance.GetFinalizers(), finalizer))
	// 		err := r.client.Update(context.TODO(), instance)
	// 		if err != nil {
	// 			return reconcile.Result{}, err
	// 		}
	// 	}
	// 	return reconcile.Result{}, nil
	// }

	// Set as the current cluster if there's no current cluster.
	// r.SetCurrentClusterIfNone(instance)

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

	// log.Printf("instance: %#v\n", instance)

	if err := r.reconcile(instance); err != nil {
		log.Error(err, "Reconcile failed")
		return reconcile.Result{}, err
	}

	log.Info("NFS Server reconcile done")

	return reconcile.Result{}, nil
}

func (r *ReconcileNFSServer) reconcile(instance *storageosv1.NFSServer) error {

	log := log.WithValues("Request.Namespace", instance.Namespace, "Request.Name", instance.Name)

	// Add our finalizer immediately so we can cleanup a partial deployment.  If
	// this is not set, the CR can simply be deleted.
	if len(instance.GetFinalizers()) == 0 {

		// Add our finalizer so that we control deletion.
		if err := r.addFinalizer(instance); err != nil {
			return err
		}

		// Return here, as the update to add the finalizer will trigger another
		// reconcile.
		return nil
	}

	d := nfs.NewDeployment(r.client, instance, r.recorder, r.scheme, log)

	// If the CR has not been marked for deletion, ensure it is deployed.
	if instance.GetDeletionTimestamp() == nil {
		if err := d.Deploy(); err != nil {
			// Ignore "Operation cannot be fulfilled" error. It happens when the
			// actual state of object is different from what is known to the operator.
			// Operator would resync and retry the failed operation on its own.
			if !strings.HasPrefix(err.Error(), "Operation cannot be fulfilled") {
				r.recorder.Event(instance, corev1.EventTypeWarning, "FailedCreation", err.Error())
			}
			return err
		}

	} else {

		log.Info("Removing the NFS server")

		// Delete the deployment once the finalizers are set on the cluster
		// resource.
		r.recorder.Event(instance, corev1.EventTypeNormal, "Terminating", "Deleting the NFS server.")

		if err := d.Delete(); err != nil {
			return err
		}

		// Reset finalizers and let k8s delete the object.
		// When finalizers are set on an object, metadata.deletionTimestamp is
		// also set. deletionTimestamp helps the garbage collector identify
		// when to delete an object. k8s deletes the object only once the
		// list of finalizers is empty.
		instance.SetFinalizers([]string{})
		return r.client.Update(context.Background(), instance)
	}

	return nil

}

func (r *ReconcileNFSServer) addFinalizer(instance *storageosv1.NFSServer) error {

	instance.SetFinalizers(append(instance.GetFinalizers(), finalizer))

	// Update CR
	err := r.client.Update(context.TODO(), instance)
	if err != nil {
		return err
	}
	return nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func remove(list []string, s string) []string {
	for i, v := range list {
		if v == s {
			list = append(list[:i], list[i+1:]...)
		}
	}
	return list
}

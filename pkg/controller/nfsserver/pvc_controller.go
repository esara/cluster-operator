package nfsserver

// The PVC controller watches for CRUD operations on PVCs, and if the PVC's
// provisioner matches nfs-storageos.com it will handle the operation.
//
// Operations are handled by creating, modifying or deleting an NFSServer
// Custom Resource (CR).  The NFSServer controller will handle the details.

import (
	"context"
	e "errors"
	"log"

	storageosv1alpha1 "github.com/storageos/cluster-operator/pkg/apis/storageos/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	// errNoProvisioner is returned when a PVC's provisioner couldn't be
	// determined.
	errNoProvisioner = e.New("pvc provisioner unknown")

	// errNoStorageClass is returned when a PVC's StorageClass couldn't be
	// determined.
	errNoStorageClass = e.New("pvc storageclass unknown")
)

// AddProvisioner creates a new PVC Controller and adds it to the Manager.
// The Manager will set fields on the Controller and Start it when the Manager
// is Started.
func AddProvisioner(mgr manager.Manager) error {
	log.Print("adding NFSServer provisioner")
	return addProvisioner(mgr, newProvisioner(mgr))
}

// newProvisioner returns a new reconcile.Reconciler that provisions new
// NFSServers for matching PVCs.
func newProvisioner(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePVC{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// addProvisioner adds a new PVC Controller to mgr with r as the
// reconcile.Reconciler.
func addProvisioner(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("nfsserver-provisioner", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource (PVC).
	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	log.Print("watching PVC events")

	// Watch for changes to secondary resource (NFSServer).
	err = c.Watch(&source.Kind{Type: &storageosv1alpha1.NFSServer{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &corev1.PersistentVolumeClaim{},
	})
	if err != nil {
		return err
	}

	log.Print("watching NFSServer events")

	return nil
}

var _ reconcile.Reconciler = &ReconcilePVC{}

// ReconcilePVC reconciles a PersistentVolumeClaim object.
type ReconcilePVC struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a PVC object and makes
// changes based on the state read and what is in the NFSServer.Spec
func (r *ReconcilePVC) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	// reqLogger.Info("Reconciling PVC")
	log.Print("Reconciling PVC")

	// Fetch the PVC instance
	instance := &corev1.PersistentVolumeClaim{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile
			// request. Owned objects are automatically garbage collected. For
			// additional cleanup logic use finalizers. Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Get PVC provisioner.
	provisioner, err := r.getProvisioner(instance)
	if err != nil {
		log.Printf("Skip reconcile: failed to get PVC provisioner for %s/%s: %s", instance.Namespace, instance.Name, err.Error())
		return reconcile.Result{}, nil
	}

	// Skip PVCs not using StorageOS provisioner.
	if provisioner != "nfs.storageos.com" {
		log.Printf("Skip reconcile: PVC not using StorageOS NFS provisioner %s/%s: %s", instance.Namespace, instance.Name, provisioner)
		// reqLogger.Info("Skip reconcile: PVC not using StorageOS provisioner", "PVC.Namespace", instance.Namespace, "PVC.Name", instance.Name, "PVC.Provisioner", provisioner)
		return reconcile.Result{}, nil
	}

	// Define a new NFSServer CR.
	nfs := newNFSServerForPVC(instance)

	// Set NFSServer instance as the owner and controller.
	if err := controllerutil.SetControllerReference(instance, nfs, r.scheme); err != nil {
		log.Printf("Skip reconcile: failed to set controller reference for %s/%s: %s", instance.Namespace, instance.Name, err.Error())
		return reconcile.Result{}, err
	}

	// Check if this NFS Server already exists.
	found := &storageosv1alpha1.NFSServer{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: nfs.Name, Namespace: nfs.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// reqLogger.Info("Creating a new NFS server", "NFSServer.Namespace", nfs.Namespace, "NFSServer.Name", nfs.Name)
		log.Printf("Creating a new NFS server for %s/%s", instance.Namespace, instance.Name)
		err = r.client.Create(context.TODO(), nfs)
		if err != nil {
			log.Printf("Failed to create a new NFS server for %s/%s: %s", instance.Namespace, instance.Name, err.Error())
			return reconcile.Result{}, err
		}

		// NFS Server created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		log.Printf("Failed to check is NFS server exists for %s/%s: %s", instance.Namespace, instance.Name, err.Error())
		return reconcile.Result{}, err
	}

	// NFS Server already exists - don't requeue
	// reqLogger.Info("Skip reconcile: NFS Server already exists", "NFSServer.Namespace", found.Namespace, "NFSServer.Name", found.Name)
	log.Printf("NFS server exists for %s/%s", instance.Namespace, instance.Name)
	return reconcile.Result{}, nil
}

// newNFSServerForPVC returns an NFSServer CR with the same name/namespace as
// the PVC.
func newNFSServerForPVC(pvc *corev1.PersistentVolumeClaim) *storageosv1alpha1.NFSServer {

	// TODO(simon): inherit pvc labels?
	labels := map[string]string{
		"app": pvc.Name,
	}
	return &storageosv1alpha1.NFSServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvc.Name,
			Namespace: pvc.Namespace,
			Labels:    labels,
		},
		Spec: storageosv1alpha1.NFSServerSpec{
			Annotations: pvc.Annotations,
			Exports: []storageosv1alpha1.ExportsSpec{
				{
					Name: pvc.Name,
					// AccessMode: pvc.AccessMode,
					// Squash
					// AllowedClients
				},
			},
		},
	}
}

// getProvisioner returns the PVC's provisioner.
func (r *ReconcilePVC) getProvisioner(pvc *corev1.PersistentVolumeClaim) (string, error) {

	class, err := r.getStorageClass(pvc)
	if err != nil {
		return "", err
	}

	return class.Provisioner, nil
}

// getStorageClass returns the PVC's StorageClass.
func (r *ReconcilePVC) getStorageClass(pvc *corev1.PersistentVolumeClaim) (*storagev1.StorageClass, error) {

	className := r.getStorageClassName(pvc)
	if className == "" {
		return nil, errNoStorageClass
	}

	class := &storagev1.StorageClass{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: className}, class)
	if err != nil {
		return nil, err
	}
	return class, nil
}

// getStorageClassName returns the name of the StorageClass responsible for
// provisioning the PVC.
func (r *ReconcilePVC) getStorageClassName(pvc *corev1.PersistentVolumeClaim) string {

	// Supported from k8s 1.6 onwards.
	if pvc.Spec.StorageClassName != nil {
		return *pvc.Spec.StorageClassName
	}

	// Deprecated in k8s 1.8 but only removed in 1.11.
	if name, ok := pvc.Annotations[corev1.BetaStorageClassAnnotation]; ok {
		return name
	}

	return ""
}

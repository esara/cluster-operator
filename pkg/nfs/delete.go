package nfs

import "log"

// Delete deletes all the storageos resources.
// This explicit delete is implemented instead of depending on the garbage
// collector because sometimes the garbage collector deletes the resources
// with owner reference as a CRD without the parent being deleted. This happens
// especially when a cluster reboots. Althrough the operator re-creates the
// resources, we want to avoid this behavior by implementing an explcit delete.
func (d *Deployment) Delete() error {

	log.Printf("deleting resources for nfs server: %s", d.nfsServer.Name)

	if err := d.deleteStatefulSet(d.nfsServer.Name, d.nfsServer.Namespace); err != nil {
		log.Printf("failed to delete statefulset: %v", err)
	}
	if err := d.deleteService(); err != nil {
		log.Printf("failed to delete statefulset: %v", err)
	}

	return nil

}

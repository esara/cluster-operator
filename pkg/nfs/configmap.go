package nfs

import (
	"fmt"
	"log"
	"strings"

	storageosv1 "github.com/storageos/cluster-operator/pkg/apis/storageos/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createConfig(instance *storageosv1.NFSServer) string {

	// id needs to be unique for each export on the server node.
	id := 57

	log.Printf("Spec.Exports: %#v", instance.Spec.Exports)

	var exports []string
	// If no export list given, use defaults
	if len(instance.Spec.Exports) == 0 {
		log.Printf("configuring default export for: %s", instance.Name)
		exports = append(exports, exportConfig(id, instance.Name, "readwrite", "none"))
	}

	// Otherwise use export list
	for _, export := range instance.Spec.Exports {
		log.Printf("configuring export for: %s", export.PersistentVolumeClaim.ClaimName)
		exports = append(exports, exportConfig(id, export.PersistentVolumeClaim.ClaimName, export.Server.AccessMode, export.Server.Squash))
		id++
	}

	return globalConfig() + logConfig() + strings.Join(exports, "\n")
}

func exportConfig(id int, ref string, access string, squash string) string {
	return `
EXPORT {
	Export_Id = ` + fmt.Sprintf("%v", id) + `;
	Path = /export/` + ref + `;
	Pseudo = /` + ref + `;
	Protocols = 4;
	Transports = TCP;
	Sectype = sys;
	Access_Type = ` + getAccessMode(access) + `;
	Squash = ` + getSquash(squash) + `;
	FSAL {
		Name = VFS;
	}
}
`
}

func globalConfig() string {
	// 	Dbus_Name_Prefix = storageos;
	return `
NFS_Core_Param {
	fsid_device = true;
}`
}

// TODO, use defualt "EVENT" level.
func logConfig() string {
	return `
LOG {
	default_log_level = DEBUG;
	Components {
		ALL = DEBUG;
	}
}`
}

func getAccessMode(mode string) string {
	switch strings.ToLower(mode) {
	case "none":
		return "None"
	case "readonly":
		return "RO"
	default:
		return "RW"
	}
}

func getSquash(squash string) string {
	if squash != "" {
		return strings.ToLower(squash)
	}
	return "none"
}

func (d *Deployment) createNFSConfigMap() error {

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            d.nfsServer.Name,
			Namespace:       d.nfsServer.Namespace,
			OwnerReferences: d.nfsServer.ObjectMeta.OwnerReferences,
			// Labels:          createAppLabels(nfsServer),
		},
		Data: map[string]string{
			d.nfsServer.Name: createConfig(d.nfsServer),
		},
	}

	return d.createOrUpdateObject(configMap)
}

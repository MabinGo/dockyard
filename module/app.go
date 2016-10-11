package module

import (
	"sync"

	"github.com/containerops/dockyard/models"
)

var AppFileLock sync.RWMutex

func GetOriginBlobsum(id int64, system, arch, appname, tag string) (blobSum string) {
	o := new(models.ArtifactV1)
	o.AppV1, o.OS, o.Arch, o.App, o.Tag = id, system, arch, appname, tag
	if exists, err := o.Read(); err != nil {
		blobSum = ""
	} else if !exists {
		blobSum = ""
	} else {
		blobSum = o.BlobSum
	}
	return
}

package synchron

import (
	"fmt"
	//"log"
	//"net/http"
	"time"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/utils/setting"
)

func InitSynchron() error {
	go func() {
		timer := time.NewTicker(time.Duration(setting.Interval) * time.Second)
		for {
			select {
			case <-timer.C:
				//create goroutine to distributed images at set intervals
				for _, region := range models.Regions {
					if !region.Active {
						continue
					}

					if err := module.TrigSyn(region.Namespace, region.Repository, region.Tag, region.Dest); err != nil {
						fmt.Printf("Syn %s/%s/%s error: %s", region.Namespace, region.Repository, region.Tag, err.Error())
						continue
					}
					//TODO: 考虑怎么存region
					region.Active = false
				}
			}
		}
	}()

	return nil
}

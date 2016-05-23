package synch

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
				for _, r := range models.Regions { //TODO: 不能用全局数组
					if err := module.TrigSynEndpoint(&r); err != nil {
						fmt.Printf("\nSynchronize %s/%s/%s error: %s\n", r.Namespace, r.Repository, r.Tag, err.Error())
					}
				}
			}
		}
	}()

	return nil
}

package synchron

import (
	"fmt"
	//"log"
	//"net/http"
	"time"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/utils/setting"
)

var RegionTabs = []models.Region{}
var SynTabs = []models.Syn{}

func InitSynchron() error {
	//TODO:
	go func() {
		timer := time.NewTicker(time.Duration(setting.Interval) * time.Second)
		for {
			select {
			case <-timer.C:
				fmt.Println("####### InitSynchron 0")
				//create goroutine to distributed images at set intervals
				//...
			}
		}
	}()

	return nil
}

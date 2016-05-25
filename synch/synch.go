package synch

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/setting"
)

func InitSynchron() error {
	rt := new(models.RegionTable)
	if existes, err := rt.Get(module.RTName); err != nil {
		return err
	} else if !existes {
		if err := rt.Save(module.RTName); err != nil {
			return err
		}
	}

	authorization := "Basic " + utils.EncodeBasicAuth(setting.SynUser, setting.SynPasswd)

	go func() {
		timer := time.NewTicker(time.Duration(setting.SynInterval) * time.Second)
		for {
			select {
			case <-timer.C:
				rt := new(models.RegionTable)
				if exists, err := rt.Get(module.RTName); err != nil {
					fmt.Printf("\nDB invalid: %s\n", err.Error())
					continue
				} else if exists {
					if rt.Regionlist == "" {
						continue
					}

					rlist := new(models.Regionlist)
					if err := json.Unmarshal([]byte(rt.Regionlist), rlist); err != nil {
						fmt.Printf("\nRegion list invalid: %s\n", err.Error())
						continue
					}

					//create goroutine to distributed images at set intervals
					for _, r := range rlist.Regions {
						if err := module.TrigSynEndpoint(&r, authorization); err != nil {
							fmt.Printf("\nSynchronize %s/%s/%s error: %s", r.Namespace, r.Repository, r.Tag, err.Error())
						}
						//else {
						//	fmt.Printf("\nSynchronize %s/%s/%s successfully", r.Namespace, r.Repository, r.Tag)
						//}
					}
				} else {

				}
			}
		}
	}()

	return nil
}

package synch

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/astaxie/beego/logs"

	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/setting"
)

var Log *logs.BeeLogger

func InitSynchron() error {
	rt := new(RegionTable)
	if existes, err := rt.Get(RTName); err != nil {
		return err
	} else if !existes {
		if err := rt.Save(RTName); err != nil {
			return err
		}
	}

	Log = logs.NewLogger(4096)
	Log.SetLogger("console", "")
	Log.SetLogger("file", fmt.Sprintf("{\"filename\":\"%s\"}", setting.LogPath))

	if setting.SynMode != "" {
		authorization := "Basic " + utils.EncodeBasicAuth(setting.SynUser, setting.SynPasswd)
		go func() {
			timer := time.NewTicker(time.Duration(setting.SynInterval) * time.Second)
			for {
				select {
				case <-timer.C:
					rt := new(RegionTable)
					if exists, err := rt.Get(RTName); err != nil {
						fmt.Printf("\nDB invalid: %s\n", err.Error())
						continue
					} else if exists {
						if rt.Regionlist == "" {
							continue
						}

						rlist := new(Regionlist)
						if err := json.Unmarshal([]byte(rt.Regionlist), rlist); err != nil {
							fmt.Printf("\nRegion list invalid: %s\n", err.Error())
							continue
						}

						//create goroutine to distributed images at set intervals
						for _, r := range rlist.Regions {
							if err := TrigSynEndpoint(&r, authorization); err != nil {
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
	}

	return nil
}

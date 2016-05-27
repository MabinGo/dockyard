package synch

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/astaxie/beego/logs"

	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/setting"
)

var synlog *logs.BeeLogger

func InitSynchron() error {
	rt := new(RegionTable)
	if existes, err := rt.Get(RTName); err != nil {
		return err
	} else if !existes {
		if err := rt.Save(RTName); err != nil {
			return err
		}
	}

	synlog = logs.NewLogger(4096)
	synlog.SetLogger("console", "")
	synlog.SetLogger("file", fmt.Sprintf("{\"filename\":\"%s\"}", setting.LogPath))

	if setting.SynMode != "" {
		authorization := "Basic " + utils.EncodeBasicAuth(setting.SynUser, setting.SynPasswd)
		go func() {
			timer := time.NewTicker(time.Duration(setting.SynInterval) * time.Second)
			for {
				select {
				case <-timer.C:
					rt := new(RegionTable)
					if exists, err := rt.Get(RTName); err != nil {
						synlog.Error("\nDB invalid: %s\n", err.Error())
						continue
					} else if exists {
						if rt.Regionlist == "" {
							continue
						}

						rlist := new(Regionlist)
						if err := json.Unmarshal([]byte(rt.Regionlist), rlist); err != nil {
							synlog.Error("\nRegion list invalid: %s\n", err.Error())
							continue
						}

						//create goroutine to distributed images at set intervals
						for _, r := range rlist.Regions {
							//TODO: repeat while timeout
							TrigSynEndpoint(&r, authorization)
						}
					} else {

					}
				}
			}
		}()
	}

	return nil
}

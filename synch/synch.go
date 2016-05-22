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
						fmt.Printf("Synchronize %s/%s/%s error: %s", r.Namespace, r.Repository, r.Tag, err.Error())
					}
					/*
						epg := new(models.Endpointgrp)
						if err := json.Unmarshal([]byte(r.Endpointlist), epg); err != nil {
							fmt.Printf("Synchronize %s/%s/%s error: %s", r.Namespace, r.Repository, r.Tag, err.Error())
							continue
						}

						for k, _ := range epg.Endpoints {
							if epg.Endpoints[k].Active == false {
								continue
							}

							if err := module.TrigSynch(r.Namespace, r.Repository, r.Tag, epg.Endpoints[k].URL); err != nil {
								fmt.Printf("Synchronize %s/%s/%s error: %s", r.Namespace, r.Repository, r.Tag, err.Error())
								continue
							} else {
								epg.Endpoints[k].Active = false
							}
						}
						result, _ := json.Marshal(epg)
						r.Endpointlist = string(result)
						if err := r.Save(namespace, repository, tag); err != nil {
							fmt.Printf("Synchronize %s/%s/%s status error: %s", r.Namespace, r.Repository, r.Tag, err.Error())
							continue
						}
					*/
				}
			}
		}
	}()

	return nil
}

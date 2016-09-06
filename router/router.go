/*
Copyright 2015 The ContainerOps Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package router

import (
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/handler"
)

//SetRouters is setting REST API interface with handler function.
func SetRouters(m *macaron.Macaron) {
	m.Options("*", func(ctx *macaron.Context) (int, []byte) {
		// TODO Handle the request by CORS configs
		return 200, nil
	})

	// Docker Registry V2
	m.Group("/v2", func() {
		m.Get("/", handler.GetPingV2Handler)
		m.Get("/_catalog", handler.GetCatalogV2Handler)

		// user mode: /namespace/repository:tag
		m.Head("/:namespace/:repository/blobs/:digest", handler.HeadBlobsV2Handler)
		m.Post("/:namespace/:repository/blobs/uploads", handler.PostBlobsV2Handler)
		m.Patch("/:namespace/:repository/blobs/uploads/:uuid", handler.PatchBlobsV2Handler)
		m.Put("/:namespace/:repository/blobs/uploads/:uuid", handler.PutBlobsV2Handler)
		m.Get("/:namespace/:repository/blobs/:digest", handler.GetBlobsV2Handler)
		m.Put("/:namespace/:repository/manifests/:tag", handler.PutManifestsV2Handler)
		m.Get("/:namespace/:repository/tags/list", handler.GetTagsListV2Handler)
		m.Get("/:namespace/:repository/manifests/:tag", handler.GetManifestsV2Handler)
		m.Delete("/:namespace/:repository/blobs/:digest", handler.DeleteBlobsV2Handler)
		m.Delete("/:namespace/:repository/manifests/:reference", handler.DeleteManifestsV2Handler)
		m.Get("/:namespace/:repository/meta", handler.GetMetaV2Handler)
		m.Get("/:namespace/:repository/metasign", handler.GetMetaSignV2Handler)
	})

	// App Discovery
	m.Group("/app", func() {
		m.Group("/v1", func() {
			// Global Search
			m.Get("/search", handler.AppGlobalSearchV1Handler)

			m.Group("/:namespace/:repository", func() {
				// Discovery
				//m.Get("/?app-discovery=1", handler.AppDiscoveryV1Handler)

				// Scoped Search
				m.Get("/search", handler.AppScopedSearchV1Handler)
				m.Get("/list", handler.AppGetListAppV1Handler)

				// Pull
				m.Get("/:os/:arch/:app/?:tag", handler.AppGetFileV1Handler)
				m.Get("/:os/:arch/:app/manifests/?:tag", handler.AppGetManifestsV1Handler)
				m.Get("/meta", handler.AppGetMetaV1Handler)
				m.Get("/metasign", handler.AppGetMetaSignV1Handler)

				// Push
				m.Post("/", handler.AppPostV1Handler)
				m.Put("/:os/:arch/:app/?:tag", handler.AppPutFileV1Handler)
				m.Put("/:os/:arch/:app/manifests/?:tag", handler.AppPutManifestV1Handler)
				m.Patch("/:os/:arch/:app/:status/?:tag", handler.AppPatchFileV1Handler)
				m.Delete("/:os/:arch/:app/?:tag", handler.AppDeleteFileV1Handler)
			})
		})
	})

	// Public Key APIs
	m.Group("/key", func() {
		m.Group("/v1", func() {
			m.Get("/", handler.KeyGetPublicV1Handler)
		})
	})

}

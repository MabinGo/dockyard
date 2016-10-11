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
	// Auth Server
	m.Group("/uam", func() {
		//Authorization service
		m.Get("/auth", handler.GetAuthorizeHander)
	})

	//REST API For Web Operations
	m.Group("/web", func() {
		m.Get("/:namespace", handler.AppAuthHandler, handler.GetNamespacePageV1Handler)
		m.Get("/:type/:namespace/:repository", handler.AppAuthHandler, handler.GetRepositoryPageV1Handler)
		m.Post("/utils/secret", handler.DockerConfJson)

		m.Group("/v1", func() {
			m.Get("/:type/:namespace/:repository/:os/:arch/:app/:tag/file/:filename", handler.AppAuthHandler, handler.AppGetFileV1Handler)

			m.Post("/:type/:namespace/:repository", handler.AppAuthHandler, handler.ValidateManifestsLength, handler.PostRepositoryRESTV1Handler)
			m.Get("/:type/:namespace/:repository", handler.AppAuthHandler, handler.GetRepositoryRESTV1Handler)
			m.Put("/:type/:namespace/:repository", handler.AppAuthHandler, handler.ValidateManifestsLength, handler.PutRepositoryRESTV1Handler)
			m.Delete("/:type/:namespace/:repository", handler.AppAuthHandler, handler.DeleteRepositoryRESTV1Handler)

			m.Post("/:type/:namespace/:repository/:package/auth", handler.AppAuthHandler, handler.CreatePackagePreAuthHandler)
			m.Post("/:type/:namespace/:repository/:package", handler.AppAuthHandler, handler.ValidateManifestsLength, handler.PostPackageRESTV1Handler)
			m.Get("/:type/:namespace/:repository/:package", handler.AppAuthHandler, handler.GetPackageRESTV1Hanfdler)
			m.Put("/:type/:namespace/:repository/:package", handler.AppAuthHandler, handler.ValidateManifestsLength, handler.PutPackageRESTV1Handler)
			m.Delete("/:type/:namespace/:repository/:package", handler.AppAuthHandler, handler.DeletePackageRESTV1Handler)

			m.Post("/:type/:namespace/:repository/:package/manifests", handler.AppAuthHandler, handler.ValidateManifestsLength, handler.PostManifestRESTV1Handler)
			m.Get("/:type/:namespace/:repository/:package/manifests", handler.AppAuthHandler, handler.GetManifestRESTV1Handler)
			m.Put("/:type/:namespace/:repository/:package/manifests", handler.AppAuthHandler, handler.ValidateManifestsLength, handler.PutManifestRESTV1Handler)
			m.Delete("/:type/:namespace/:repository/:package/manifests", handler.AppAuthHandler, handler.DeleteManifestRESTV1Handler)

		})
	})

	//Docker Registry V1
	m.Group("/v1", func() {
		m.Get("/_ping", handler.TokenValidateHandler, handler.GetPingV1Handler)

		m.Get("/users", handler.TokenValidateHandler, handler.GetUsersV1Handler)
		m.Post("/users", handler.TokenValidateHandler, handler.PostUsersV1Handler)

		m.Group("/repositories", func() {
			m.Put("/:namespace/:repository/tags/:tag", handler.TokenValidateHandler, handler.PutTagV1Handler)
			m.Put("/:namespace/:repository/images", handler.TokenValidateHandler, handler.PutRepositoryImagesV1Handler)
			m.Get("/:namespace/:repository/images", handler.TokenValidateHandler, handler.GetRepositoryImagesV1Handler)
			m.Get("/:namespace/:repository/tags", handler.TokenValidateHandler, handler.GetTagV1Handler)
			m.Put("/:namespace/:repository", handler.TokenValidateHandler, handler.PutRepositoryV1Handler)
		})

		m.Group("/images", func() {
			m.Get("/:image/ancestry", handler.TokenValidateHandler, handler.GetImageAncestryV1Handler)
			m.Get("/:image/json", handler.TokenValidateHandler, handler.GetImageJSONV1Handler)
			m.Get("/:image/layer", handler.TokenValidateHandler, handler.GetImageLayerV1Handler)
			m.Put("/:image/json", handler.TokenValidateHandler, handler.PutImageJSONV1Handler)
			m.Put("/:image/layer", handler.TokenValidateHandler, handler.PutImageLayerV1Handler)
			m.Put("/:image/checksum", handler.TokenValidateHandler, handler.PutImageChecksumV1Handler)
		})
	})

	// Docker Registry V2
	m.Group("/v2", func() {
		m.Get("/", handler.PingAuthHandler, handler.GetPingV2Handler)
		m.Get("/_catalog", handler.GetCatalogV2Handler)

		// user mode: /namespace/repository:tag
		m.Head("/:namespace/:repository/blobs/:digest", handler.TokenValidateHandler, handler.HeadBlobsV2Handler)
		m.Post("/:namespace/:repository/blobs/uploads", handler.TokenValidateHandler, handler.PostBlobsV2Handler)
		m.Patch("/:namespace/:repository/blobs/uploads/:uuid", handler.TokenValidateHandler, handler.PatchBlobsV2Handler)
		m.Put("/:namespace/:repository/blobs/uploads/:uuid", handler.TokenValidateHandler, handler.PutBlobsV2Handler)
		m.Get("/:namespace/:repository/blobs/:digest", handler.TokenValidateHandler, handler.GetBlobsV2Handler)
		m.Put("/:namespace/:repository/manifests/:tag", handler.TokenValidateHandler, handler.PutManifestsV2Handler)
		m.Get("/:namespace/:repository/tags/list", handler.TokenValidateHandler, handler.GetTagsListV2Handler)
		m.Get("/:namespace/:repository/manifests/:tag", handler.TokenValidateHandler, handler.GetManifestsV2Handler)
		m.Delete("/:namespace/:repository/blobs/:digest", handler.TokenValidateHandler, handler.DeleteBlobsV2Handler)
		m.Delete("/:namespace/:repository/manifests/:reference", handler.TokenValidateHandler, handler.DeleteManifestsV2Handler)
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
				m.Get("/search", handler.AppAuthHandler, handler.AppScopedSearchV1Handler)
				m.Get("/list", handler.AppAuthHandler, handler.AppGetListAppV1Handler)

				// Pull
				m.Get("/:os/:arch/:app/?:tag", handler.AppAuthHandler, handler.AppGetFileV1Handler)
				m.Get("/:os/:arch/:app/manifests/?:tag", handler.AppAuthHandler, handler.AppGetManifestsV1Handler)
				m.Get("/meta", handler.AppAuthHandler, handler.AppGetMetaV1Handler)
				m.Get("/metasign", handler.AppAuthHandler, handler.AppGetMetaSignV1Handler)

				// Push
				m.Post("/", handler.AppAuthHandler, handler.AppPostV1Handler)
				m.Put("/:os/:arch/:app/?:tag", handler.AppAuthHandler, handler.AppPutFileV1Handler)
				m.Put("/:os/:arch/:app/manifests/?:tag", handler.AppAuthHandler, handler.AppPutManifestV1Handler)
				m.Patch("/:os/:arch/:app/:status/?:tag", handler.AppAuthHandler, handler.AppPatchFileV1Handler)
				m.Delete("/:os/:arch/:app/?:tag", handler.AppAuthHandler, handler.AppDeleteFileV1Handler)
				m.Delete("/recycle", handler.AppAuthHandler, handler.AppRecycleV1Handler)
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

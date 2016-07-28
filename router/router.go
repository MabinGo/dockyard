package router

import (
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/auth/controller"
	"github.com/containerops/dockyard/handler"
	"github.com/containerops/dockyard/oss"
	"github.com/containerops/dockyard/oss/apiserver"
	"github.com/containerops/dockyard/synch"
)

func SetRouters(m *macaron.Macaron) {
	//Docker Registry & Hub V1 API
	m.Group("/v1", func() {
		m.Get("/_ping", handler.GetPingV1Handler)

		m.Get("/users", handler.GetUsersV1Handler)
		m.Post("/users", handler.PostUsersV1Handler)

		m.Group("/repositories", func() {
			m.Put("/:namespace/:repository/tags/:tag", handler.PutTagV1Handler)
			m.Put("/:namespace/:repository/images", handler.PutRepositoryImagesV1Handler)
			m.Get("/:namespace/:repository/images", handler.GetRepositoryImagesV1Handler)
			m.Get("/:namespace/:repository/tags", handler.GetTagV1Handler)
			m.Put("/:namespace/:repository", handler.PutRepositoryV1Handler)
		})

		m.Group("/images", func() {
			m.Get("/:imageId/ancestry", handler.GetImageAncestryV1Handler)
			m.Get("/:imageId/json", handler.GetImageJSONV1Handler)
			m.Get("/:imageId/layer", handler.GetImageLayerV1Handler)
			m.Put("/:imageId/json", handler.PutImageJSONV1Handler)
			m.Put("/:imageId/layer", handler.PutImageLayerv1Handler)
			m.Put("/:imageId/checksum", handler.PutImageChecksumV1Handler)
		})
	})

	//Docker Registry & Hub V2 API
	m.Group("/v2", func() {
		m.Get("/", handler.GetPingV2Handler)
		m.Get("/_catalog", handler.GetCatalogV2Handler)

		//user mode: /namespace/repository:tag
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

		//library mode: /repository:tag
		m.Head("/:repository/blobs/:digest", handler.HeadBlobsV2Handler)
		m.Post("/:repository/blobs/uploads", handler.PostBlobsV2Handler)
		m.Patch("/:repository/blobs/uploads/:uuid", handler.PatchBlobsV2Handler)
		m.Put("/:repository/blobs/uploads/:uuid", handler.PutBlobsV2Handler)
		m.Get("/:repository/blobs/:digest", handler.GetBlobsV2Handler)
		m.Put("/:repository/manifests/:tag", handler.PutManifestsV2Handler)
		m.Get("/:repository/tags/list", handler.GetTagsListV2Handler)
		m.Get("/:repository/manifests/:tag", handler.GetManifestsV2Handler)
		m.Delete("/:repository/blobs/:digest", handler.DeleteBlobsV2Handler)
		m.Delete("/:repository/manifests/:reference", handler.DeleteManifestsV2Handler)
	})

	//Rkt Registry & Hub API
	m.Get("/:namespace/:repository/?ac-discovery=1", handler.DiscoveryACIHandler)
	m.Group("/ac", func() {
		m.Group("/fetch", func() {
			m.Get("/:namespace/:repository/pubkeys", handler.GetPubkeysHandler)
			m.Get("/:namespace/:repository/:acifilename", handler.GetACIHandler)
		})

		m.Group("/push", func() {
			m.Post("/:namespace/:repository/uploaded/:acifile", handler.PostUploadHandler)
			m.Put("/:namespace/:repository/:imageId/manifest", handler.PutManifestHandler)
			m.Put("/:namespace/:repository/:imageId/signature/:signfile", handler.PutSignHandler)
			m.Put("/:namespace/:repository/:imageId/aci/:acifile", handler.PutAciHandler)
			m.Post("/:namespace/:repository/:imageId/complete/:acifile/:signfile", handler.PostCompleteHandler)
		})
	})

	//Object storage service API
	m.Group("/oss", func() {
		m.Post("/chunkserver", oss.StartLocalServer)
		m.Put("/chunkserver/info", oss.ReceiveChunkserverInfo)
		m.Group("/api", func() {
			m.Get("/file/info", apiserver.GetFileInfo)
			m.Get("/file", apiserver.DownloadFile)
			m.Post("/file", apiserver.UploadFile)
			m.Delete("/file", apiserver.DeleteFile)
		})
	})

	m.Group("/uam", func() {
		//Authorization service
		m.Get("/auth", controller.GetAuthorize)
		m.Delete("/auth", controller.DeleteAuthorize)
		m.Post("/auth", controller.PostAuthorize)

		//user
		m.Group("/user", func() {
			m.Post("/signup", controller.SignUp)
			m.Get("/signin", controller.SignIn)
			m.Put("/update", controller.UpdateUser)
			m.Post("/add", controller.CreateUser)
			m.Get("/list", controller.GetUserList)
			m.Delete("/:user", controller.DeleteUser)
			m.Delete("/deactive/:user", controller.DeactiveUser)
			m.Put("/active/:user", controller.ActiveUser)
		})

		//repository
		m.Group("/repository", func() {
			m.Post("/", controller.CreateRepository)
			m.Delete("/:namespace/:repository", controller.DeleteRepository)
			m.Delete("/deactive/:namespace/:repository", controller.DeactiveRepository)
			m.Put("/active/:namespace/:repository", controller.ActiveRepository)
			m.Put("/update", controller.UpdateRepository)
			m.Get("/:namespace/list", controller.GetRepositoryList)
			m.Get("/:repository/fuzzylist", controller.GetFuzzyRepositoryList)
		})

		//organization
		m.Group("/organization", func() {
			m.Post("/", controller.CreateOrganization)
			m.Delete("/:organization", controller.DeleteOrganization)
			m.Delete("/deactive/:organization", controller.DeactiveOrganization)
			m.Put("/active/:organization", controller.ActiveOrganization)
			m.Post("/adduser/", controller.AddUserToOrganization)
			m.Delete("/removeuser/:organization/:user", controller.RemoveUserFromOrganization)
			m.Delete("/deactiveuser/:organization/:user", controller.DeactiveUserFromOrganization)
			m.Put("/activeuser/:organization/:user", controller.ActiveUserFromOrganization)
			m.Put("/update", controller.UpdateOrganization)
			m.Put("/updateorganizationusermap", controller.UpdateOrganizationUserMap)
			m.Get("/list", controller.GetOrganizationList)
			m.Get("/:organization/listuser", controller.GetUserListFromOrganization)
		})

		//team
		m.Group("/team", func() {
			m.Post("/", controller.CreateTeam)
			m.Delete("/:organization/:team", controller.DeleteTeam)
			m.Delete("/deactive/:organization/:team", controller.DeactiveTeam)
			m.Put("/active/:organization/:team", controller.ActiveTeam)
			m.Post("/adduser/", controller.AddUserToTeam)
			m.Delete("/removeuser/:organization/:team/:user", controller.RemoveUserFromTeam)
			m.Delete("/deactiveuser/:organization/:team/:user", controller.DeactiveUserFromTeam)
			m.Put("/activeuser/:organization/:team/:user", controller.ActiveUserFromTeam)
			m.Post("/addrepository/", controller.AddRepositoryToTeam)
			m.Delete("/removerepository/:organization/:team/:repository", controller.RemoveRepositoryFromTeam)
			m.Delete("/deactiverepository/:organization/:team/:repository", controller.DeactiveRepositoryFromTeam)
			m.Put("/activerepository/:organization/:team/:repository", controller.ActiveRepositoryFromTeam)
			m.Put("/update", controller.UpdateTeam)
			m.Put("/updateteamusermap", controller.UpdateTeamUserMap)
			m.Put("/updateteamrepositorymap", controller.UpdateTeamRepositoryMap)
			m.Get("/:organization/listteam", controller.GetTeamListFromOrganization)
			m.Get("/:organization/:team/listuser", controller.GetUserListFromTeam)
			m.Get("/:organization/:team/listrepository", controller.GetRepositoryListFromTeam)
		})
	})

	//images synchron distributed
	m.Group("/syn", func() {
		m.Post("/:namespace/:repository/:tag/region", synch.PostSynRegionHandler)
		m.Post("/:namespace/:repository/:tag/trigger", synch.PostSynTrigHandler)
		m.Put("/:namespace/:repository/:tag/content", synch.PutSynContentHandler)
		m.Get("/:namespace/:repository/:tag/region", synch.GetSynRegionHandler)
		m.Delete("/:namespace/:repository/:tag/region", synch.DelSynRegionHandler)
		m.Post("/master", synch.PostSynMasterHandler)
		m.Get("/master", synch.GetSynMasterHandler)
		m.Delete("/master", synch.DelSynMasterHandler)
		m.Post("/drc", synch.PostSynDRCHandler)
		m.Get("/drc", synch.GetSynDRCHandler)
		m.Delete("/drc", synch.DelSynDRCHandler)
	})
}

package middleware

import (
	"net/http"
	"strings"

	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/validate"
)

func dockerValidate() macaron.Handler {
	return func(ctx *macaron.Context) {
		if !strings.Contains(ctx.Req.RequestURI, "/v2") ||
			utils.Compare(ctx.Req.RequestURI, "/v2/") == 0 ||
			utils.Compare(ctx.Req.RequestURI, "/v2/_catalog") == 0 {
			return
		}

		namespace := ctx.Params(":namespace")
		repository := ctx.Params(":repository")
		if !validate.IsNameValid(namespace) || !validate.IsNameValid(repository) {
			ctx.Resp.WriteHeader(http.StatusBadRequest)
			result, _ := module.ReportError(module.NAME_INVALID, "Invalid namespace or repository format")
			ctx.Resp.Write(result)
			return
		}

		if strings.Contains(ctx.Req.RequestURI, "/manifests/") &&
			(ctx.Req.Method == "PUT" || ctx.Req.Method == "GET") {
			tag := ctx.Params(":tag")
			if !validate.IsTagValid(tag) {
				ctx.Resp.WriteHeader(http.StatusBadRequest)
				result, _ := module.ReportError(module.TAG_INVALID, "Invalid tag format")
				ctx.Resp.Write(result)
				return
			}
		}

		if strings.Contains(ctx.Req.RequestURI, "/blobs/") &&
			(ctx.Req.Method == "HEAD" || ctx.Req.Method == "GET" || ctx.Req.Method == "DELETE") {
			digest := ctx.Params(":digest")
			if !validate.IsDigestValid(digest) {
				ctx.Resp.WriteHeader(http.StatusBadRequest)
				result, _ := module.ReportError(module.DIGEST_INVALID, "Invalid digest format")
				ctx.Resp.Write(result)
				return
			}
		}

		if strings.Contains(ctx.Req.RequestURI, "/manifests/") && ctx.Req.Method == "DELETE" {
			reference := ctx.Params(":reference")
			if !validate.IsDigestValid(reference) {
				ctx.Resp.WriteHeader(http.StatusBadRequest)
				result, _ := module.ReportError(module.DIGEST_INVALID, "Invalid reference format")
				ctx.Resp.Write(result)
				return
			}
		}
	}
}

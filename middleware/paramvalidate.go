package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/validate"
)

func paramChk() macaron.Handler {
	return func(ctx *macaron.Context) {
		if !strings.HasPrefix(ctx.Req.RequestURI, "/v2") ||
			utils.Compare(ctx.Req.RequestURI, "/v2/") == 0 ||
			strings.Contains(ctx.Req.RequestURI, "/v2/_catalog") {
			return
		}

		namespace := ctx.Params(":namespace")
		repository := ctx.Params(":repository")
		if !validate.IsNameValid(namespace) || !validate.IsNameValid(repository) {
			ctx.Resp.WriteHeader(http.StatusBadRequest)
			detail := fmt.Sprintf("%s/%s", namespace, repository)
			result, _ := module.ReportError(module.NAME_INVALID, detail)
			ctx.Resp.Write(result)
			return
		}

		if strings.Contains(ctx.Req.RequestURI, "/manifests/") &&
			(ctx.Req.Method == "PUT" || ctx.Req.Method == "GET") {
			tag := ctx.Params(":tag")
			if !validate.IsTagValid(tag) {
				ctx.Resp.WriteHeader(http.StatusBadRequest)
				detail := fmt.Sprintf("%s", tag)
				result, _ := module.ReportError(module.TAG_INVALID, detail)
				ctx.Resp.Write(result)
				return
			}
		}

		if strings.Contains(ctx.Req.RequestURI, "/blobs/") &&
			(ctx.Req.Method == "HEAD" || ctx.Req.Method == "GET" || ctx.Req.Method == "DELETE") {
			digest := ctx.Params(":digest")
			if !validate.IsDigestValid(digest) {
				ctx.Resp.WriteHeader(http.StatusBadRequest)
				detail := fmt.Sprintf("%s", digest)
				result, _ := module.ReportError(module.DIGEST_INVALID, detail)
				ctx.Resp.Write(result)
				return
			}
		}

		if strings.Contains(ctx.Req.RequestURI, "/manifests/") && ctx.Req.Method == "DELETE" {
			reference := ctx.Params(":reference")
			if !validate.IsDigestValid(reference) {
				ctx.Resp.WriteHeader(http.StatusBadRequest)
				detail := fmt.Sprintf("%s", reference)
				result, _ := module.ReportError(module.DIGEST_INVALID, detail)
				ctx.Resp.Write(result)
				return
			}
		}
	}
}

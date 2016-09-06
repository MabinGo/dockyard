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

func dockerParamChk() macaron.Handler {
	return func(ctx *macaron.Context) {
		if !strings.HasPrefix(ctx.Req.RequestURI, "/v2") ||
			utils.Compare(ctx.Req.RequestURI, "/v2/") == 0 ||
			strings.Contains(ctx.Req.RequestURI, "/v2/_catalog") {
			return
		}

		namespace := ctx.Params(":namespace")
		if !validate.IsNameValid(namespace) {
			ctx.Resp.WriteHeader(http.StatusBadRequest)
			detail := fmt.Sprintf("%s", namespace)
			result, _ := module.ReportError(module.NAME_INVALID, "Invalid namespace format", detail)
			ctx.Resp.Write(result)
			return
		}
		repository := ctx.Params(":repository")
		if !validate.IsRepoValid(repository) {
			ctx.Resp.WriteHeader(http.StatusBadRequest)
			detail := fmt.Sprintf("%s", repository)
			result, _ := module.ReportError(module.NAME_INVALID, "Invalid repository format", detail)
			ctx.Resp.Write(result)
			return
		}

		if strings.Contains(ctx.Req.RequestURI, "/manifests/") &&
			(ctx.Req.Method == "PUT" || ctx.Req.Method == "GET") {
			tag := ctx.Params(":tag")
			if !validate.IsTagValid(tag) {
				ctx.Resp.WriteHeader(http.StatusBadRequest)
				detail := fmt.Sprintf("%s", tag)
				result, _ := module.ReportError(module.TAG_INVALID, "Invalid tag format", detail)
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
				result, _ := module.ReportError(module.DIGEST_INVALID, "Invalid digest format", detail)
				ctx.Resp.Write(result)
				return
			}
		}

		if strings.Contains(ctx.Req.RequestURI, "/manifests/") && ctx.Req.Method == "DELETE" {
			reference := ctx.Params(":reference")
			if !validate.IsDigestValid(reference) {
				ctx.Resp.WriteHeader(http.StatusBadRequest)
				detail := fmt.Sprintf("%s", reference)
				result, _ := module.ReportError(module.DIGEST_INVALID, "Invalid reference format", detail)
				ctx.Resp.Write(result)
				return
			}
		}
	}
}

func appParamChk() macaron.Handler {
	return func(ctx *macaron.Context) {
		if !strings.HasPrefix(ctx.Req.RequestURI, "/app/v1") ||
			strings.Contains(ctx.Req.RequestURI, "/app/v1/search") {
			return
		}

		namespace := ctx.Params(":namespace")
		if !validate.IsNameValid(namespace) {
			ctx.Resp.WriteHeader(http.StatusBadRequest)
			detail := fmt.Sprintf("%s", namespace)
			result, _ := module.ReportError(module.NAME_INVALID, "Invalid namespace format", detail)
			ctx.Resp.Write(result)
			return
		}
		repository := ctx.Params(":repository")
		if !validate.IsRepoValid(repository) {
			ctx.Resp.WriteHeader(http.StatusBadRequest)
			detail := fmt.Sprintf("%s", repository)
			result, _ := module.ReportError(module.NAME_INVALID, "Invalid repository format", detail)
			ctx.Resp.Write(result)
			return
		}

		if strings.Contains(ctx.Req.RequestURI, "/search") ||
			strings.Contains(ctx.Req.RequestURI, "/list") ||
			strings.Contains(ctx.Req.RequestURI, "/meta") ||
			strings.Contains(ctx.Req.RequestURI, "/metasign") ||
			(ctx.Req.Method == "POST") {
			return
		}

		os := ctx.Params(":os")
		arch := ctx.Params(":arch")
		osLen := len(os)
		archLen := len(arch)

		if osLen == 0 || osLen > 128 || archLen == 0 || archLen > 128 {
			ctx.Resp.WriteHeader(http.StatusBadRequest)
			detail := fmt.Sprintf("%s/%s", os, arch)
			result, _ := module.ReportError(module.NAME_UNKNOWN, "Invalid os or arch name", detail)
			ctx.Resp.Write(result)
			return
		}

		if ctx.Req.Method == "PATCH" {
			status := ctx.Params(":status")
			if (status != "done") && (status != "error") {
				ctx.Resp.WriteHeader(http.StatusBadRequest)
				detail := fmt.Sprintf("%s", status)
				result, _ := module.ReportError(module.UNKNOWN, "Invalid status", detail)
				ctx.Resp.Write(result)
				return
			}
		}

		app := ctx.Params(":app")
		if !validate.IsAppValid(app) {
			ctx.Resp.WriteHeader(http.StatusBadRequest)
			detail := fmt.Sprintf("%s", app)
			result, _ := module.ReportError(module.NAME_INVALID, "Invalid app format", detail)
			ctx.Resp.Write(result)
			return
		}

		tag := ctx.Params(":tag")
		if tag != "" && !validate.IsTagValid(tag) {
			ctx.Resp.WriteHeader(http.StatusBadRequest)
			detail := fmt.Sprintf("%s", tag)
			result, _ := module.ReportError(module.TAG_INVALID, "Invalid tag format", detail)
			ctx.Resp.Write(result)
			return
		}
	}
}

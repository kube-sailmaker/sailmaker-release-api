package controller

import (
	"bytes"
	"encoding/json"
	"github.com/skhatri/api-router-go/router"
	"github.com/skhatri/api-router-go/router/model"
	"github.com/skhatri/kube-sailmaker-release/k8s/middleware"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func getCrdInstance(web *router.WebRequest) *model.Container {
	gvr := schema.GroupVersionResource{
		Resource: web.GetQueryParam("resource-type"),
		Group:    web.GetQueryParam("resource-group"),
		Version:  web.GetQueryParam("resource-version"),
	}
	cres, err := middleware.GetCrdByName(
		web.GetQueryParam("namespace"), gvr, web.GetQueryParam("resource-name"))
	if err != nil {
		return model.ErrorResponse(model.MessageItem{
			Code:    "crd get error",
			Message: err.Error(),
		}, 500)
	}
	return model.Response(cres)
}

func getCrdInstanceList(web *router.WebRequest) *model.Container {
	gvr := schema.GroupVersionResource{
		Resource: web.GetQueryParam("resource-type"),
		Group:    web.GetQueryParam("resource-group"),
		Version:  web.GetQueryParam("resource-version"),
	}
	cresList, err := middleware.GetCrdInstanceList(
		web.GetQueryParam("namespace"), gvr)
	if err != nil {
		return model.ErrorResponse(model.MessageItem{
			Code:    "crd instance list error",
			Message: err.Error(),
		}, 500)
	}
	return model.Response(cresList)
}

func getCrds(web *router.WebRequest) *model.Container {
	crdList, err := middleware.GetCrds()
	if err != nil {
		return model.ErrorResponse(model.MessageItem{
			Code:    "crd list error",
			Message: err.Error(),
		}, 500)
	}
	return model.Response(crdList)
}

func performRelease(web *router.WebRequest) *model.Container {
	releaseRequest := ReleaseRequest{}
	buff := bytes.NewBuffer(web.Body)
	err := json.NewDecoder(buff).Decode(&releaseRequest)
	if err != nil {
		return model.ErrorResponse(model.MessageItem{
			Code:    "request-error",
			Message: err.Error(),
		}, 400)
	}

	res, err := middleware.CreateCustomResourceInstance(&middleware.CrdInstanceInput{
		Namespace:  releaseRequest.Namespace,
		Name:       releaseRequest.Name,
		Spec:       releaseRequest.Spec,
		CrdKind:    "ReleaseRequest",
		CrdName:    "releaserequests",
		CrdGroup:   "deploy.kubesailmaker.io",
		CrdVersion: "v1alpha1",
	})

	if err != nil {
		return model.ErrorResponse(model.MessageItem{
			Code:    "create-error",
			Message: err.Error(),
		}, 500)
	}
	return model.Response(res)
}

type ReleaseRequestSpec struct {
	Apps []*ReleaseItem `json:"apps"`
}

type ReleaseItem struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
type ReleaseRequest struct {
	Name      string             `json:"name"`
	Namespace string             `json:"namespace"`
	Spec      ReleaseRequestSpec `json:"spec"`
}

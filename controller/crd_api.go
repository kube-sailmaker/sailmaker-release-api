package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/skhatri/api-router-go/router"
	"github.com/skhatri/api-router-go/router/model"
	"github.com/skhatri/kube-sailmaker-release/k8s/middleware"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

func getCrdInstance(web *router.WebRequest) *model.Container {
	gvr := schema.GroupVersionResource{
		Resource: web.GetQueryParam("resource-type"),
		Group:    web.GetQueryParam("resource-group"),
		Version:  web.GetQueryParam("resource-version"),
	}
	resourceName := ""
	if resourceName = web.GetPathParam("resource_name"); resourceName == "" {
		resourceName = web.GetQueryParam("resource-name")
	}
	cres, err := middleware.GetCrdByName(
		web.GetQueryParam("namespace"), gvr, resourceName)
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
	Namespace string         `json:"release-namespace"`
	Apps      []*ReleaseItem `json:"apps"`
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

func updateRelease(web *router.WebRequest) *model.Container {

	releaseName := web.GetPathParam("release-name")
	appId := web.GetPathParam("app-id")
	namespace := web.GetQueryParam("namespace")

	updateRequest := ReleaseUpdateRequest{}
	buff := bytes.NewBuffer(web.Body)
	err := json.NewDecoder(buff).Decode(&updateRequest)
	if err != nil {
		return model.ErrorResponse(model.MessageItem{
			Code:    "request-error",
			Message: err.Error(),
		}, 400)
	}

	resourceInstance, err := middleware.GetCrdByName(namespace, schema.GroupVersionResource{
		Group:    "deploy.kubesailmaker.io",
		Version:  "v1alpha1",
		Resource: "releaserequests",
	}, releaseName)
	if err != nil {
		return model.ErrorResponse(model.MessageItem{
			Code:    "request-update-error",
			Message: fmt.Sprintf("Release %s not found %s", releaseName, err.Error()),
		}, 400)
	}
	metadata := resourceInstance.Metadata

	appsSpec := resourceInstance.Spec

	existingPayloadBuffer := bytes.Buffer{}
	json.NewEncoder(&existingPayloadBuffer).Encode(appsSpec)
	existingPayload := existingPayloadBuffer.String()

	apps := AppSpec{}
	json.NewDecoder(bytes.NewBuffer([]byte(existingPayload))).Decode(&apps)

	updatedList := make([]AppItem, 0)
	found := false
	appNames := make([]string, 0)
	for _, app := range apps.Apps {
		appName := app.Name
		if app.Alias != "" {
			appName = app.Alias
		}
		if appId == appName &&
			updateRequest.Name == appName &&
			updateRequest.Version == app.Version {

			appMetadata := app.Metadata
			if appMetadata == nil {
				appMetadata = make(map[string]string)
			}
			appMetadata["resource/name"] = updateRequest.Resource.Name
			appMetadata["resource/created"] = updateRequest.Resource.Created
			appMetadata["resource/type"] = updateRequest.Resource.Type
			appMetadata["resource/updated"] = updateRequest.Resource.Updated

			updatedList = append(updatedList, AppItem{
				Name:     app.Name,
				Alias:    app.Alias,
				Version:  app.Version,
				Status:   updateRequest.Status,
				Metadata: appMetadata,
			})
			found = true
		} else {
			updatedList = append(updatedList, app)
		}
		appNames = append(appNames, appName)
	}

	if !found {
		return model.ErrorResponse(model.MessageItem{
			Code:    "update-error",
			Message: fmt.Sprintf("unknown app %s in release %s, available apps [%s]", appId, releaseName, strings.Join(appNames, ", ")),
		}, 400)
	}

	apps.Apps = updatedList

	res, err := middleware.UpdateCustomResourceInstance(metadata, &middleware.CrdInstanceInput{
		Namespace:  namespace,
		Name:       releaseName,
		Spec:       apps,
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

type ReleaseUpdateRequest struct {
	ReleaseItem
	Status   string         `json:"status"`
	Resource ResourceDetail `json:"resource"`
}

type ResourceDetail struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

type AppSpec struct {
	Apps []AppItem `json:"apps"`
}

type AppItem struct {
	Name     string            `json:"name"`
	Alias    string            `json:"alias"`
	Version  string            `json:"version"`
	Metadata map[string]string `json:"metadata""`
	Status   string            `json:"status"`
	Message  string            `json:"message"`
}

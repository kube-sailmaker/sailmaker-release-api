package controller

import (
	"github.com/skhatri/api-router-go/router"
	"github.com/skhatri/api-router-go/router/model"
	"github.com/skhatri/kube-sailmaker-release/k8s/middleware"
)

func namespaceApiHandler(_ *router.WebRequest) *model.Container {
	namespaceList, err := middleware.GetNamespace()
	if err != nil {
		return model.ErrorResponse(model.MessageItem{
			Code:    "namespace-error",
			Message: "Could not get namespace list",
		}, 500)
	}
	return model.Response(namespaceList)
}

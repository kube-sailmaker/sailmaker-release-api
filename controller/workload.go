package controller

import (
	"github.com/skhatri/api-router-go/router"
	"github.com/skhatri/api-router-go/router/model"
	"github.com/skhatri/kube-sailmaker-release/k8s/middleware"
)

func fetchDeployments(web *router.WebRequest) *model.Container {
	var namespace = web.GetQueryParam("namespace")
	return getWorkload(namespace, "deployment")
}

func fetchJobs(web *router.WebRequest) *model.Container {
	var namespace = web.GetQueryParam("namespace")
	return getWorkload(namespace, "job")
}

func fetchDaemonsets(web *router.WebRequest) *model.Container {
	var namespace = web.GetQueryParam("namespace")
	return getWorkload(namespace, "daemonset")
}

func fetchStatefulsets(web *router.WebRequest) *model.Container {
	var namespace = web.GetQueryParam("namespace")
	return getWorkload(namespace, "statefulset")
}

func getWorkload(namespace string, kind string) *model.Container {
	workload, err := middleware.GetWorkload(namespace, kind)
	if err != nil {
		return model.ErrorResponse(model.MessageItem{
			Code:    "list-error",
			Message: err.Error(),
		}, 500)
	}
	var items = make([]interface{}, 0, len(workload))
	for _, w := range workload {
		items = append(items, w)
	}
	return model.ListResponse(items)
}

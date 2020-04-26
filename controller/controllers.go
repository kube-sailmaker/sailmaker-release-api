package controller

import (
	"github.com/skhatri/api-router-go/router"
	"github.com/skhatri/api-router-go/router/functions"
	"github.com/skhatri/api-router-go/router/settings"
)

func Configure(configurer router.ApiConfigurer) {
	var _settings = settings.GetSettings()
	configurer.Get("/api/namespaces", namespaceApiHandler).
		Get("/status", functions.StatusFunc).
		Get("/api/deployments", fetchDeployments).
		Get("/api/statefulsets", fetchStatefulsets).
		Get("/api/jobs", fetchJobs).
		GetIf(_settings.IsToggleOn("can_read_crds")).
		Add("/api/crd-instances", getCrdInstanceList).
		Add("/api/crd-instances/:resource_name", getCrdInstance).
		Add("/api/crd-instance", getCrdInstance).
		Add("/api/crds", getCrds).
		Done().
		PostIf(_settings.IsToggleOn("can_write_crds")).
			Register("/api/release", performRelease).
		PostIf(_settings.IsToggleOn("can_write_crds")).
			Add("/api/release/:release-name/apps/:app-id", updateRelease).
		Done().
		GetIf(_settings.IsToggleOn("daemonset_endpoint")).Register("/api/daemonsets", fetchDaemonsets)
}

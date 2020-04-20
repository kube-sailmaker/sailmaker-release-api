package main

import (
	"github.com/skhatri/api-router-go/router"
	"github.com/skhatri/api-router-go/starter"
	"github.com/skhatri/kube-sailmaker-release/controller"
	"os"
)

func main() {
	starter.StartApp(os.Args, 6264, func(configurer router.ApiConfigurer) {
		controller.Configure(configurer)
	})
}

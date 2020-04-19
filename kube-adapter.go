package main

import (
	"github.com/skhatri/api-router-go/starter"
	"github.com/skhatri/api-router-go/router"
	"github.com/skhatri/k8s-read/controller"
	"os"
)

func main() {
	starter.StartApp(os.Args, 6100, func(configurer router.ApiConfigurer) {
		controller.Configure(configurer)
	})
}

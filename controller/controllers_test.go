package controller

import (
	"fmt"
	"github.com/skhatri/api-router-go/router"
	"github.com/skhatri/kube-sailmaker-release/test"
	"testing"
)

type MockApiConfigurer struct {
	m map[string]router.HandlerFunc
}

func (mc *MockApiConfigurer) Get(uri string, hf router.HandlerFunc) router.ApiConfigurer {

	mc.m[makeKey("get", uri)] = hf
	return mc
}

func makeKey(method string, uri string) string {
	return fmt.Sprintf("%s%s", method, uri)
}

func (mc *MockApiConfigurer) Post(uri string, hf router.HandlerFunc) router.ApiConfigurer {
	mc.m[makeKey("post", uri)] = hf
	return mc
}

func (mc *MockApiConfigurer) Method(method string, uri string, hf router.HandlerFunc) router.ApiConfigurer {
	mc.m[makeKey(method,uri)] = hf
	return mc
}

func (mc *MockApiConfigurer) GetIf(cond bool) *router.ConditionalMethodBuilder {
	return &router.ConditionalMethodBuilder{
		Method: "GET",
		Check:    cond,
		Delegate: nil,
	}
}
func (mc *MockApiConfigurer) PostIf(cond bool) *router.ConditionalMethodBuilder {
	return &router.ConditionalMethodBuilder{
		Method: "POST",
		Check:    cond,
		Delegate: nil,
	}
}


func TestRegistersApis(t *testing.T) {
	m := make(map[string]router.HandlerFunc)
	apiConfigurer := &MockApiConfigurer{
		m: m,
	}
	Configure(apiConfigurer)
	test.NotNull(t, m[makeKey("get", "/api/namespaces")])
	test.NotNull(t, m[makeKey("get", "/api/deployments")])
}

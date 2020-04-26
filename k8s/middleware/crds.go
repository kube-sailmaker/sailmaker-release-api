package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skhatri/kube-sailmaker-release/k8s/client"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/dynamic"
	"log"
)

//GetCrdByName Kubernetes Workload of given Custom Resource Type in a namespace.
func GetCrdByName(namespace string, gvr schema.GroupVersionResource, resourceName string) (*CustomResourceInstance, error) {
	if namespace == "" {
		return nil, errors.New("namespace is required")
	}
	dynamicClient := *(client.GetDynamicClient())

	namespaceResInt := dynamicClient.Resource(gvr).Namespace(namespace)
	resource, err := namespaceResInt.Get(context.TODO(), resourceName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("CRD Resource List error %s", err.Error()))
	}

	resourceData, err := resource.MarshalJSON()
	var cres = CustomResourceInstance{}
	if err == nil {
		buff := bytes.NewBuffer(resourceData)
		json.NewDecoder(buff).Decode(&cres)
	}

	return &cres, nil
}

//GetCrdInstanceList returns custom resource instances for a group
func GetCrdInstanceList(namespace string, gvr schema.GroupVersionResource) ([]CustomResourceInstanceSummary, error) {
	dynamicClient := *(client.GetDynamicClient())

	var namespaceResInt dynamic.ResourceInterface = dynamicClient.Resource(gvr)
	if namespace != "" {
		namespaceResInt = namespaceResInt.(dynamic.NamespaceableResourceInterface).Namespace(namespace)
	}
	resources, err := namespaceResInt.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("CRD Resource List error %s", err.Error()))
	}

	customResources := make([]CustomResourceInstanceSummary, 0)
	for _, res := range resources.Items {
		groupKind := res.GroupVersionKind()
		customResources = append(customResources, CustomResourceInstanceSummary{
			Namespace: res.GetNamespace(),
			Name:      res.GetName(),
			Version:   groupKind.Version,
			Group:     groupKind.Group,
			Resource:  gvr.Resource,
			Link: fmt.Sprintf("/api/crd-instance?resource-group=%s&resource-type=%s&resource-version=%s&namespace=%s&resource-name=%s",
				groupKind.Group, gvr.Resource, groupKind.Version, res.GetNamespace(), res.GetName()),
		})
	}
	return customResources, nil
}

//GetCrds returns list of CRDs registered against the Api Server
func GetCrds() ([]CrdSummary, error) {
	crds, err := client.GetExtensionsClient().CustomResourceDefinitions().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	crdList := make([]CrdSummary, 0)
	for _, crd := range crds.Items {
		effectiveVersion := findEffectiveVersion(crd)
		crdList = append(crdList, CrdSummary{
			Name:         crd.Name,
			Group:        crd.Spec.Group,
			Kind:         crd.Spec.Names.Kind,
			Version:      effectiveVersion,
			ResourceType: crd.Spec.Names.Plural,
			Link: fmt.Sprintf("/api/crd-instances?resource-group=%s&resource-type=%s&resource-version=%s",
				crd.Spec.Group, crd.Spec.Names.Plural, effectiveVersion),
		})
	}
	return crdList, nil
}

func findEffectiveVersion(crd v1beta1.CustomResourceDefinition) string {
	var effectiveVersion = ""
	for _, v := range crd.Spec.Versions {
		if v.Storage && v.Served {
			effectiveVersion = v.Name
			break
		}
	}
	return effectiveVersion
}

//CreateeCustomResourceInstance register an instance of a releasereequest
func CreateCustomResourceInstance(crdInstanceInput *CrdInstanceInput) (*CustomResourceResponse, error) {
	requestKind := crdInstanceInput.CrdKind
	crdVersion := crdInstanceInput.CrdVersion
	crdName := crdInstanceInput.CrdName
	crdGroup := crdInstanceInput.CrdGroup

	crdApiVersion := fmt.Sprintf("%s/%s", crdGroup, crdVersion)
	crdReqAnnotation := fmt.Sprintf("%s/request-id", crdGroup)
	crdPayloadAnnotation := fmt.Sprintf("%s/payload", crdGroup)
	gvr := schema.GroupVersionResource{
		Resource: crdName,
		Group:    crdGroup,
		Version:  crdVersion,
	}

	var uid = uuid.NewUUID()
	var requestId = string(uid)
	if crdInstanceInput.Name == "" {
		crdInstanceInput.Name = fmt.Sprintf("release-%s", requestId)
	}
	spec := crdInstanceInput.Spec

	buff := bytes.Buffer{}
	json.NewEncoder(&buff).Encode(&crdInstanceInput)
	payload := buff.String()
	var resourceDef = unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       requestKind,
			"apiVersion": crdApiVersion,
			"metadata": map[string]interface{}{
				"name": crdInstanceInput.Name,
				"labels": map[string]string{
					"category": "release",
				},
				"annotations": map[string]string{
					crdReqAnnotation:     requestId,
					crdPayloadAnnotation: payload,
				},
			},
			"spec": spec,
		},
	}
	crdResult, err := (*client.GetDynamicClient()).Resource(gvr).Namespace(crdInstanceInput.Namespace).
		Create(context.TODO(), &resourceDef, metav1.CreateOptions{})

	if err != nil {
		log.Println("error creating crd instance", err)
		return nil, err
	}

	return &CustomResourceResponse{
		RequestId: requestId,
		ResourceReference: &CustomResourceInstanceSummary{
			Resource:  crdName,
			Version:   crdVersion,
			Group:     crdGroup,
			Namespace: crdResult.GetNamespace(),
			Name:      crdResult.GetName(),
			Link: fmt.Sprintf("/api/crd-instance?resource-type=%s&resource-group=%s&resource-version=v1alpha1&namespace=%s&resource-name=%s",
				crdName, crdGroup, crdResult.GetNamespace(), crdResult.GetName()),
		},
	}, nil
}

//CreateeCustomResourceInstance register an instance of a releasereequest
func UpdateCustomResourceInstance(metadata map[string]interface{}, crdInstanceInput *CrdInstanceInput) (*CustomResourceResponse, error) {
	requestKind := crdInstanceInput.CrdKind
	crdVersion := crdInstanceInput.CrdVersion
	crdName := crdInstanceInput.CrdName
	crdGroup := crdInstanceInput.CrdGroup

	crdApiVersion := fmt.Sprintf("%s/%s", crdGroup, crdVersion)
	crdPayloadAnnotation := fmt.Sprintf("%s/payload", crdGroup)
	
	gvr := schema.GroupVersionResource{
		Resource: crdName,
		Group:    crdGroup,
		Version:  crdVersion,
	}

	var uid = uuid.NewUUID()
	var requestId = string(uid)
	if crdInstanceInput.Name == "" {
		crdInstanceInput.Name = fmt.Sprintf("release-%s", requestId)
	}
	spec := crdInstanceInput.Spec

	buff := bytes.Buffer{}
	json.NewEncoder(&buff).Encode(&crdInstanceInput)
	payload := buff.String()
	
	metadata[crdPayloadAnnotation] = payload

	var resourceDef = unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       requestKind,
			"apiVersion": crdApiVersion,
			"metadata": metadata,
			"spec": spec,
		},
	}
	crdResult, err := (*client.GetDynamicClient()).Resource(gvr).Namespace(crdInstanceInput.Namespace).
		Create(context.TODO(), &resourceDef, metav1.CreateOptions{})

	if err != nil {
		log.Println("error creating crd instance", err)
		return nil, err
	}

	return &CustomResourceResponse{
		RequestId: requestId,
		ResourceReference: &CustomResourceInstanceSummary{
			Resource:  crdName,
			Version:   crdVersion,
			Group:     crdGroup,
			Namespace: crdResult.GetNamespace(),
			Name:      crdResult.GetName(),
			Link: fmt.Sprintf("/api/crd-instance?resource-type=%s&resource-group=%s&resource-version=v1alpha1&namespace=%s&resource-name=%s",
				crdName, crdGroup, crdResult.GetNamespace(), crdResult.GetName()),
		},
	}, nil
}

type CustomResourceInstance struct {
	Spec     map[string]interface{} `json:"spec"`
	Metadata map[string]interface{} `json:"metadata"`
}
type CustomResourceInstanceSummary struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Group     string `json:"group"`
	Version   string `json:"version"`
	Resource  string `json:"resource"`
	Link      string `json:"link"`
}

type CrdSummary struct {
	Name         string `json:"name"`
	Group        string `json:"group"`
	ResourceType string `json:"resource-type"`
	Kind         string `json:"kind"`
	Version      string `json:"version"`
	Link         string `json:"link"`
}

type CrdInstanceInput struct {
	Name       string      `json:"name"`
	Namespace  string      `json:"namespace"`
	Spec       interface{} `json:"spec"`
	CrdKind    string      `json:"kind"`
	CrdVersion string      `json:"version"`
	CrdName    string      `json:"crd-name"`
	CrdGroup   string      `json:"group-name"`
}

type CustomResourceResponse struct {
	RequestId         string                         `json:"request-id"`
	ResourceReference *CustomResourceInstanceSummary `json:"ref"`
}

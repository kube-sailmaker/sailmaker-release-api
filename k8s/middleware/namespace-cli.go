package middleware

import (
	"context"
	"github.com/skhatri/kube-sailmaker-release/k8s/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NamespaceList struct {
	Namespaces []string `json:"namespaces"`
}

//GetNamespace returns list of namespaces in the current cluster
func GetNamespace() (*NamespaceList, error) {
	client := client.GetClient()
	namespaceList, nameErr := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if nameErr != nil {
		return nil, nameErr
	}
	var names = make([]string, 0, len(namespaceList.Items))
	for _, item := range namespaceList.Items {
		names = append(names, item.Name)
	}
	return &NamespaceList{Namespaces: names}, nil
}

package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/skhatri/kube-sailmaker-release/k8s/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Workload struct {
	Namespace string `json:"namespace,omitempty"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Replicas  int32  `json:"replicas"`
}

//Get Kubernetes Workload of given kind in a namespace.
func GetWorkload(namespace string, kind string) ([]Workload, error) {
	k8s := client.GetClient()
	if namespace == "" {
		return nil, errors.New("namespace is required")
	}
	if namespace == "any" {
		namespace = ""
	}
	switch kind {
	case "deployment":
		return getDeployments(k8s, namespace)
	case "statefulset":
		return getStatefulSets(k8s, namespace)
	case "daemonset":
		return getDaemonSets(k8s, namespace)
	case "job":
		return getJobs(k8s, namespace)
	}
	return nil, errors.New(fmt.Sprintf("workload retrieval for kind %s is unsupported", kind))
}

func getDeployments(k8s *kubernetes.Clientset, namespace string) ([]Workload, error) {
	depList, depErr := k8s.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if depErr != nil {
		return nil, depErr
	}
	workload := make([]Workload, 0)
	for _, deployment := range depList.Items {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			workload = append(workload, Workload{
				Namespace: deployment.Namespace,
				Kind:      "deployment",
				Name:      deployment.Name,
				Image:     container.Image,
				Replicas:  *deployment.Spec.Replicas,
			})
		}
	}
	return workload, nil
}

func getStatefulSets(k8s *kubernetes.Clientset, namespace string) ([]Workload, error) {
	stList, stErr := k8s.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if stErr != nil {
		return nil, stErr
	}
	workload := make([]Workload, 0)
	for _, statefulset := range stList.Items {
		for _, container := range statefulset.Spec.Template.Spec.Containers {
			workload = append(workload, Workload{
				Namespace: statefulset.Namespace,
				Kind:      "statefulset",
				Name:      statefulset.Name,
				Image:     container.Image,
				Replicas:  *statefulset.Spec.Replicas,
			})

		}
	}
	return workload, nil
}

func getDaemonSets(k8s *kubernetes.Clientset, namespace string) ([]Workload, error) {
	stList, stErr := k8s.AppsV1().DaemonSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if stErr != nil {
		return nil, stErr
	}
	workload := make([]Workload, 0)
	for _, daemonset := range stList.Items {
		for _, container := range daemonset.Spec.Template.Spec.Containers {
			workload = append(workload, Workload{
				Namespace: daemonset.Namespace,
				Kind:      "daemonset",
				Name:      daemonset.Name,
				Image:     container.Image,
				Replicas:  1,
			})

		}
	}
	return workload, nil
}

func getJobs(k8s *kubernetes.Clientset, namespace string) ([]Workload, error) {
	jobList, jobErr := k8s.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{})
	if jobErr != nil {
		return nil, jobErr
	}
	workload := make([]Workload, 0)
	for _, job := range jobList.Items {

		for _, container := range job.Spec.Template.Spec.Containers {
			workload = append(workload, Workload{
				Namespace: job.Namespace,
				Kind:      "job",
				Name:      job.Name,
				Image:     container.Image,
				Replicas:  1,
			})

		}
	}
	return workload, nil
}

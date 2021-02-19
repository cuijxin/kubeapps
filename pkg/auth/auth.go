package auth

import (
	"context"

	authorizationapi "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	discovery "k8s.io/client-go/discovery"
	authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

type resource struct {
	APIVersion string
	Kind       string
	Namespace  string
}

type k8sAuthInterface interface {
	GetResourceList(groupVersion string) (*metav1.APIResourceList, error)
	CanI(verb, group, resource, namespace string) (bool, error)
}

type k8sAuth struct {
	AuthCli      authorizationv1.AuthorizationV1Interface
	DiscoveryCli discovery.DiscoveryInterface
}

func (u k8sAuth) GetResourceList(groupVersion string) (*metav1.APIResourceList, error) {
	return u.DiscoveryCli.ServerResourcesForGroupVersion(groupVersion)
}

func (u k8sAuth) CanI(verb, group, resource, namespace string) (bool, error) {
	attr := &authorizationapi.ResourceAttributes{
		Group:     group,
		Resource:  resource,
		Verb:      verb,
		Namespace: namespace,
	}
	res, err := u.AuthCli.SelfSubjectAccessReviews().Create(context.TODO(), &authorizationapi.SelfSubjectAccessReview{
		Spec: authorizationapi.SelfSubjectAccessReviewSpec{
			ResourceAttributes: attr,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	return res.Status.Allowed, nil
}

// UserAuth contains information to check user permissions
type UserAuth struct {
	k8sAuth k8sAuthInterface
}

// Action represents a specific set of verbs against a resource
type Action struct {
	APIVersion  string   `json:"apiGroup"`
	Resource    string   `json:"resource"`
	Namespace   string   `json:"namespace"`
	ClusterWide bool     `json:"clusterWide"`
	Verbs       []string `json:"verbs"`
}

// Checker for the exported funcs
type Checker interface {
	ValidateForNamespace(namespace string) (bool, error)
	GetForbiddenActions(namespace, action, manifest string) ([]Action, error)
}

// NewAuth creates an auth agent
// func NewAuth(token, clusterName string, clustersConfig)

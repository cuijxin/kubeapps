package auth

import (
	"strings"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	discovery "k8s.io/client-go/discovery"
)

type fakeK8sAuth struct {
	DiscoveryCli discovery.DiscoveryInterface
	canIResult   bool
	canIError    error
}

func (u fakeK8sAuth) Validate() error {
	return nil
}

func (u fakeK8sAuth) GetResourcelist(groupVersion string) (*metav1.APIResourceList, error) {
	g, err := u.DiscoveryCli.ServerResourcesForGroupVersion(groupVersion)
	if err != nil && strings.Contains(err.Error(), "not found") {
		// Fake DiscoveryCli doesn;t return a valid NotFound error so we need to forge it
		err = k8sErrors.NewNotFound(schema.GroupResource{}, groupVersion)
	}
	return g, err
}

func (u fakeK8sAuth) CanI(verb, group, resource, namespace string) (bool, error) {
	return u.canIResult, u.canIError
}

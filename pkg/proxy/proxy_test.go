package proxy

import (
	"reflect"
	"testing"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
)

type AppOverviewTest struct {
	AppOverview
	Manifest string
}

func newFakeProxyWithManifest(existingTillerReleases []AppOverviewTest) *Proxy {
	helmClient := helm.FakeClient{}
	// Populate Fake helm client with releases
	for _, r := range existingTillerReleases {
		status := release.Status_DEPLOYED
		if r.Status == "DELETED" {
			status = release.Status_DELETED
		} else if r.Status == "FAILED" {
			status = release.Status_FAILED
		}
		version := int32(1)
		// Increment version number (helm revision counter)
		// if the same release name has been already added
		for _, versionAdded := range helmClient.Rels {
			if r.ReleaseName == versionAdded.GetName() {
				version++
			}
		}
		helmClient.Rels = append(helmClient.Rels, &release.Release{
			Name:      r.ReleaseName,
			Namespace: r.Namespace,
			Version:   version,
			Chart: &chart.Chart{
				Metadata: &chart.Metadata{
					Version: r.Version,
					Icon:    r.Icon,
					Name:    r.Chart,
				},
			},
			Info: &release.Info{
				Status: &release.Status{
					Code: status,
				},
			},
			Manifest: r.Manifest,
		})
	}
	kubeClient := fake.NewSimpleClientset()
	return NewProxy(kubeClient, &helmClient, 300)
}

func newFakeProxy(existingTillerReleases []AppOverview) *Proxy {
	releasesWithManifest := []AppOverviewTest{}
	for _, r := range existingTillerReleases {
		releasesWithManifest = append(releasesWithManifest, AppOverviewTest{r, ""})
	}
	return newFakeProxyWithManifest(releasesWithManifest)
}

func TestListAllReleases(t *testing.T) {
	app1 := AppOverview{
		"foo",
		"1.0.0",
		"my_ns",
		"icon.png",
		"DEPLOYED",
		"wordpress",
		chart.Metadata{
			Version: "1.0.0",
			Icon:    "icon.png",
			Name:    "wordpress",
		},
	}
	app2 := AppOverview{
		"bar",
		"1.0.0",
		"other_ns",
		"icon2.png",
		"DELETED",
		"wordpress",
		chart.Metadata{
			Version: "1.0.0",
			Icon:    "icon2.png",
			Name:    "wordpress",
		},
	}
	proxy := newFakeProxy([]AppOverview{app1, app2})

	// Should return all the releases if no namespace is given
	releases, err := proxy.ListReleases("", 256, "")
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if len(releases) != 2 {
		t.Errorf("It should return both releases")
	}
	if !reflect.DeepEqual([]AppOverview{app1, app2}, releases) {
		t.Log(releases[0].ChartMetadata)
		t.Log(app1.ChartMetadata)
		t.Errorf("Unexpected list of releases %v", releases)
	}
}

func TestListNamespacedRelease(t *testing.T) {
	app1 := AppOverview{
		"foo",
		"1.0.0",
		"my_ns",
		"icon.png",
		"DEPLOYED",
		"wordpress",
		chart.Metadata{
			Version: "1.0.0",
			Icon:    "icon.png",
			Name:    "wordpress",
		},
	}
	app2 := AppOverview{
		"bar",
		"1.0.0",
		"other_ns",
		"icon2.png",
		"DELETED",
		"wordpress",
		chart.Metadata{
			Version: "1.0.0",
			Icon:    "icon2.png",
			Name:    "wordpress",
		},
	}
	proxy := newFakeProxy([]AppOverview{app1, app2})

	releases, err := proxy.ListReleases(app1.Namespace, 256, "")
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if len(releases) != 1 {
		t.Errorf("It should return both releases")
	}
	if !reflect.DeepEqual([]AppOverview{app1}, releases) {
		t.Errorf("Unexpected list of releases %v", releases)
	}
}

func TestListOldRelease(t *testing.T) {
	app := AppOverview{
		"foo",
		"1.0.0",
		"my_ns",
		"icon.png",
		"DEPLOYED",
		"wordpress",
		chart.Metadata{
			Version: "1.0.0",
			Icon:    "icon.png",
			Name:    "wordpress",
		},
	}
	appUpgraded := AppOverview{
		"foo",
		"1.0.1",
		"my_ns",
		"icon.png",
		"FAILED",
		"wordpress",
		chart.Metadata{
			Version: "1.0.1",
			Icon:    "icon.png",
			Name:    "wordpress",
		},
	}
	proxy := newFakeProxy([]AppOverview{app, appUpgraded})

	// Should avoid old release versions
	releases, err := proxy.ListReleases(app.Namespace, 256, "")
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if len(releases) != 1 {
		t.Errorf("It should return a single release")
	}
	if releases[0].ReleaseName != "foo" && releases[0].Status != "FAILED" {
		t.Errorf("It should group releases by release name")
	}
	if !reflect.DeepEqual([]AppOverview{appUpgraded}, releases) {
		t.Errorf("Unexpected list of releases %v", releases)
	}
}

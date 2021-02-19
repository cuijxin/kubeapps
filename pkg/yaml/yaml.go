package yaml

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func flattenToV1(objs []runtime.Object) []*unstructured.Unstructured {
	ret := make([]*unstructured.Unstructured, 0, len(objs))
	for _, obj := range objs {
		switch o := obj.(type) {
		case *unstructured.UnstructuredList:
			for i := range o.Items {
				ret = append(ret, &o.Items[i])
			}
		case *unstructured.Unstructured:
			ret = append(ret, o)
		default:
			panic("Unexpected unstructured  object type")
		}
	}

	return ret
}

// ParseObjects returns an Unstructured object list based on the content of a YAML manifest
func ParseObjects(manifest string) ([]*unstructured.Unstructured, error) {
	r := strings.NewReader(manifest)
	decoder := yaml.NewYAMLReader(bufio.NewReader(r))
	ret := []runtime.Object{}
	nullResult := []byte("null")

	for {
		// This reader will return a single K8s resource at the time based on the --- separator
		objManifest, err := decoder.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		jsondata, err := yaml.ToJSON(objManifest)
		if err != nil {
			return nil, err
		}

		if bytes.Equal(jsondata, nullResult) {
			continue
		}

		obj, _, err := unstructured.UnstructuredJSONScheme.Decode(jsondata, nil, nil)
		if err != nil {
			return nil, err
		}
		ret = append(ret, obj)
	}

	return flattenToV1(ret), nil
}

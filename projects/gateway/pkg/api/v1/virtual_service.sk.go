// Code generated by solo-kit. DO NOT EDIT.

package v1

import (
	"log"
	"sort"

	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewVirtualService(namespace, name string) *VirtualService {
	virtualservice := &VirtualService{}
	virtualservice.SetMetadata(core.Metadata{
		Name:      name,
		Namespace: namespace,
	})
	return virtualservice
}

func (r *VirtualService) SetMetadata(meta core.Metadata) {
	r.Metadata = meta
}

func (r *VirtualService) SetStatus(status core.Status) {
	r.Status = status
}

func (r *VirtualService) Hash() uint64 {
	metaCopy := r.GetMetadata()
	metaCopy.ResourceVersion = ""
	return hashutils.HashAll(
		metaCopy,
		r.VirtualHost,
		r.SslConfig,
	)
}

type VirtualServiceList []*VirtualService

// namespace is optional, if left empty, names can collide if the list contains more than one with the same name
func (list VirtualServiceList) Find(namespace, name string) (*VirtualService, error) {
	for _, virtualService := range list {
		if virtualService.GetMetadata().Name == name {
			if namespace == "" || virtualService.GetMetadata().Namespace == namespace {
				return virtualService, nil
			}
		}
	}
	return nil, errors.Errorf("list did not find virtualService %v.%v", namespace, name)
}

func (list VirtualServiceList) AsResources() resources.ResourceList {
	var ress resources.ResourceList
	for _, virtualService := range list {
		ress = append(ress, virtualService)
	}
	return ress
}

func (list VirtualServiceList) AsInputResources() resources.InputResourceList {
	var ress resources.InputResourceList
	for _, virtualService := range list {
		ress = append(ress, virtualService)
	}
	return ress
}

func (list VirtualServiceList) Names() []string {
	var names []string
	for _, virtualService := range list {
		names = append(names, virtualService.GetMetadata().Name)
	}
	return names
}

func (list VirtualServiceList) NamespacesDotNames() []string {
	var names []string
	for _, virtualService := range list {
		names = append(names, virtualService.GetMetadata().Namespace+"."+virtualService.GetMetadata().Name)
	}
	return names
}

func (list VirtualServiceList) Sort() VirtualServiceList {
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].GetMetadata().Less(list[j].GetMetadata())
	})
	return list
}

func (list VirtualServiceList) Clone() VirtualServiceList {
	var virtualServiceList VirtualServiceList
	for _, virtualService := range list {
		virtualServiceList = append(virtualServiceList, resources.Clone(virtualService).(*VirtualService))
	}
	return virtualServiceList
}

func (list VirtualServiceList) Each(f func(element *VirtualService)) {
	for _, virtualService := range list {
		f(virtualService)
	}
}

func (list VirtualServiceList) EachResource(f func(element resources.Resource)) {
	for _, virtualService := range list {
		f(virtualService)
	}
}

func (list VirtualServiceList) AsInterfaces() []interface{} {
	var asInterfaces []interface{}
	list.Each(func(element *VirtualService) {
		asInterfaces = append(asInterfaces, element)
	})
	return asInterfaces
}

var _ resources.Resource = &VirtualService{}

// Kubernetes Adapter for VirtualService

func (o *VirtualService) GetObjectKind() schema.ObjectKind {
	t := VirtualServiceCrd.TypeMeta()
	return &t
}

func (o *VirtualService) DeepCopyObject() runtime.Object {
	return resources.Clone(o).(*VirtualService)
}

var (
	VirtualServiceGVK = schema.GroupVersionKind{
		Version: "v1",
		Group:   "gateway.solo.io",
		Kind:    "VirtualService",
	}
	VirtualServiceCrd = crd.NewCrd(
		"virtualservices",
		VirtualServiceGVK.Group,
		VirtualServiceGVK.Version,
		VirtualServiceGVK.Kind,
		"vs",
		false,
		&VirtualService{})
)

func init() {
	if err := crd.AddCrd(VirtualServiceCrd); err != nil {
		log.Fatalf("could not add crd to global registry")
	}
}

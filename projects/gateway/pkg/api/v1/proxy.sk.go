// Code generated by solo-kit. DO NOT EDIT.

package v1

import (
	"sort"

	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewProxy(namespace, name string) *Proxy {
	proxy := &Proxy{}
	proxy.SetMetadata(core.Metadata{
		Name:      name,
		Namespace: namespace,
	})
	return proxy
}

func (r *Proxy) SetMetadata(meta core.Metadata) {
	r.Metadata = meta
}

func (r *Proxy) SetStatus(status core.Status) {
	r.Status = status
}

func (r *Proxy) Hash() uint64 {
	metaCopy := r.GetMetadata()
	metaCopy.ResourceVersion = ""
	return hashutils.HashAll(
		metaCopy,
		r.Listeners,
	)
}

type ProxyList []*Proxy

// namespace is optional, if left empty, names can collide if the list contains more than one with the same name
func (list ProxyList) Find(namespace, name string) (*Proxy, error) {
	for _, proxy := range list {
		if proxy.GetMetadata().Name == name {
			if namespace == "" || proxy.GetMetadata().Namespace == namespace {
				return proxy, nil
			}
		}
	}
	return nil, errors.Errorf("list did not find proxy %v.%v", namespace, name)
}

func (list ProxyList) AsResources() resources.ResourceList {
	var ress resources.ResourceList
	for _, proxy := range list {
		ress = append(ress, proxy)
	}
	return ress
}

func (list ProxyList) AsInputResources() resources.InputResourceList {
	var ress resources.InputResourceList
	for _, proxy := range list {
		ress = append(ress, proxy)
	}
	return ress
}

func (list ProxyList) Names() []string {
	var names []string
	for _, proxy := range list {
		names = append(names, proxy.GetMetadata().Name)
	}
	return names
}

func (list ProxyList) NamespacesDotNames() []string {
	var names []string
	for _, proxy := range list {
		names = append(names, proxy.GetMetadata().Namespace+"."+proxy.GetMetadata().Name)
	}
	return names
}

func (list ProxyList) Sort() ProxyList {
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].GetMetadata().Less(list[j].GetMetadata())
	})
	return list
}

func (list ProxyList) Clone() ProxyList {
	var proxyList ProxyList
	for _, proxy := range list {
		proxyList = append(proxyList, resources.Clone(proxy).(*Proxy))
	}
	return proxyList
}

func (list ProxyList) Each(f func(element *Proxy)) {
	for _, proxy := range list {
		f(proxy)
	}
}

func (list ProxyList) EachResource(f func(element resources.Resource)) {
	for _, proxy := range list {
		f(proxy)
	}
}

func (list ProxyList) AsInterfaces() []interface{} {
	var asInterfaces []interface{}
	list.Each(func(element *Proxy) {
		asInterfaces = append(asInterfaces, element)
	})
	return asInterfaces
}

var _ resources.Resource = &Proxy{}

// Kubernetes Adapter for Proxy

func (o *Proxy) GetObjectKind() schema.ObjectKind {
	t := ProxyCrd.TypeMeta()
	return &t
}

func (o *Proxy) DeepCopyObject() runtime.Object {
	return resources.Clone(o).(*Proxy)
}

var ProxyCrd = crd.NewCrd("gateway.solo.io",
	"proxies",
	"gateway.solo.io",
	"v1",
	"Proxy",
	"px",
	false,
	&Proxy{})

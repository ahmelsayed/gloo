package extauth

import (
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoyroute "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoyauth "github.com/envoyproxy/go-control-plane/envoy/config/filter/http/ext_authz/v2"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	extauthv1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1/enterprise/plugins/extauth/v1"
	"github.com/solo-io/gloo/projects/gloo/pkg/plugins"
	"github.com/solo-io/gloo/projects/gloo/pkg/plugins/pluginutils"
)

const (
	FilterName        = "envoy.ext_authz"
	DefaultAuthHeader = "x-user-id"
	HttpServerUri     = "http://not-used.example.com/"
)

// Note that although this configures the "envoy.ext_authz" filter, we still want the ordering to be within the
// AuthNStage because we are using this filter for authentication purposes
var filterStage = plugins.DuringStage(plugins.AuthNStage)

var _ plugins.Plugin = &Plugin{}

func NewCustomAuthPlugin() *Plugin {
	return &Plugin{}
}

type Plugin struct {
	extAuthSettings *extauthv1.Settings
}

func (p *Plugin) Init(params plugins.InitParams) error {
	p.extAuthSettings = params.Settings.GetExtauth()
	return nil
}

func (p *Plugin) HttpFilters(params plugins.Params, _ *v1.HttpListener) ([]plugins.StagedHttpFilter, error) {
	// Delegate to a function with a simpler signature,will make it easier to reuse
	return BuildHttpFilters(p.extAuthSettings, params.Snapshot.Upstreams)
}

// This function generates the ext_authz PerFilterConfig for this virtual host. If the ext_authz filter was not
// configured on the listener, do nothing. If the filter is configured and the virtual host does not define
// an extauth configuration OR explicitly disables extauth, we disable the ext_authz filter.
// This is done to disable authentication by default on a virtual host and its child resources (routes, weighted
// destinations). Extauth is currently opt-in.
func (p *Plugin) ProcessVirtualHost(params plugins.VirtualHostParams, in *v1.VirtualHost, out *route.VirtualHost) error {

	// Ext_authz filter is not configured on listener, do nothing
	if !p.isExtAuthzFilterConfigured(params.Snapshot.Upstreams) {
		return nil
	}

	// If extauth is explicitly disabled on this virtual host, disable it
	if in.GetVirtualHostPlugins().GetExtauth().GetDisable() {
		return markVirtualHostNoAuth(out)
	}

	customAuthConfig := in.GetVirtualHostPlugins().GetExtauth().GetCustomAuth()

	// No extauth config on this virtual host, disable it
	if customAuthConfig == nil {
		return markVirtualHostNoAuth(out)
	}

	config := &envoyauth.ExtAuthzPerRoute{
		Override: &envoyauth.ExtAuthzPerRoute_CheckSettings{
			CheckSettings: &envoyauth.CheckSettings{
				ContextExtensions: customAuthConfig.GetContextExtensions(),
			},
		},
	}

	return pluginutils.SetVhostPerFilterConfig(out, FilterName, config)
}

// This function generates the ext_authz PerFilterConfig for this route:
// - if the route defines custom auth configuration, set the filter correspondingly;
// - if auth is explicitly disabled, disable the filter (will apply by default also to WeightedDestinations);
// - else, do nothing (will inherit config from parent virtual host).
func (p *Plugin) ProcessRoute(params plugins.RouteParams, in *v1.Route, out *route.Route) error {

	// Ext_authz is not configured, do nothing
	if !p.isExtAuthzFilterConfigured(params.Snapshot.Upstreams) {
		return nil
	}

	// Extauth is explicitly disabled, disable it on route
	if in.GetRoutePlugins().GetExtauth().GetDisable() {
		return markRouteNoAuth(out)
	}

	customAuthConfig := in.GetRoutePlugins().GetExtauth().GetCustomAuth()

	// No custom config, do nothing
	if customAuthConfig == nil {
		return nil
	}

	config := &envoyauth.ExtAuthzPerRoute{
		Override: &envoyauth.ExtAuthzPerRoute_CheckSettings{
			CheckSettings: &envoyauth.CheckSettings{
				ContextExtensions: customAuthConfig.GetContextExtensions(),
			},
		},
	}

	return pluginutils.SetRoutePerFilterConfig(out, FilterName, config)
}

// This function generates the ext_authz PerFilterConfig for this weightedDestination:
// - if the weightedDestination defines custom auth configuration, set the filter correspondingly;
// - if auth is explicitly disabled, disable the filter;
// - else, do nothing (will inherit config from parent virtual host and/or route).
func (p *Plugin) ProcessWeightedDestination(params plugins.RouteParams, in *v1.WeightedDestination, out *route.WeightedCluster_ClusterWeight) error {

	// Ext_authz is not configured, do nothing
	if !p.isExtAuthzFilterConfigured(params.Snapshot.Upstreams) {
		return nil
	}

	// Extauth is explicitly disabled, disable it on weighted destination
	if in.GetWeightedDestinationPlugins().GetExtauth().GetDisable() {
		return markWeightedClusterNoAuth(out)
	}

	customAuthConfig := in.GetWeightedDestinationPlugins().GetExtauth().GetCustomAuth()

	// No custom config, do nothing
	if customAuthConfig == nil {
		return nil
	}

	config := &envoyauth.ExtAuthzPerRoute{
		Override: &envoyauth.ExtAuthzPerRoute_CheckSettings{
			CheckSettings: &envoyauth.CheckSettings{
				ContextExtensions: customAuthConfig.GetContextExtensions(),
			},
		},
	}

	return pluginutils.SetWeightedClusterPerFilterConfig(out, FilterName, config)
}

func (p *Plugin) isExtAuthzFilterConfigured(upstreams v1.UpstreamList) bool {
	// Call the same function called by HttpFilters to verify whether the filter was created
	filters, err := BuildHttpFilters(p.extAuthSettings, upstreams)
	if err != nil {
		// If it returned an error, the filter was not configured
		return false
	}

	// Check for a filter called "envoy.ext_authz"
	for _, filter := range filters {
		if filter.HttpFilter.GetName() == FilterName {
			return true
		}
	}

	return false
}

func markVirtualHostNoAuth(out *envoyroute.VirtualHost) error {
	return pluginutils.SetVhostPerFilterConfig(out, FilterName, getNoAuthConfig())
}

func markWeightedClusterNoAuth(out *envoyroute.WeightedCluster_ClusterWeight) error {
	return pluginutils.SetWeightedClusterPerFilterConfig(out, FilterName, getNoAuthConfig())
}

func markRouteNoAuth(out *envoyroute.Route) error {
	return pluginutils.SetRoutePerFilterConfig(out, FilterName, getNoAuthConfig())
}

func getNoAuthConfig() *envoyauth.ExtAuthzPerRoute {
	return &envoyauth.ExtAuthzPerRoute{
		Override: &envoyauth.ExtAuthzPerRoute_Disabled{
			Disabled: true,
		},
	}
}

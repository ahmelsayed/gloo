package printers

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/solo-io/gloo/projects/gloo/cli/pkg/xdsinspection"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/plugins"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/plugins/aws/ec2"
	"github.com/solo-io/go-utils/cliutils"

	"github.com/olekukonko/tablewriter"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
)

func PrintUpstreams(upstreams v1.UpstreamList, outputType OutputType, xdsDump *xdsinspection.XdsDump) error {
	if outputType == KUBE_YAML {
		return PrintKubeCrdList(upstreams.AsInputResources(), v1.UpstreamCrd)
	}
	return cliutils.PrintList(outputType.String(), "", upstreams,
		func(data interface{}, w io.Writer) error {
			UpstreamTable(xdsDump, data.(v1.UpstreamList), w)
			return nil
		}, os.Stdout)
}

// PrintTable prints upstreams using tables to io.Writer
func UpstreamTable(xdsDump *xdsinspection.XdsDump, upstreams []*v1.Upstream, w io.Writer) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Upstream", "type", "status", "details"})

	for _, us := range upstreams {
		name := us.GetMetadata().Name
		s := us.Status.State.String()

		u := upstreamType(us)
		details := upstreamDetails(us, xdsDump)

		if len(details) == 0 {
			details = []string{""}
		}
		for i, line := range details {
			if i == 0 {
				table.Append([]string{name, u, s, line})
			} else {
				table.Append([]string{"", "", "", line})
			}
		}

	}

	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
}

func upstreamType(up *v1.Upstream) string {
	if up.UpstreamSpec == nil {
		return "Invalid"
	}

	switch up.UpstreamSpec.UpstreamType.(type) {
	case *v1.UpstreamSpec_Aws:
		return "AWS Lambda"
	case *v1.UpstreamSpec_Azure:
		return "Azure"
	case *v1.UpstreamSpec_Consul:
		return "Consul"
	case *v1.UpstreamSpec_AwsEc2:
		return "AWS EC2"
	case *v1.UpstreamSpec_Kube:
		return "Kubernetes"
	case *v1.UpstreamSpec_Static:
		return "Static"
	default:
		return "Unknown"
	}
}

func upstreamDetails(up *v1.Upstream, xdsDump *xdsinspection.XdsDump) []string {
	if up.UpstreamSpec == nil {
		return []string{"invalid: spec was nil"}
	}

	var details []string
	add := func(s ...string) {
		details = append(details, s...)
	}
	switch usType := up.UpstreamSpec.UpstreamType.(type) {
	case *v1.UpstreamSpec_Aws:
		var functions []string
		for _, fn := range usType.Aws.LambdaFunctions {
			functions = append(functions, fn.LambdaFunctionName)
		}
		add(
			fmt.Sprintf("region: %v", usType.Aws.Region),
			fmt.Sprintf("secret: %v", usType.Aws.SecretRef.Key()),
		)
		for i := range functions {
			if i == 0 {
				add("functions:")
			}
			add(fmt.Sprintf("- %v", functions[i]))
		}
	case *v1.UpstreamSpec_AwsEc2:
		add(
			fmt.Sprintf("role:           %v", usType.AwsEc2.RoleArn),
			fmt.Sprintf("uses public ip: %v", usType.AwsEc2.PublicIp),
			fmt.Sprintf("port:           %v", usType.AwsEc2.Port),
		)
		add(getEc2TagFiltersString(usType.AwsEc2.Filters)...)
		instances := xdsDump.GetEc2InstancesForUpstream(up.Metadata.Ref())
		add(
			"EC2 Instance Ids:",
		)
		add(
			instances...,
		)
	case *v1.UpstreamSpec_Azure:
		var functions []string
		for _, fn := range usType.Azure.Functions {
			functions = append(functions, fn.FunctionName)
		}
		add(
			fmt.Sprintf("function app name: %v", usType.Azure.FunctionAppName),
			fmt.Sprintf("secret: %v", usType.Azure.SecretRef.Key()),
		)

		for i := range functions {
			if i == 0 {
				add("functions:")
			}
			add(fmt.Sprintf("- %v", functions[i]))
		}
	case *v1.UpstreamSpec_Consul:
		add(
			fmt.Sprintf("svc name: %v", usType.Consul.ServiceName),
			fmt.Sprintf("svc tags: %v", usType.Consul.ServiceTags),
		)
		if usType.Consul.ServiceSpec != nil {
			add(linesForServiceSpec(usType.Consul.ServiceSpec)...)
		}
	case *v1.UpstreamSpec_Kube:
		add(
			fmt.Sprintf("svc name:      %v", usType.Kube.ServiceName),
			fmt.Sprintf("svc namespace: %v", usType.Kube.ServiceNamespace),
			fmt.Sprintf("port:          %v", usType.Kube.ServicePort),
		)
		if usType.Kube.ServiceSpec != nil {
			add(linesForServiceSpec(usType.Kube.ServiceSpec)...)
		}
	case *v1.UpstreamSpec_Static:
		for i := range usType.Static.Hosts {
			if i == 0 {
				add("hosts:")
			}
			add(fmt.Sprintf("- %v:%v", usType.Static.Hosts[i].Addr, usType.Static.Hosts[i].Port))
		}
		if usType.Static.ServiceSpec != nil {
			add(linesForServiceSpec(usType.Static.ServiceSpec)...)
		}
	}
	add("")
	return details
}

func linesForServiceSpec(serviceSpec *plugins.ServiceSpec) []string {
	var spec []string
	add := func(s ...string) {
		spec = append(spec, s...)
	}
	switch plug := serviceSpec.PluginType.(type) {
	case *plugins.ServiceSpec_Rest:
		add("REST service:")
		var functions []string
		for restFunc, transform := range plug.Rest.Transformations {
			var method, path string
			methodP, ok := transform.Headers[":method"]
			if ok {
				method = methodP.Text
			}
			pathP, ok := transform.Headers[":path"]
			if ok {
				path = pathP.Text
			}
			if false {
				//TODO ilackarms: save it for -o wide
				functions = append(functions, fmt.Sprintf("- %v (%v %v)", restFunc, method, path))
			}
			functions = append(functions, fmt.Sprintf("- %v", restFunc))
		}
		// needed because map
		sort.Strings(functions)

		for i := range functions {
			if i == 0 {
				add("functions:")
			}
			add(functions[i])
		}
	case *plugins.ServiceSpec_Grpc:
		add("gRPC service:")
		for _, grpcService := range plug.Grpc.GrpcServices {
			add(fmt.Sprintf("  %v", grpcService.ServiceName))
			for _, fn := range grpcService.FunctionNames {
				add(fmt.Sprintf("  - %v", fn))
			}
		}
	}

	return spec
}

func getEc2TagFiltersString(filters []*ec2.TagFilter) []string {
	var out []string
	add := func(s ...string) {
		out = append(out, s...)
	}

	var kFilters []*ec2.TagFilter_Key
	var kvFilters []*ec2.TagFilter_KvPair
	for _, f := range filters {
		switch x := f.Spec.(type) {
		case *ec2.TagFilter_Key:
			kFilters = append(kFilters, x)
		case *ec2.TagFilter_KvPair_:
			kvFilters = append(kvFilters, x.KvPair)
		}
	}
	if len(kFilters) == 0 {
		add(fmt.Sprintf("key filters: (none)"))
	} else {
		add(fmt.Sprintf("key filters:"))
		for _, f := range kFilters {
			add(fmt.Sprintf("- %v", f.Key))
		}
	}
	if len(kvFilters) == 0 {
		add(fmt.Sprintf("key-value filters: (none)"))
	} else {
		add(fmt.Sprintf("key-value filters:"))
		for _, f := range kvFilters {
			add(fmt.Sprintf("- %v: %v", f.Key, f.Value))
		}
	}
	return out
}

package outputs

import (
	mongodbcontextconfig "github.com/plantoncloud/mongodb-cluster-pulumi-blueprint/pkg/kubernetes/contextconfig"
	pulumicommonsloadbalancerservice "github.com/plantoncloud/pulumi-blueprint-commons/pkg/kubernetes/loadbalancer/service"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	rootUsername = "root"
	MongoDbPort  = 27017
)

type input struct {
	ResourceId                    string
	ResourceName                  string
	EnvironmentName               string
	EndpointDomainName            string
	NamespaceName                 string
	ExternalLoadBalancerIpAddress string
	InternalLoadBalancerIpAddress string
	InternalHostname              string
	ExternalHostname              string
	KubeServiceName               string
	KubeLocalEndpoint             string
}

func extractInput(ctx *pulumi.Context) *input {
	var ctxConfig = ctx.Value(mongodbcontextconfig.Key).(mongodbcontextconfig.ContextConfig)
	var externalLoadBalancerIpAddress = ""
	var internalLoadBalancerIpAddress = ""

	if ctxConfig.Status.AddedResources.LoadBalancerExternalService != nil {
		externalLoadBalancerIpAddress = pulumicommonsloadbalancerservice.GetIpAddress(ctxConfig.Status.AddedResources.LoadBalancerExternalService)
	}

	if ctxConfig.Status.AddedResources.LoadBalancerInternalService != nil {
		internalLoadBalancerIpAddress = pulumicommonsloadbalancerservice.GetIpAddress(ctxConfig.Status.AddedResources.LoadBalancerExternalService)
	}

	return &input{
		ResourceId:                    ctxConfig.Spec.ResourceId,
		ResourceName:                  ctxConfig.Spec.ResourceName,
		EnvironmentName:               ctxConfig.Spec.EnvironmentInfo.EnvironmentName,
		EndpointDomainName:            ctxConfig.Spec.EndpointDomainName,
		NamespaceName:                 ctxConfig.Spec.NamespaceName,
		ExternalLoadBalancerIpAddress: externalLoadBalancerIpAddress,
		InternalLoadBalancerIpAddress: internalLoadBalancerIpAddress,
		InternalHostname:              ctxConfig.Spec.InternalHostname,
		ExternalHostname:              ctxConfig.Spec.ExternalHostname,
		KubeServiceName:               ctxConfig.Spec.KubeServiceName,
		KubeLocalEndpoint:             ctxConfig.Spec.KubeLocalEndpoint,
	}
}

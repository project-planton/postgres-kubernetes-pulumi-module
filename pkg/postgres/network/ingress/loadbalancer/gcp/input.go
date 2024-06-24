package gcp

import (
	postgresdbcontextconfig "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/contextconfig"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type input struct {
	ResourceId         string
	ResourceName       string
	Namespace          *kubernetescorev1.Namespace
	ExternalEndpoint   string
	InternalEndpoint   string
	EndpointDomainName string
	ServiceName        string
}

func extractInput(ctx *pulumi.Context) *input {
	var ctxConfig = ctx.Value(postgresdbcontextconfig.Key).(postgresdbcontextconfig.ContextConfig)

	return &input{
		ResourceId:         ctxConfig.Spec.ResourceId,
		ResourceName:       ctxConfig.Spec.ResourceName,
		Namespace:          ctxConfig.Status.AddedResources.Namespace,
		ExternalEndpoint:   ctxConfig.Spec.ExternalHostname,
		InternalEndpoint:   ctxConfig.Spec.InternalHostname,
		EndpointDomainName: ctxConfig.Spec.EndpointDomainName,
		ServiceName:        ctxConfig.Spec.KubeServiceName,
	}
}

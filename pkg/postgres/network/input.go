package network

import (
	postgresdbcontextconfig "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/contextstate"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type input struct {
	IsIngressEnabled   bool
	EndpointDomainName string
}

func extractInput(ctx *pulumi.Context) *input {
	var ctxConfig = ctx.Value(postgresdbcontextconfig.Key).(postgresdbcontextconfig.ContextState)

	return &input{
		IsIngressEnabled:   ctxConfig.Spec.IsIngressEnabled,
		EndpointDomainName: ctxConfig.Spec.EndpointDomainName,
	}
}

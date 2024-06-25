package stunnel

import (
	postgresdbcontextconfig "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/contextstate"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type input struct {
	Namespace *kubernetescorev1.Namespace
}

func extractInput(ctx *pulumi.Context) *input {
	var ctxConfig = ctx.Value(postgresdbcontextconfig.Key).(postgresdbcontextconfig.ContextState)

	return &input{
		Namespace: ctxConfig.Status.AddedResources.Namespace,
	}
}

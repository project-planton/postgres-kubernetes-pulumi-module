package virtualservice

import (
	postgresdbcontextconfig "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/contextstate"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type input struct {
	namespaceName    string
	workspaceDir     string
	namespace        *kubernetescorev1.Namespace
	externalHostname string
	internalHostname string
}

func extractInput(ctx *pulumi.Context) *input {
	var ctxConfig = ctx.Value(postgresdbcontextconfig.Key).(postgresdbcontextconfig.ContextState)

	return &input{
		namespaceName:    ctxConfig.Spec.NamespaceName,
		workspaceDir:     ctxConfig.Spec.WorkspaceDir,
		namespace:        ctxConfig.Status.AddedResources.Namespace,
		externalHostname: ctxConfig.Spec.ExternalHostname,
		internalHostname: ctxConfig.Spec.InternalHostname,
	}
}

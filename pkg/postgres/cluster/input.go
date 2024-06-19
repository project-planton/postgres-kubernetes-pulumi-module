package cluster

import (
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubecluster/enums/kubernetesworkloadingresstype"
	plantoncloudpostgresdbmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/postgreskubernetes/model"
	postgresdbcontextconfig "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/contextconfig"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type input struct {
	WorkspaceDir     string
	ResourceId       string
	Namespace        *kubernetescorev1.Namespace
	NamespaceName    string
	KubeServiceName  string
	ContainerSpec    *plantoncloudpostgresdbmodel.PostgresKubernetesSpecContainerSpec
	IngressType      kubernetesworkloadingresstype.KubernetesWorkloadIngressType
	IsIngressEnabled bool
	Labels           map[string]string
}

func extractInput(ctx *pulumi.Context) *input {
	var ctxConfig = ctx.Value(postgresdbcontextconfig.Key).(postgresdbcontextconfig.ContextConfig)

	return &input{
		WorkspaceDir:     ctxConfig.Spec.WorkspaceDir,
		ResourceId:       ctxConfig.Spec.ResourceId,
		Namespace:        ctxConfig.Status.AddedResources.Namespace,
		NamespaceName:    ctxConfig.Spec.NamespaceName,
		ContainerSpec:    ctxConfig.Spec.ContainerSpec,
		IngressType:      ctxConfig.Spec.IngressType,
		IsIngressEnabled: ctxConfig.Spec.IsIngressEnabled,
		Labels:           ctxConfig.Spec.Labels,
		KubeServiceName:  ctxConfig.Spec.KubeServiceName,
	}
}

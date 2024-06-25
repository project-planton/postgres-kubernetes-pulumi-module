package cluster

import (
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubecluster/enums/kubernetesworkloadingresstype"
	plantoncloudpostgresdbmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/postgreskubernetes/model"
	plantoncommonsapiresourcemodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/apiresource/model"
	postgresdbcontextconfig "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/contextstate"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type input struct {
	WorkspaceDir     string
	ResourceId       string
	Metadata         *plantoncommonsapiresourcemodel.ApiResourceMetadata
	Namespace        *kubernetescorev1.Namespace
	NamespaceName    string
	ContainerSpec    *plantoncloudpostgresdbmodel.PostgresKubernetesSpecContainerSpec
	IngressType      kubernetesworkloadingresstype.KubernetesWorkloadIngressType
	IsIngressEnabled bool
	Labels           map[string]string
}

func extractInput(ctx *pulumi.Context) *input {
	var ctxConfig = ctx.Value(postgresdbcontextconfig.Key).(postgresdbcontextconfig.ContextState)

	return &input{
		WorkspaceDir:     ctxConfig.Spec.WorkspaceDir,
		Namespace:        ctxConfig.Status.AddedResources.Namespace,
		ContainerSpec:    ctxConfig.Spec.ContainerSpec,
		IngressType:      ctxConfig.Spec.IngressType,
		IsIngressEnabled: ctxConfig.Spec.IsIngressEnabled,
		Labels:           ctxConfig.Spec.Labels,
		NamespaceName:    ctxConfig.Spec.NamespaceName,
	}
}

package network

import (
	"github.com/pkg/errors"
	code2cloudv1deploypgcstackk8smodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/postgreskubernetes/stack/kubernetes/model"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/hostname"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress"
	pulk8scv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Input struct {
	WorkspaceDir       string
	NamespaceName      string
	AddedNamespace     *pulk8scv1.Namespace
	StackResourceInput *code2cloudv1deploypgcstackk8smodel.PostgresClusterKubernetesStackResourceInput
	Labels             map[string]string
}

func Resources(ctx *pulumi.Context, input *Input) error {
	if err := hostname.Resources(ctx, &hostname.Input{
		PostgresCluster: input.StackResourceInput.PostgresCluster,
	}); err != nil {
		return errors.Wrap(err, "failed to add hostname resources")
	}

	if input.StackResourceInput.PostgresCluster.Spec.Kubernetes.Ingress == nil ||
		!input.StackResourceInput.PostgresCluster.Spec.Kubernetes.Ingress.IsEnabled {
		return nil
	}

	if err := ingress.Resources(ctx, &ingress.Input{
		WorkspaceDir:       input.WorkspaceDir,
		NamespaceName:      input.NamespaceName,
		AddedNamespace:     input.AddedNamespace,
		StackResourceInput: input.StackResourceInput,
		Labels:             input.Labels,
		PostgresClusterId:  input.StackResourceInput.PostgresCluster.Metadata.Id,
		EnvironmentInfo:    input.StackResourceInput.PostgresCluster.Spec.EnvironmentInfo,
		EndpointDomainName: input.StackResourceInput.PostgresCluster.Spec.Kubernetes.Ingress.EndpointDomainName,
	}); err != nil {
		return errors.Wrap(err, "failed to add ingress resources")
	}
	return nil
}

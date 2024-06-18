package istio

import (
	"github.com/pkg/errors"
	code2cloudv1deploypgcstackk8smodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/postgreskubernetes/stack/kubernetes/model"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/hostname"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress/istio/cert"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress/istio/stunnel"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress/istio/virtualservice"
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
	if err := stunnel.Resources(ctx, &stunnel.Input{
		Namespace: input.AddedNamespace,
	}); err != nil {
		return errors.Wrap(err, "failed to add service resources")
	}

	externalEndpointHostname := hostname.GetExternalClusterHostname(
		input.StackResourceInput.PostgresCluster.Metadata.Id,
		input.StackResourceInput.PostgresCluster.Spec.EnvironmentInfo.EnvironmentName,
		input.StackResourceInput.PostgresCluster.Spec.Kubernetes.Ingress.EndpointDomainName)
	internalEndpointHostname := hostname.GetInternalClusterHostname(
		input.StackResourceInput.PostgresCluster.Metadata.Id,
		input.StackResourceInput.PostgresCluster.Spec.EnvironmentInfo.EnvironmentName,
		input.StackResourceInput.PostgresCluster.Spec.Kubernetes.Ingress.EndpointDomainName)

	err := cert.Resources(ctx, &cert.Input{
		Namespace:     input.AddedNamespace,
		NamespaceName: input.NamespaceName,
		WorkspaceDir:  input.WorkspaceDir,
		Hostnames:     []string{internalEndpointHostname, externalEndpointHostname},
	})

	if err != nil {
		return errors.Wrap(err, "failed to add gateway resources")
	}
	err = virtualservice.Resources(ctx, &virtualservice.Input{
		StackResourceInput: input.StackResourceInput,
		Labels:             input.Labels,
		NamespaceName:      input.NamespaceName,
		WorkspaceDir:       input.WorkspaceDir,
		Namespace:          input.AddedNamespace,
	})
	if err != nil {
		return errors.Wrap(err, "failed to add virtual service resources")
	}
	return nil
}

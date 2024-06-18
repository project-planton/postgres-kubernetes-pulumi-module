package postgres

import (
	"github.com/pkg/errors"
	code2cloudv1deploypgcstackk8smodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/postgreskubernetes/stack/kubernetes/model"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/cluster"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/namespace"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network"
	pulumikubernetesprovider "github.com/plantoncloud/pulumi-stack-runner-go-sdk/pkg/automation/provider/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ResourceStack struct {
	WorkspaceDir     string
	Input            *code2cloudv1deploypgcstackk8smodel.PostgresKubernetesStackInput
	KubernetesLabels map[string]string
}

func (s *ResourceStack) Resources(ctx *pulumi.Context) error {
	kubernetesProvider, err := pulumikubernetesprovider.GetWithStackCredentials(ctx, s.Input.CredentialsInput)
	if err != nil {
		return errors.Wrap(err, "failed to setup kubernetes provider")
	}

	namespaceName := s.Input.ResourceInput.Metadata.Id

	addedNamespace, err := namespace.Resources(ctx, &namespace.Input{
		KubernetesProvider: kubernetesProvider,
		NamespaceName:      namespaceName,
		Labels:             s.KubernetesLabels,
	})
	if err != nil {
		return errors.Wrap(err, "failed to add namespace resources")
	}

	if err := cluster.Resources(ctx, &cluster.Input{
		WorkspaceDir:                        s.WorkspaceDir,
		NamespaceName:                       namespaceName,
		Namespace:                           addedNamespace,
		PostgresClusterKubernetesStackInput: s.Input,
		Labels:                              s.KubernetesLabels,
	}); err != nil {
		return errors.Wrap(err, "failed to add cluster resources")
	}

	if err := network.Resources(ctx, &network.Input{
		WorkspaceDir:       s.WorkspaceDir,
		NamespaceName:      namespaceName,
		AddedNamespace:     addedNamespace,
		StackResourceInput: s.Input.ResourceInput,
		Labels:             s.KubernetesLabels,
	}); err != nil {
		return errors.Wrap(err, "failed to add network resources")
	}
	return nil
}

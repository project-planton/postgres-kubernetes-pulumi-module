package postgres

import (
	"github.com/pkg/errors"
	code2cloudv1deploypgcstackk8smodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/postgreskubernetes/stack/kubernetes/model"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/cluster"
	postgresdbcontextconfig "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/contextconfig"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network"
	commonsnamespaceresources "github.com/plantoncloud/pulumi-blueprint-commons/pkg/kubernetes/namespace"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ResourceStack struct {
	WorkspaceDir     string
	Input            *code2cloudv1deploypgcstackk8smodel.PostgresKubernetesStackInput
	KubernetesLabels map[string]string
}

func (resourceStack *ResourceStack) Resources(ctx *pulumi.Context) error {
	//load context config
	var ctxConfig, err = loadConfig(ctx, resourceStack)
	if err != nil {
		return errors.Wrap(err, "failed to initiate context config")
	}
	ctx = ctx.WithValue(postgresdbcontextconfig.Key, *ctxConfig)

	// Create the namespace resource
	namespace, err := commonsnamespaceresources.Resources(ctx, resourceStack.Input.ProtoReflect())
	if err != nil {
		return errors.Wrap(err, "failed to create namespace resource")
	}

	if err := cluster.Resources(ctx, &cluster.Input{
		WorkspaceDir:                        resourceStack.WorkspaceDir,
		NamespaceName:                       namespaceName,
		Namespace:                           addedNamespace,
		PostgresClusterKubernetesStackInput: resourceStack.Input,
		Labels:                              resourceStack.KubernetesLabels,
	}); err != nil {
		return errors.Wrap(err, "failed to add cluster resources")
	}

	if err := network.Resources(ctx, &network.Input{
		WorkspaceDir:       resourceStack.WorkspaceDir,
		NamespaceName:      namespaceName,
		AddedNamespace:     addedNamespace,
		StackResourceInput: resourceStack.Input.ResourceInput,
		Labels:             resourceStack.KubernetesLabels,
	}); err != nil {
		return errors.Wrap(err, "failed to add network resources")
	}
	return nil
}

package postgres

import (
	"github.com/pkg/errors"
	code2cloudv1deploypgcstackk8smodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/postgreskubernetes/stack/model"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/cluster"
	postgresdbcontextconfig "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/contextstate"
	postgreskubernetesnamespace "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/namespace"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network"
	postgresblueprintoutputs "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/outputs"
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
	ctx, err = postgreskubernetesnamespace.Resources(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace resource")
	}

	if err := cluster.Resources(ctx); err != nil {
		return errors.Wrap(err, "failed to add cluster resources")
	}

	ctx, err = network.Resources(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to add network resources")
	}

	err = postgresblueprintoutputs.Export(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to export postgres kubernetes outputs")
	}

	return nil
}

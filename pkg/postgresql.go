package pkg

import (
	"github.com/pkg/errors"
	zalandov1 "github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/zalandooperator/acid/v1"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/locals"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func postgresql(ctx *pulumi.Context,
	createdNamespace *kubernetescorev1.Namespace, labels map[string]string) error {

	//create zalando postgresql resource
	_, err := zalandov1.NewPostgresql(ctx,
		"database",
		&zalandov1.PostgresqlArgs{
			Metadata: metav1.ObjectMetaArgs{
				// for zolando operator the name is required to be always prefixed by teamId
				// a kubernetes service with the same name is created by the operator
				Name:      pulumi.Sprintf("%s-%s", vars.TeamId, locals.PostgresKubernetes.Metadata.Id),
				Namespace: createdNamespace.Metadata.Name(),
				Labels:    pulumi.ToStringMap(labels),
			},
			Spec: zalandov1.PostgresqlSpecArgs{
				NumberOfInstances: pulumi.Int(locals.PostgresKubernetes.Spec.Container.Replicas),
				Patroni:           zalandov1.PostgresqlSpecPatroniArgs{},
				PodAnnotations: pulumi.ToStringMap(map[string]string{
					"postgres-cluster-id": locals.PostgresKubernetes.Metadata.Id,
				}),
				Postgresql: zalandov1.PostgresqlSpecPostgresqlArgs{
					Version: pulumi.String(vars.PostgresVersion),
				},
				Resources: zalandov1.PostgresqlSpecResourcesArgs{
					Limits: zalandov1.PostgresqlSpecResourcesLimitsArgs{
						Cpu:    pulumi.String(locals.PostgresKubernetes.Spec.Container.Resources.Limits.Cpu),
						Memory: pulumi.String(locals.PostgresKubernetes.Spec.Container.Resources.Limits.Memory),
					},
					Requests: zalandov1.PostgresqlSpecResourcesRequestsArgs{
						Cpu:    pulumi.String(locals.PostgresKubernetes.Spec.Container.Resources.Requests.Cpu),
						Memory: pulumi.String(locals.PostgresKubernetes.Spec.Container.Resources.Requests.Memory),
					},
				},
				TeamId: pulumi.String(vars.TeamId),
				Volume: zalandov1.PostgresqlSpecVolumeArgs{
					Size: pulumi.String(locals.PostgresKubernetes.Spec.Container.DiskSize),
				},
			},
		})
	if err != nil {
		return errors.Wrap(err, "failed to create postgresql")
	}
	return nil
}

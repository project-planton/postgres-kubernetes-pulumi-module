package pkg

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/rediskubernetes/model"
	"github.com/plantoncloud/pulumi-module-golang-commons/pkg/provider/kubernetes/containerresources"
	"github.com/plantoncloud/pulumi-module-golang-commons/pkg/provider/kubernetes/helm/convertmaps"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (s *ResourceStack) helmChart(ctx *pulumi.Context,
	createdNamespace *kubernetescorev1.Namespace) error {

	redisKubernetes := s.Input.ApiResource

	helmValues := getHelmValues(redisKubernetes, s.Labels)

	//install helm-chart
	_, err := helmv3.NewChart(ctx,
		redisKubernetes.Metadata.Id,
		helmv3.ChartArgs{
			Chart:     pulumi.String("redis"),
			Version:   pulumi.String("17.10.1"),
			Namespace: createdNamespace.Metadata.Name().Elem(),
			Values:    helmValues,
			//if you need to add the repository, you can specify `repo url`:
			FetchArgs: helmv3.FetchArgs{
				Repo: pulumi.String("https://charts.bitnami.com/bitnami"),
			},
		}, pulumi.Parent(createdNamespace))
	if err != nil {
		return errors.Wrap(err, "failed to create helm chart")
	}

	return nil
}

func getHelmValues(redisKubernetes *model.RedisKubernetes, labels map[string]string) pulumi.Map {
	// HelmVal https://github.com/bitnami/charts/blob/main/bitnami/redis/values.yaml
	return pulumi.Map{
		"fullnameOverride": pulumi.String(redisKubernetes.Metadata.Name),
		"architecture":     pulumi.String("standalone"),
		"master": pulumi.Map{
			"podLabels": convertmaps.ConvertGoMapToPulumiMap(labels),
			"resources": containerresources.ConvertToPulumiMap(redisKubernetes.Spec.Container.Resources),
			"persistence": pulumi.Map{
				"enabled": pulumi.Bool(redisKubernetes.Spec.Container.IsPersistenceEnabled),
				"size":    pulumi.String(redisKubernetes.Spec.Container.DiskSize),
			},
		},
		"replica": pulumi.Map{
			"podLabels":    convertmaps.ConvertGoMapToPulumiMap(labels),
			"replicaCount": pulumi.Int(redisKubernetes.Spec.Container.Replicas),
			"resources":    containerresources.ConvertToPulumiMap(redisKubernetes.Spec.Container.Resources),
			"persistence": pulumi.Map{
				"enabled": pulumi.Bool(redisKubernetes.Spec.Container.IsPersistenceEnabled),
				"size":    pulumi.String(redisKubernetes.Spec.Container.DiskSize),
			},
		},
		"auth": pulumi.Map{
			"existingSecret":            pulumi.String(redisKubernetes.Metadata.Name),
			"existingSecretPasswordKey": pulumi.String("redis-password"),
		},
	}
}

package loadbalancer

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/cloudaccount/enums/kubernetesprovider"
	code2cloudv1envmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/environment/model"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress/loadbalancer/gcp"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Input struct {
	PostgresClusterId  string
	EnvironmentInfo    *code2cloudv1envmodel.ApiResourceEnvironmentInfo
	AddedNamespace     *pulumikubernetescorev1.Namespace
	EndpointDomainName string
}

func Resources(ctx *pulumi.Context, input *Input) error {
	if input.EnvironmentInfo.KubernetesProvider == kubernetesprovider.KubernetesProvider_gcp_gke {
		if err := gcp.Resources(ctx, &gcp.Input{
			AddedNamespace:     input.AddedNamespace,
			PostgresClusterId:  input.PostgresClusterId,
			EnvironmentName:    input.EnvironmentInfo.EnvironmentName,
			EndpointDomainName: input.EndpointDomainName,
		}); err != nil {
			return errors.Wrap(err, "failed to create load balancer resources for gke cluster")
		}
	}
	return nil
}

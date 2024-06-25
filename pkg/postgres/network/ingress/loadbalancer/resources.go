package loadbalancer

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/cloudaccount/enums/kubernetesprovider"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress/loadbalancer/gcp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context) (*pulumi.Context, error) {
	i := extractInput(ctx)
	if i.kubeProvider == kubernetesprovider.KubernetesProvider_gcp_gke {
		newCtx, err := gcp.Resources(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create load balancer resources for gke cluster")
		}
		return newCtx, nil
	}
	return ctx, nil
}

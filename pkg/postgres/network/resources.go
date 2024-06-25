package network

import (
	"github.com/pkg/errors"
	postgresdbingress "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context) (newCtx *pulumi.Context, err error) {
	i := extractInput(ctx)
	if !i.IsIngressEnabled || i.EndpointDomainName == "" {
		return ctx, nil
	}
	if ctx, err = postgresdbingress.Resources(ctx); err != nil {
		return ctx, errors.Wrap(err, "failed to add gateway resources")
	}
	return ctx, nil
}

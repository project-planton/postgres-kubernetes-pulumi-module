package istio

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress/istio/cert"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress/istio/stunnel"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress/istio/virtualservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context) error {
	if err := stunnel.Resources(ctx); err != nil {
		return errors.Wrap(err, "failed to add stunnel service resources")
	}

	err := cert.Resources(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to add cert resources")
	}

	err = virtualservice.Resources(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to add virtual service resources")
	}
	return nil
}

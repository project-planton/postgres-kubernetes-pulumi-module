package ingress

import (
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubecluster/enums/kubernetesworkloadingresstype"
	postgresdbcontextconfig "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/contextstate"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type input struct {
	IngressType kubernetesworkloadingresstype.KubernetesWorkloadIngressType
}

func extractInput(ctx *pulumi.Context) *input {
	var ctxConfig = ctx.Value(postgresdbcontextconfig.Key).(postgresdbcontextconfig.ContextState)

	return &input{
		IngressType: ctxConfig.Spec.IngressType,
	}
}

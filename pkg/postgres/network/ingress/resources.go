package ingress

import (
	"github.com/pkg/errors"
	code2cloudv1envmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/environment/model"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubecluster/enums/kubernetesworkloadingresstype"
	code2cloudv1deploypgcstackk8smodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/postgreskubernetes/stack/kubernetes/model"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress/istio"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress/loadbalancer"
	pulk8scv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Input struct {
	WorkspaceDir       string
	NamespaceName      string
	AddedNamespace     *pulk8scv1.Namespace
	StackResourceInput *code2cloudv1deploypgcstackk8smodel.PostgresClusterKubernetesStackResourceInput
	Labels             map[string]string
	PostgresClusterId  string
	EnvironmentInfo    *code2cloudv1envmodel.ApiResourceEnvironmentInfo
	EndpointDomainName string
}

func Resources(ctx *pulumi.Context, input *Input) error {
	switch input.StackResourceInput.PostgresCluster.Spec.Kubernetes.Ingress.IngressType {
	case kubernetesworkloadingresstype.KubernetesWorkloadIngressType_load_balancer:
		if err := loadbalancer.Resources(ctx, &loadbalancer.Input{
			PostgresClusterId:  input.PostgresClusterId,
			EnvironmentInfo:    input.EnvironmentInfo,
			AddedNamespace:     input.AddedNamespace,
			EndpointDomainName: input.EndpointDomainName,
		}); err != nil {
			return errors.Wrap(err, "failed to add load balancer resources")
		}
	case kubernetesworkloadingresstype.KubernetesWorkloadIngressType_ingress_controller:
		if err := istio.Resources(ctx, &istio.Input{
			WorkspaceDir:       input.WorkspaceDir,
			NamespaceName:      input.NamespaceName,
			AddedNamespace:     input.AddedNamespace,
			StackResourceInput: input.StackResourceInput,
			Labels:             input.Labels,
		}); err != nil {
			return errors.Wrap(err, "failed to add ingress resources")
		}
	}
	return nil
}

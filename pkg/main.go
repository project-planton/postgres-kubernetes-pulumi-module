package pkg

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/rediskubernetes/model"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/kubernetes/enums/kubernetesworkloadingresstype"
	"github.com/plantoncloud/pulumi-module-golang-commons/pkg/provider/kubernetes/pulumikubernetesprovider"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ResourceStack struct {
	Input  *model.RedisKubernetesStackInput
	Labels map[string]string
}

func (s *ResourceStack) Resources(ctx *pulumi.Context) error {
	//create kubernetes-provider from the credential in the stack-input
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesClusterCredential(ctx,
		s.Input.KubernetesClusterCredential)
	if err != nil {
		return errors.Wrap(err, "failed to setup gcp provider")
	}

	//create a new descriptive variable for the api-resource in the input.
	redisKubernetes := s.Input.ApiResource

	//decide on the name of the namespace
	namespaceName := redisKubernetes.Metadata.Id

	//create namespace resource
	createdNamespace, err := kubernetescorev1.NewNamespace(ctx, namespaceName, &kubernetescorev1.NamespaceArgs{
		Metadata: metav1.ObjectMetaPtrInput(&metav1.ObjectMetaArgs{
			Name:   pulumi.String(namespaceName),
			Labels: pulumi.ToStringMap(s.Labels),
		}),
	}, pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "5s", Update: "5s", Delete: "5s"}),
		pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create %s namespace", namespaceName)
	}

	//export name of the namespace
	ctx.Export(NamespaceOutputName, createdNamespace.Metadata.Name())

	//install the redis helm-chart
	if err := s.helmChart(ctx, createdNamespace); err != nil {
		return errors.Wrap(err, "failed to create helm-chart resources")
	}

	//export kubernetes service name
	ctx.Export(ServiceOutputName, pulumi.Sprintf("%s-master", redisKubernetes.Metadata.Name))

	//export kubernetes endpoint
	ctx.Export(KubeEndpointOutputName,
		pulumi.Sprintf("%s-master.%s.svc.cluster.local.",
			redisKubernetes.Metadata.Name,
			namespaceName))

	//export kube-port-forward command
	ctx.Export(PortForwardCommandOutputName, pulumi.Sprintf(
		"kubectl port-forward -n %s service/%s 8080:8080",
		namespaceName,
		fmt.Sprintf("%s-master", redisKubernetes.Metadata.Name)))

	//no ingress resources required when ingress is not enabled
	if !redisKubernetes.Spec.Ingress.IsEnabled || redisKubernetes.Spec.Ingress.EndpointDomainName == "" {
		return nil
	}

	//depending on the ingress-type in the input, create either istio-ingress resources or
	//create load-balancer resources
	switch redisKubernetes.Spec.Ingress.IngressType {
	case kubernetesworkloadingresstype.KubernetesWorkloadIngressType_load_balancer:
		if err := s.loadBalancerIngress(ctx, createdNamespace); err != nil {
			return errors.Wrap(err, "failed to create load-balancer ingress resources")
		}
	case kubernetesworkloadingresstype.KubernetesWorkloadIngressType_ingress_controller:
		if err := s.istioIngress(ctx, createdNamespace); err != nil {
			return errors.Wrap(err, "failed to create istio ingress resources")
		}
	}

	//export ingress hostnames
	ctx.Export(IngressExternalHostnameOutputName, pulumi.Sprintf("%s.%s",
		redisKubernetes.Metadata.Id, redisKubernetes.Spec.Ingress.EndpointDomainName))
	ctx.Export(IngressInternalHostnameOutputName, pulumi.Sprintf("%s-internal.%s",
		redisKubernetes.Metadata.Id, redisKubernetes.Spec.Ingress.EndpointDomainName))

	return nil
}

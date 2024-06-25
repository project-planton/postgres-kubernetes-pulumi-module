package postgres

import (
	"github.com/pkg/errors"
	environmentblueprinthostnames "github.com/plantoncloud/environment-pulumi-blueprint/pkg/gcpgke/endpointdomains/hostnames"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubecluster/enums/kubernetesworkloadingresstype"
	postgresdbcontextconfig "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/contextstate"
	postgressingresscert "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress/istio/cert"
	postgresdbnetutilshostname "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress/netutils/hostname"
	postgresdbnetutilsservice "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress/netutils/service"
	plantoncloudpulumisdkkubernetes "github.com/plantoncloud/pulumi-stack-runner-go-sdk/pkg/automation/provider/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func loadConfig(ctx *pulumi.Context, resourceStack *ResourceStack) (*postgresdbcontextconfig.ContextState, error) {

	kubernetesProvider, err := plantoncloudpulumisdkkubernetes.GetWithStackCredentials(ctx, resourceStack.Input.CredentialsInput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup kubernetes provider")
	}

	var resourceId = resourceStack.Input.ResourceInput.Metadata.Id
	var resourceName = resourceStack.Input.ResourceInput.Metadata.Name
	var environmentInfo = resourceStack.Input.ResourceInput.Spec.EnvironmentInfo
	var isIngressEnabled = false

	if resourceStack.Input.ResourceInput.Spec.Ingress != nil {
		isIngressEnabled = resourceStack.Input.ResourceInput.Spec.Ingress.IsEnabled
	}

	var endpointDomainName = ""
	var envDomainName = ""
	var ingressType = kubernetesworkloadingresstype.KubernetesWorkloadIngressType_unspecified
	var internalHostname = ""
	var externalHostname = ""
	var certSecretName = ""

	if isIngressEnabled {
		endpointDomainName = resourceStack.Input.ResourceInput.Spec.Ingress.EndpointDomainName
		envDomainName = environmentblueprinthostnames.GetExternalEnvHostname(environmentInfo.EnvironmentName, endpointDomainName)
		ingressType = resourceStack.Input.ResourceInput.Spec.Ingress.IngressType

		internalHostname = postgresdbnetutilshostname.GetInternalHostname(resourceId, environmentInfo.EnvironmentName, endpointDomainName)
		externalHostname = postgresdbnetutilshostname.GetExternalHostname(resourceId, environmentInfo.EnvironmentName, endpointDomainName)
		certSecretName = postgressingresscert.GetCertSecretName(resourceName)
	}

	return &postgresdbcontextconfig.ContextState{
		Spec: &postgresdbcontextconfig.Spec{
			KubeProvider:       kubernetesProvider,
			ResourceId:         resourceId,
			ResourceName:       resourceName,
			ContainerSpec:      resourceStack.Input.ResourceInput.Spec.Container,
			Labels:             resourceStack.KubernetesLabels,
			WorkspaceDir:       resourceStack.WorkspaceDir,
			NamespaceName:      resourceId,
			EnvironmentInfo:    resourceStack.Input.ResourceInput.Spec.EnvironmentInfo,
			IsIngressEnabled:   isIngressEnabled,
			IngressType:        ingressType,
			EndpointDomainName: endpointDomainName,
			EnvDomainName:      envDomainName,
			InternalHostname:   internalHostname,
			ExternalHostname:   externalHostname,
			KubeServiceName:    postgresdbnetutilsservice.GetKubeServiceName(resourceName),
			KubeLocalEndpoint:  postgresdbnetutilsservice.GetKubeServiceNameFqdn(resourceName, resourceId),
			CertSecretName:     certSecretName,
		},
		Status: &postgresdbcontextconfig.Status{},
	}, nil
}

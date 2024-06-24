package gcp

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/english/enums/englishword"
	postgrescluster "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/cluster"
	postgrescontextconfig "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/contextconfig"
	postgresloadbalancercommon "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress/loadbalancer/common"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context) (*pulumi.Context, error) {
	// Create a Kubernetes Service of type LoadBalancer
	externalLoadBalancerService, err := addExternal(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add external load balancer")
	}
	internalLoadBalancerService, err := addInternal(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add internal load balancer")
	}

	var ctxConfig = ctx.Value(postgrescontextconfig.Key).(postgrescontextconfig.ContextConfig)

	addLoadBalancerExternalServiceToContext(&ctxConfig, externalLoadBalancerService)
	addLoadBalancerInternalServiceToContext(&ctxConfig, internalLoadBalancerService)
	ctx = ctx.WithValue(postgrescontextconfig.Key, ctxConfig)

	return ctx, nil
}

func addExternal(ctx *pulumi.Context) (*pulumikubernetescorev1.Service, error) {
	i := extractInput(ctx)
	addedKubeService, err := pulumikubernetescorev1.NewService(ctx,
		postgresloadbalancercommon.ExternalLoadBalancerServiceName,
		getLoadBalancerServiceArgs(i, postgresloadbalancercommon.ExternalLoadBalancerServiceName, i.ExternalEndpoint),
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "30s", Update: "30s", Delete: "30s"}), pulumi.Parent(i.Namespace))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kubernetes service of type load balancer")
	}
	return addedKubeService, nil
}

func addInternal(ctx *pulumi.Context) (*pulumikubernetescorev1.Service, error) {
	i := extractInput(ctx)
	addedKubeService, err := pulumikubernetescorev1.NewService(ctx,
		postgresloadbalancercommon.InternalLoadBalancerServiceName,
		getInternalLoadBalancerServiceArgs(i, i.InternalEndpoint, i.Namespace),
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "30s", Update: "30s", Delete: "30s"}), pulumi.Parent(i.Namespace))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kubernetes service of type load balancer")
	}
	return addedKubeService, nil
}

func getInternalLoadBalancerServiceArgs(i *input, hostname string, namespace *pulumikubernetescorev1.Namespace) *pulumikubernetescorev1.ServiceArgs {
	resp := getLoadBalancerServiceArgs(i, postgresloadbalancercommon.InternalLoadBalancerServiceName, hostname)
	resp.Metadata = &metav1.ObjectMetaArgs{
		Name:      pulumi.String(postgresloadbalancercommon.InternalLoadBalancerServiceName),
		Namespace: namespace.Metadata.Name(),
		Labels:    namespace.Metadata.Labels(),
		Annotations: pulumi.StringMap{
			"cloud.google.com/load-balancer-type":       pulumi.String("Internal"),
			"planton.cloud/endpoint-domain-name":        pulumi.String(i.EndpointDomainName),
			"external-dns.alpha.kubernetes.io/hostname": pulumi.String(hostname),
		},
	}
	return resp
}

func getLoadBalancerServiceArgs(input *input, serviceName, hostname string) *pulumikubernetescorev1.ServiceArgs {
	return &pulumikubernetescorev1.ServiceArgs{
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String(serviceName),
			Namespace: input.Namespace.Metadata.Name(),
			Annotations: pulumi.StringMap{
				"planton.cloud/endpoint-domain-name":        pulumi.String(input.EndpointDomainName),
				"external-dns.alpha.kubernetes.io/hostname": pulumi.String(hostname),
			},
		},
		Spec: &corev1.ServiceSpecArgs{
			Type: pulumi.String("LoadBalancer"), // Service type is LoadBalancer
			Selector: pulumi.StringMap{
				//postgres-operator generated labels for the postgres pod
				englishword.EnglishWord_application.String(): pulumi.String(englishword.EnglishWord_spilo.String()),
				englishword.EnglishWord_team.String():        pulumi.String(postgrescluster.TeamId),
				"cluster-name":                               pulumi.String(postgrescluster.GetDatabaseName()),
			},
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Name:       pulumi.String(englishword.EnglishWord_postgres.String()),
					Protocol:   pulumi.String("TCP"),
					Port:       pulumi.Int(postgrescluster.PostgresContainerPort),
					TargetPort: pulumi.Int(postgrescluster.PostgresContainerPort),
				},
			},
		},
	}
}

func addLoadBalancerExternalServiceToContext(existingConfig *postgrescontextconfig.ContextConfig, loadBalancerService *pulumikubernetescorev1.Service) {
	if existingConfig.Status.AddedResources == nil {
		existingConfig.Status.AddedResources = &postgrescontextconfig.AddedResources{
			LoadBalancerExternalService: loadBalancerService,
		}
		return
	}
	existingConfig.Status.AddedResources.LoadBalancerExternalService = loadBalancerService
}

func addLoadBalancerInternalServiceToContext(existingConfig *postgrescontextconfig.ContextConfig, loadBalancerService *pulumikubernetescorev1.Service) {
	if existingConfig.Status.AddedResources == nil {
		existingConfig.Status.AddedResources = &postgrescontextconfig.AddedResources{
			LoadBalancerInternalService: loadBalancerService,
		}
		return
	}
	existingConfig.Status.AddedResources.LoadBalancerInternalService = loadBalancerService
}

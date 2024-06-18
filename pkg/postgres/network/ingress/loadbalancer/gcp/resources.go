package gcp

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/english/enums/englishword"
	postgrescluster "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/cluster"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/hostname"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress/loadbalancer/common"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Input struct {
	AddedNamespace     *pulumikubernetescorev1.Namespace
	PostgresClusterId  string
	EnvironmentName    string
	EndpointDomainName string
}

func Resources(ctx *pulumi.Context, input *Input) error {
	// Create a Kubernetes Service of type LoadBalancer
	if err := addExternal(ctx, input); err != nil {
		return errors.Wrap(err, "failed to add external load balancer")
	}
	if err := addInternal(ctx, input); err != nil {
		return errors.Wrap(err, "failed to add internal load balancer")
	}
	return nil
}

func addExternal(ctx *pulumi.Context, input *Input) error {
	hostname := hostname.GetExternalClusterHostname(input.PostgresClusterId, input.EnvironmentName, input.EndpointDomainName)
	addedKubeService, err := pulumikubernetescorev1.NewService(ctx,
		common.ExternalLoadBalancerServiceName,
		getLoadBalancerServiceArgs(input, common.ExternalLoadBalancerServiceName, hostname), pulumi.Parent(input.AddedNamespace))
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes service of type load balancer")
	}

	exportIpAddress(ctx, addedKubeService, "postgres-ingress-external-lb-ip")
	return nil
}

func addInternal(ctx *pulumi.Context, input *Input) error {
	hostname := hostname.GetInternalClusterHostname(input.PostgresClusterId, input.EnvironmentName, input.EndpointDomainName)
	addedKubeService, err := pulumikubernetescorev1.NewService(ctx,
		common.InternalLoadBalancerServiceName,
		getInternalLoadBalancerServiceArgs(input, hostname), pulumi.Parent(input.AddedNamespace))
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes service of type load balancer")
	}

	exportIpAddress(ctx, addedKubeService, "postgres-ingress-internal-lb-ip")
	return nil
}

func exportIpAddress(ctx *pulumi.Context, addedKubeService *pulumikubernetescorev1.Service, outputName string) {
	// Wait for the LoadBalancer IP to be available and export it
	externalLoadBalancerIp := addedKubeService.Status.ApplyT(func(status *pulumikubernetescorev1.ServiceStatus) (string, error) {
		if status.LoadBalancer.Ingress == nil || len(status.LoadBalancer.Ingress) == 0 {
			return "", errors.New("ingress LoadBalancer not found after service initialization is complete")
		}
		ingressIP := status.LoadBalancer.Ingress[0].Ip
		if ingressIP == nil {
			return "", errors.New("ingress LoadBalancer does not have an ip after service initialization is complete")
		}
		return *ingressIP, nil
	})

	ctx.Export(outputName, externalLoadBalancerIp)
}

func getInternalLoadBalancerServiceArgs(input *Input, hostname string) *pulumikubernetescorev1.ServiceArgs {
	resp := getLoadBalancerServiceArgs(input, common.InternalLoadBalancerServiceName, hostname)
	resp.Metadata = &metav1.ObjectMetaArgs{
		Name:      pulumi.String(common.InternalLoadBalancerServiceName),
		Namespace: input.AddedNamespace.Metadata.Name(),
		Annotations: pulumi.StringMap{
			"cloud.google.com/load-balancer-type":       pulumi.String("Internal"),
			"planton.cloud/endpoint-domain-name":        pulumi.String(input.EndpointDomainName),
			"external-dns.alpha.kubernetes.io/hostname": pulumi.String(hostname),
		},
	}
	return resp
}

func getLoadBalancerServiceArgs(input *Input, serviceName, hostname string) *pulumikubernetescorev1.ServiceArgs {
	return &pulumikubernetescorev1.ServiceArgs{
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String(serviceName),
			Namespace: input.AddedNamespace.Metadata.Name(),
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

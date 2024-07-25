package pkg

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/locals"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmetav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func loadBalancerIngress(ctx *pulumi.Context,
	createdNamespace *kubernetescorev1.Namespace) error {
	_, err := kubernetescorev1.NewService(ctx,
		"ingress-external-lb",
		&kubernetescorev1.ServiceArgs{
			Metadata: &kubernetesmetav1.ObjectMetaArgs{
				Name:      pulumi.String("ingress-external-lb"),
				Namespace: createdNamespace.Metadata.Name(),
				Labels:    createdNamespace.Metadata.Labels(),
				Annotations: pulumi.StringMap{
					"planton.cloud/endpoint-domain-name":        pulumi.String(locals.PostgresKubernetes.Spec.Ingress.EndpointDomainName),
					"external-dns.alpha.kubernetes.io/hostname": pulumi.String(locals.IngressExternalHostname),
				},
			},
			Spec: &kubernetescorev1.ServiceSpecArgs{
				Type: pulumi.String("LoadBalancer"), // Service type is LoadBalancer
				Ports: kubernetescorev1.ServicePortArray{
					&kubernetescorev1.ServicePortArgs{
						Name:     pulumi.String("tcp-redis"),
						Port:     pulumi.Int(6379),
						Protocol: pulumi.String("TCP"),
						// This assumes your Redis pod has a port named 'redis'
						TargetPort: pulumi.String("redis"),
					},
				},
				Selector: pulumi.ToStringMap(locals.PostgresPodSectorLabels),
			},
		}, pulumi.Parent(createdNamespace))
	if err != nil {
		return errors.Wrapf(err, "failed to create external load balancer service")
	}

	_, err = kubernetescorev1.NewService(ctx,
		"ingress-internal-lb",
		&kubernetescorev1.ServiceArgs{
			Metadata: &kubernetesmetav1.ObjectMetaArgs{
				Name:      pulumi.String("ingress-internal-lb"),
				Namespace: createdNamespace.Metadata.Name(),
				Labels:    createdNamespace.Metadata.Labels(),
				Annotations: pulumi.StringMap{
					"cloud.google.com/load-balancer-type":       pulumi.String("Internal"),
					"planton.cloud/endpoint-domain-name":        pulumi.String(locals.PostgresKubernetes.Spec.Ingress.EndpointDomainName),
					"external-dns.alpha.kubernetes.io/hostname": pulumi.String(locals.IngressInternalHostname),
				},
			},
			Spec: &kubernetescorev1.ServiceSpecArgs{
				Type: pulumi.String("LoadBalancer"), // Service type is LoadBalancer
				Ports: kubernetescorev1.ServicePortArray{
					&kubernetescorev1.ServicePortArgs{
						Name:     pulumi.String("tcp-redis"),
						Port:     pulumi.Int(6379),
						Protocol: pulumi.String("TCP"),
						// This assumes your Redis pod has a port named 'redis'
						TargetPort: pulumi.String("redis"),
					},
				},
				Selector: pulumi.ToStringMap(locals.PostgresPodSectorLabels),
			},
		}, pulumi.Parent(createdNamespace))
	if err != nil {
		return errors.Wrapf(err, "failed to create external load balancer service")
	}

	return nil
}
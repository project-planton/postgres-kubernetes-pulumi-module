package pkg

import (
	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmetav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (s *ResourceStack) loadBalancerIngress(ctx *pulumi.Context,
	createdNamespace *kubernetescorev1.Namespace) error {

	redisKubernetes := s.Input.ApiResource

	_, err := kubernetescorev1.NewService(ctx,
		"ingress-external-lb",
		&kubernetescorev1.ServiceArgs{
			Metadata: &kubernetesmetav1.ObjectMetaArgs{
				Name:      pulumi.String("ingress-external-lb"),
				Namespace: createdNamespace.Metadata.Name(),
				Labels:    createdNamespace.Metadata.Labels(),
				Annotations: pulumi.StringMap{
					"planton.cloud/endpoint-domain-name": pulumi.String(redisKubernetes.Spec.Ingress.EndpointDomainName),
					"external-dns.alpha.kubernetes.io/hostname": pulumi.Sprintf("%s.%s",
						redisKubernetes.Metadata.Id,
						redisKubernetes.Spec.Ingress.EndpointDomainName)}},
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
				Selector: pulumi.StringMap{
					"app.kubernetes.io/component": pulumi.String("master"),
					"app.kubernetes.io/instance":  pulumi.String(redisKubernetes.Metadata.Id),
					"app.kubernetes.io/name":      pulumi.String("redis"),
				},
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
					"cloud.google.com/load-balancer-type": pulumi.String("Internal"),
					"planton.cloud/endpoint-domain-name":  pulumi.String(redisKubernetes.Spec.Ingress.EndpointDomainName),
					"external-dns.alpha.kubernetes.io/hostname": pulumi.Sprintf("%s-internal.%s",
						redisKubernetes.Metadata.Id,
						redisKubernetes.Spec.Ingress.EndpointDomainName),
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
				Selector: pulumi.StringMap{
					"app.kubernetes.io/component": pulumi.String("master"),
					"app.kubernetes.io/instance":  pulumi.String(redisKubernetes.Metadata.Id),
					"app.kubernetes.io/name":      pulumi.String("redis"),
				},
			},
		}, pulumi.Parent(createdNamespace))
	if err != nil {
		return errors.Wrapf(err, "failed to create external load balancer service")
	}

	return nil
}

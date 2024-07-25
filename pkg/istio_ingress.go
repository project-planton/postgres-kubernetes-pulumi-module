package pkg

import (
	"github.com/pkg/errors"
	certmanagerv1 "github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/certmanager/certmanager/v1"
	istiov1 "github.com/plantoncloud/kubernetes-crd-pulumi-types/pkg/istio/networking/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	v1 "istio.io/api/networking/v1"
)

const (
	IstioIngressNamespace = "istio-ingress"
)

func (s *ResourceStack) istioIngress(ctx *pulumi.Context, createdNamespace *kubernetescorev1.Namespace) error {
	//create variable with descriptive name for the api-resource in the input
	redisKubernetes := s.Input.ApiResource

	//create certificate
	createdCertificate, err := certmanagerv1.NewCertificate(ctx,
		"ingress-certificate",
		&certmanagerv1.CertificateArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(redisKubernetes.Metadata.Id),
				Namespace: createdNamespace.Metadata.Name(),
				Labels:    pulumi.ToStringMap(s.Labels),
			},
			Spec: certmanagerv1.CertificateSpecArgs{
				DnsNames: pulumi.StringArray{
					pulumi.Sprintf("%s.%s", redisKubernetes.Metadata.Id,
						redisKubernetes.Spec.Ingress.EndpointDomainName),
					pulumi.Sprintf("%s-internal.%s", redisKubernetes.Metadata.Id,
						redisKubernetes.Spec.Ingress.EndpointDomainName),
				},
				SecretName: nil,
				IssuerRef: certmanagerv1.CertificateSpecIssuerRefArgs{
					Kind: pulumi.String("ClusterIssuer"),
					//note: a ClusterIssuer resource should have already exist on the kubernetes-cluster.
					//this is typically taken care of by the kubernetes cluster administrator.
					//if the kubernetes-cluster is created using Planton Cloud, then the cluster-issuer name will be
					//same as the ingress-domain-name as long as the same ingress-domain-name is added to the list of
					//ingress-domain-names for the GkeCluster/EksCluster/AksCluster spec.
					Name: pulumi.String(redisKubernetes.Spec.Ingress.EndpointDomainName),
				},
			},
		})
	if err != nil {
		return errors.Wrap(err, "error creating certificate")
	}

	//create gateway
	_, err = istiov1.NewGateway(ctx,
		redisKubernetes.Metadata.Id,
		&istiov1.GatewayArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String(redisKubernetes.Metadata.Id),
				//all istio gateways should be created in istio-ingress deployment namespace
				Namespace: pulumi.String(IstioIngressNamespace),
				Labels:    pulumi.ToStringMap(s.Labels),
			},
			Spec: istiov1.GatewaySpecArgs{
				//the selector labels map should match the desired istio-ingress deployment.
				Selector: pulumi.StringMap{
					"app":   pulumi.String("istio-ingress"),
					"istio": pulumi.String("ingress"),
				},
				Servers: istiov1.GatewaySpecServersArray{
					&istiov1.GatewaySpecServersArgs{
						Name: pulumi.String("redis"),
						Port: &istiov1.GatewaySpecServersPortArgs{
							//important: istio-ingress load-balancer service
							//should accept connections on this port.
							//this is typically archived by including this port in
							//the istio-ingress helm chart values.
							Number:   pulumi.Int(6379),
							Protocol: pulumi.String("TLS"),
							Name:     pulumi.String("redis"),
						},
						Hosts: pulumi.StringArray{
							pulumi.Sprintf("%s.%s", redisKubernetes.Metadata.Id,
								redisKubernetes.Spec.Ingress.EndpointDomainName),
							pulumi.Sprintf("%s-internal.%s", redisKubernetes.Metadata.Id,
								redisKubernetes.Spec.Ingress.EndpointDomainName),
						},
						Tls: &istiov1.GatewaySpecServersTlsArgs{
							CredentialName: createdCertificate.Spec.SecretName(),
							Mode:           pulumi.String(v1.ServerTLSSettings_SIMPLE.String()),
						},
					},
					&istiov1.GatewaySpecServersArgs{
						Name: pulumi.String("redis-http"),
						Port: &istiov1.GatewaySpecServersPortArgs{
							Number:   pulumi.Int(80),
							Name:     pulumi.String("redis-http"),
							Protocol: pulumi.String("HTTP"),
						},
						Hosts: pulumi.StringArray{
							pulumi.Sprintf("%s.%s", redisKubernetes.Metadata.Id,
								redisKubernetes.Spec.Ingress.EndpointDomainName),
							pulumi.Sprintf("%s-internal.%s", redisKubernetes.Metadata.Id,
								redisKubernetes.Spec.Ingress.EndpointDomainName),
						},
						Tls: &istiov1.GatewaySpecServersTlsArgs{
							HttpsRedirect: pulumi.Bool(true),
						},
					},
				},
			},
		})
	if err != nil {
		return errors.Wrap(err, "error creating gateway")
	}

	//create virtual-service
	_, err = istiov1.NewVirtualService(ctx,
		redisKubernetes.Metadata.Id,
		&istiov1.VirtualServiceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(redisKubernetes.Metadata.Id),
				Namespace: createdNamespace.Metadata.Name(),
				Labels:    pulumi.ToStringMap(s.Labels),
			},
			Spec: istiov1.VirtualServiceSpecArgs{
				Gateways: pulumi.StringArray{
					pulumi.Sprintf("%s/%s", IstioIngressNamespace,
						redisKubernetes.Metadata.Id),
				},
				Hosts: pulumi.StringArray{
					pulumi.Sprintf("%s.%s", redisKubernetes.Metadata.Id,
						redisKubernetes.Spec.Ingress.EndpointDomainName),
					pulumi.Sprintf("%s-internal.%s", redisKubernetes.Metadata.Id,
						redisKubernetes.Spec.Ingress.EndpointDomainName),
				},
				Tcp: istiov1.VirtualServiceSpecTcpArray{
					&istiov1.VirtualServiceSpecTcpArgs{
						Match: istiov1.VirtualServiceSpecTcpMatchArray{
							istiov1.VirtualServiceSpecTcpMatchArgs{
								//important: istio-ingress load-balancer service
								//should accept connections on this port.
								//this is typically archived by including this port in
								//the istio-ingress helm chart values.
								Port: pulumi.Int(6739),
							},
						},
						Route: istiov1.VirtualServiceSpecTcpRouteArray{
							&istiov1.VirtualServiceSpecTcpRouteArgs{
								Destination: istiov1.VirtualServiceSpecTcpRouteDestinationArgs{
									//kubernetes service name will have '-master` as suffix
									Host: pulumi.Sprintf("%s-master.%s.svc.cluster.local.",
										redisKubernetes.Metadata.Name,
										createdNamespace.Metadata.Name()),
									Port: istiov1.VirtualServiceSpecTcpRouteDestinationPortArgs{
										Number: pulumi.Int(6739),
									},
								},
							},
						},
					},
				},
			},
			Status: nil,
		})
	if err != nil {
		return errors.Wrap(err, "error creating virtual-service")
	}
	return nil
}

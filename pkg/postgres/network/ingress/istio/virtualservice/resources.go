package virtualservice

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/kubernetes/manifest"
	commonskubernetesdns "github.com/plantoncloud-inc/go-commons/kubernetes/network/dns"
	ingressnamespace "github.com/plantoncloud/kube-cluster-pulumi-blueprint/pkg/gcp/container/addon/istio/ingress/namespace"
	"github.com/plantoncloud/kube-cluster-pulumi-blueprint/pkg/gcp/container/ingress/gateway/postgres"
	code2cloudv1deploypgcstackk8smodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/postgreskubernetes/stack/kubernetes/model"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/cluster"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/hostname"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress/istio/stunnel"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	pulumik8syaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"path/filepath"

	istionetworkingv1beta1 "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Input struct {
	StackResourceInput *code2cloudv1deploypgcstackk8smodel.PostgresClusterKubernetesStackResourceInput
	Labels             map[string]string
	NamespaceName      string
	WorkspaceDir       string
	Namespace          *v1.Namespace
}

func Resources(ctx *pulumi.Context, input *Input) error {
	kubeSvcName := cluster.GetDatabaseName()

	externalEndpointHostname := getExternalClusterHostname(input.StackResourceInput)
	externalName := fmt.Sprintf("%s-external", kubeSvcName)
	externalVirtualServiceObject := buildVirtualServiceObject(externalName, input.NamespaceName, externalEndpointHostname,
		cluster.PostgresContainerPort)
	if err := addVirtualService(ctx, externalVirtualServiceObject, input.Namespace, input.WorkspaceDir); err != nil {
		return errors.Wrapf(err, "failed to add virtual-service for %s domain", externalEndpointHostname)
	}
	internalEndpointHostname := getInternalClusterHostname(input.StackResourceInput)
	internalName := fmt.Sprintf("%s-internal", kubeSvcName)
	internalVirtualServiceObject := buildVirtualServiceObject(internalName, input.NamespaceName, internalEndpointHostname,
		cluster.PostgresContainerPort)
	if err := addVirtualService(ctx, internalVirtualServiceObject, input.Namespace, input.WorkspaceDir); err != nil {
		return errors.Wrapf(err, "failed to add virtual-service for %s domain", internalEndpointHostname)
	}

	return nil
}

func getExternalClusterHostname(stackResourceInput *code2cloudv1deploypgcstackk8smodel.PostgresClusterKubernetesStackResourceInput) string {
	return hostname.GetExternalClusterHostname(stackResourceInput.PostgresCluster.Metadata.Id,
		stackResourceInput.PostgresCluster.Spec.EnvironmentInfo.EnvironmentName,
		stackResourceInput.PostgresCluster.Spec.Kubernetes.Ingress.EndpointDomainName)
}

func getInternalClusterHostname(stackResourceInput *code2cloudv1deploypgcstackk8smodel.PostgresClusterKubernetesStackResourceInput) string {
	return hostname.GetInternalClusterHostname(stackResourceInput.PostgresCluster.Metadata.Id,
		stackResourceInput.PostgresCluster.Spec.EnvironmentInfo.EnvironmentName,
		stackResourceInput.PostgresCluster.Spec.Kubernetes.Ingress.EndpointDomainName)
}

//buildVirtualServiceObject builds virtual service object
/*
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: product-name
  namespace: tenant-product-product_env-name
spec:
  hosts:
  - <postgres-cluster-id>.dev.example.com
  gateways:
  - istio-ingress/postgres
  tls:
  - match:
    - port: 5432
      sniHosts:
		- <postgres-cluster-id>.dev.example.com
    route:
    - destination:
        host: product-name.tenant-product-product_env-name.svc.cluster.local
        port:
          number: 5432
*/
func buildVirtualServiceObject(name, namespaceName, hostname string, port int32) *v1beta1.VirtualService {
	return &v1beta1.VirtualService{
		TypeMeta: k8smetav1.TypeMeta{
			APIVersion: "networking.istio.io/v1beta1",
			Kind:       "VirtualService",
		},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      name,
			Namespace: namespaceName,
		},
		Spec: istionetworkingv1beta1.VirtualService{
			Gateways: []string{fmt.Sprintf("%s/%s", ingressnamespace.Name, postgres.GatewayName)},
			Hosts:    []string{hostname},
			Tls: []*networkingv1beta1.TLSRoute{{
				Match: []*networkingv1beta1.TLSMatchAttributes{
					{
						Port:     uint32(cluster.PostgresContainerPort),
						SniHosts: []string{hostname},
					},
				},
				Route: []*istionetworkingv1beta1.RouteDestination{{
					Destination: &istionetworkingv1beta1.Destination{
						Host: fmt.Sprintf("%s.%s.%s", stunnel.StunnelServiceName, namespaceName, commonskubernetesdns.DefaultDomain),
						Port: &istionetworkingv1beta1.PortSelector{Number: uint32(port)},
					},
				}},
			}},
		},
	}
}

func addVirtualService(ctx *pulumi.Context, virtualServiceObject *v1beta1.VirtualService, namespace *v1.Namespace, workspace string) error {
	resourceName := fmt.Sprintf("virtual-service-%s", virtualServiceObject.Name)
	manifestPath := filepath.Join(workspace, fmt.Sprintf("%s.yaml", resourceName))
	if err := manifest.Create(manifestPath, virtualServiceObject); err != nil {
		return errors.Wrapf(err, "failed to create %s manifest file", manifestPath)
	}
	_, err := pulumik8syaml.NewConfigFile(ctx, resourceName, &pulumik8syaml.ConfigFileArgs{
		File: manifestPath,
	}, pulumi.DependsOn([]pulumi.Resource{namespace}), pulumi.Parent(namespace))
	if err != nil {
		return errors.Wrap(err, "failed to add virtual-service manifest")
	}
	return nil
}

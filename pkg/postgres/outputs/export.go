package outputs

import (
	"fmt"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/english/enums/englishword"
	"github.com/plantoncloud/pulumi-stack-runner-go-sdk/pkg/name/output/custom"
	puluminamekubeoutput "github.com/plantoncloud/pulumi-stack-runner-go-sdk/pkg/name/provider/kubernetes/output"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Export(ctx *pulumi.Context) error {
	var i = extractInput(ctx)
	var kubePortForwardCommand = getKubePortForwardCommand(i.NamespaceName, i.ResourceName)

	ctx.Export(GetExternalClusterHostnameOutputName(), pulumi.String(i.ExternalHostname))
	ctx.Export(GetInternalClusterHostnameOutputName(), pulumi.String(i.InternalHostname))

	ctx.Export(GetKubeServiceNameOutputName(), pulumi.String(i.KubeServiceName))

	ctx.Export(GetKubeEndpointOutputName(), pulumi.String(i.KubeLocalEndpoint))

	ctx.Export(GetKubePortForwardCommandOutputName(), pulumi.String(kubePortForwardCommand))
	ctx.Export(GetExternalLoadBalancerIp(), pulumi.String(i.ExternalLoadBalancerIpAddress))
	ctx.Export(GetInternalLoadBalancerIp(), pulumi.String(i.InternalLoadBalancerIpAddress))
	ctx.Export(GetNamespaceNameOutputName(), pulumi.String(i.NamespaceName))

	return nil
}

func GetExternalClusterHostnameOutputName() string {
	return custom.Name("postgres-cluster-external-hostname")
}

func GetInternalClusterHostnameOutputName() string {
	return custom.Name("postgres-cluster-internal-hostname")
}

func GetKubeServiceNameOutputName() string {
	return custom.Name("postgres-cluster-kubernetes-service-name")
}

func GetKubeEndpointOutputName() string {
	return custom.Name("postgres-cluster-kubernetes-endpoint")
}

func GetKubePortForwardCommandOutputName() string {
	return custom.Name("postgres-cluster-kube-port-forward-command")
}

func GetExternalLoadBalancerIp() string {
	return custom.Name("postgres-ingress-external-lb-ip")
}

func GetInternalLoadBalancerIp() string {
	return custom.Name("postgres-ingress-internal-lb-ip")
}

func GetNamespaceNameOutputName() string {
	return puluminamekubeoutput.Name(kubernetescorev1.Namespace{}, englishword.EnglishWord_namespace.String())
}

// getKubePortForwardCommand returns kubectl port-forward command that can be used by developers.
// ex: "kubectl port-forward -n kubernetes_namespace  service/main-mongodb-cluster 8080:8080"
func getKubePortForwardCommand(namespaceName, kubeServiceName string) string {
	return fmt.Sprintf("kubectl port-forward -n %s service/%s %d:%d",
		namespaceName, kubeServiceName, MongoDbPort, MongoDbPort)
}

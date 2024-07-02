package hostname

import (
	"fmt"
	"github.com/plantoncloud/pulumi-blueprint-golang-commons/pkg/pulumi/pulumicustomoutput"

	code2cloudv1deploypgcmodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/postgreskubernetes/model"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Input struct {
	PostgresCluster *code2cloudv1deploypgcmodel.PostgresKubernetes
}

func Resources(ctx *pulumi.Context, input *Input) error {
	ctx.Export(GetKubeEndpointOutputName(), pulumi.String(GetKubeEndpoint(input.PostgresCluster.Metadata.Id, input.PostgresCluster.Spec.EnvironmentInfo.EnvironmentName, input.PostgresCluster.Spec.Ingress.EndpointDomainName)))

	if input.PostgresCluster.Spec.Ingress == nil ||
		!input.PostgresCluster.Spec.Ingress.IsEnabled {
		ctx.Export(GetExternalClusterHostnameOutputName(), pulumi.String("n/a"))
		ctx.Export(GetInternalClusterHostnameOutputName(), pulumi.String("n/a"))
	} else {
		ctx.Export(GetExternalClusterHostnameOutputName(), pulumi.String(GetExternalClusterHostname(input.PostgresCluster.Metadata.Id, input.PostgresCluster.Spec.EnvironmentInfo.EnvironmentName, input.PostgresCluster.Spec.Ingress.EndpointDomainName)))
		ctx.Export(GetInternalClusterHostnameOutputName(), pulumi.String(GetInternalClusterHostname(input.PostgresCluster.Metadata.Id, input.PostgresCluster.Spec.EnvironmentInfo.EnvironmentName, input.PostgresCluster.Spec.Ingress.EndpointDomainName)))
	}

	return nil
}

func GetExternalClusterHostname(postgresClusterId, envName, endpointDomainName string) string {
	return fmt.Sprintf("%s.%s.%s", postgresClusterId, envName, endpointDomainName)
}

func GetInternalClusterHostname(postgresClusterId, envName, endpointDomainName string) string {
	return fmt.Sprintf("%s.%s-internal.%s", postgresClusterId, envName, endpointDomainName)
}

func GetKubeEndpoint(productId, postgresClusterName, kubernetesNamespaceName string) string {
	return fmt.Sprintf("%s-%s.%s", productId, postgresClusterName, kubernetesNamespaceName)
}

func GetExternalClusterHostnameOutputName() string {
	return pulumicustomoutput.Name("external-cluster-hostname")
}

func GetInternalClusterHostnameOutputName() string {
	return pulumicustomoutput.Name("internal-cluster-hostname")
}

func GetKubeEndpointOutputName() string {
	return pulumicustomoutput.Name("kube-endpoint")
}

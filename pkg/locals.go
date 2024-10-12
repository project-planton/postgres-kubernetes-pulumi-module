package pkg

import (
	postgreskubernetesv1 "buf.build/gen/go/project-planton/apis/protocolbuffers/go/project/planton/provider/kubernetes/postgreskubernetes/v1"
	"fmt"
	"github.com/project-planton/postgres-kubernetes-pulumi-module/pkg/outputs"
	"github.com/project-planton/pulumi-module-golang-commons/pkg/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"strconv"
)

type Locals struct {
	IngressExternalHostname string
	IngressInternalHostname string
	KubePortForwardCommand  string
	KubeServiceFqdn         string
	KubeServiceName         string
	Namespace               string
	PostgresKubernetes      *postgreskubernetesv1.PostgresKubernetes
	PostgresPodSectorLabels map[string]string
	Labels                  map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *postgreskubernetesv1.PostgresKubernetesStackInput) *Locals {
	locals := &Locals{}

	//if the id is empty, use name as id
	if stackInput.Target.Metadata.Id == "" {
		stackInput.Target.Metadata.Id = stackInput.Target.Metadata.Name
	}

	postgresKubernetes := stackInput.Target

	//assign value for the local variable to make it available across the module.
	locals.PostgresKubernetes = postgresKubernetes

	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceId:   postgresKubernetes.Metadata.Id,
		kuberneteslabelkeys.ResourceKind: "postgres_kubernetes",
	}

	if postgresKubernetes.Spec.EnvironmentInfo != nil {
		locals.Labels[kuberneteslabelkeys.Environment] = postgresKubernetes.Spec.EnvironmentInfo.EnvId
		locals.Labels[kuberneteslabelkeys.Organization] = postgresKubernetes.Spec.EnvironmentInfo.OrgId
	}

	//decide on the namespace
	locals.Namespace = postgresKubernetes.Metadata.Id

	ctx.Export(outputs.Namespace, pulumi.String(locals.Namespace))

	locals.PostgresPodSectorLabels = map[string]string{
		"planton.cloud/resource-kind": "postgres_kubernetes",
		"planton.cloud/resource-id":   postgresKubernetes.Metadata.Id,
	}

	ctx.Export(outputs.PostgresUserCredentialsSecretName,
		pulumi.Sprintf("postgres.db-%s.credentials.postgresql.acid.zalan.do",
			postgresKubernetes.Metadata.Id))
	ctx.Export(outputs.PostgresUsernameSecretKey, pulumi.String("username"))
	ctx.Export(outputs.PostgresPasswordSecretKey, pulumi.String("password"))

	locals.KubeServiceName = fmt.Sprintf("%s-master", postgresKubernetes.Metadata.Name)

	//export kubernetes service name
	ctx.Export(outputs.Service, pulumi.String(locals.KubeServiceName))

	locals.KubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local", locals.KubeServiceName, locals.Namespace)

	//export kubernetes endpoint
	ctx.Export(outputs.KubeEndpoint, pulumi.String(locals.KubeServiceFqdn))

	locals.KubePortForwardCommand = fmt.Sprintf("kubectl port-forward -n %s service/%s 8080:8080",
		locals.Namespace, locals.KubeServiceName)

	//export kube-port-forward command
	ctx.Export(outputs.KubePortForwardCommand, pulumi.String(locals.KubePortForwardCommand))

	if postgresKubernetes.Spec.Ingress == nil ||
		!postgresKubernetes.Spec.Ingress.IsEnabled ||
		postgresKubernetes.Spec.Ingress.EndpointDomainName == "" {
		return locals
	}

	locals.IngressExternalHostname = fmt.Sprintf("%s.%s", postgresKubernetes.Metadata.Id,
		postgresKubernetes.Spec.Ingress.EndpointDomainName)

	locals.IngressInternalHostname = fmt.Sprintf("%s-internal.%s", postgresKubernetes.Metadata.Id,
		postgresKubernetes.Spec.Ingress.EndpointDomainName)

	//export ingress hostnames
	ctx.Export(outputs.IngressExternalHostname, pulumi.String(locals.IngressExternalHostname))
	ctx.Export(outputs.IngressInternalHostname, pulumi.String(locals.IngressInternalHostname))

	return locals
}

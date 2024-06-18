package cluster

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubecluster/enums/kubernetesworkloadingresstype"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/english/enums/englishword"

	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/kubernetes/manifest"
	kubernetesv1model "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubecluster/model"
	code2cloudv1deploypgcstackk8smodel "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/postgreskubernetes/stack/kubernetes/model"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress/istio/cert"
	pulk8scv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	pulumik8syaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	v1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8sapimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// TeamId required by zalando operator
	TeamId                    = "db"
	EnvVarStunnelSidecarImage = "STUNNEL_SIDECAR_IMAGE"
	StunnelContainerName      = englishword.EnglishWord_stunnel
	StunnelContainerPort      = 15432
	StunnelCertMountPath      = "/server/ca.pem"
)

func init() {

}

type Input struct {
	WorkspaceDir                 string
	NamespaceName                string
	Namespace                    *pulk8scv1.Namespace
	PostgresKubernetesStackInput *code2cloudv1deploypgcstackk8smodel.PostgresKubernetesStackInput
	Labels                       map[string]string
}

func Resources(ctx *pulumi.Context, input *Input) error {
	stunnelSidecarImage, isEnvVarSet := os.LookupEnv(EnvVarStunnelSidecarImage)
	if !isEnvVarSet {
		return errors.Errorf("%s environment variables is not set", EnvVarStunnelSidecarImage)
	}
	if err := addPostgresKubernetes(ctx, input, stunnelSidecarImage); err != nil {
		return errors.Wrap(err, "failed to add postgres cluster")
	}
	return nil
}

func addPostgresKubernetes(ctx *pulumi.Context, input *Input, stunnelSidecarImage string) error {
	postgresKubernetesObject := buildPostgresKubernetesObject(input, stunnelSidecarImage)
	resourceName := fmt.Sprintf("postgres-cluster-%s", postgresKubernetesObject.Name)
	manifestPath := filepath.Join(input.WorkspaceDir, fmt.Sprintf("%s.yaml", resourceName))
	if err := manifest.Create(manifestPath, postgresKubernetesObject); err != nil {
		return errors.Wrapf(err, "failed to create %s manifest file", manifestPath)
	}
	_, err := pulumik8syaml.NewConfigFile(ctx, resourceName, &pulumik8syaml.ConfigFileArgs{
		File: manifestPath,
	}, pulumi.DependsOn([]pulumi.Resource{input.Namespace}), pulumi.Parent(input.Namespace))
	if err != nil {
		return errors.Wrap(err, "failed to add virtual-service manifest")
	}
	return nil
}

func buildPostgresKubernetesObject(input *Input, stunnelSidecarImage string) *v1.Postgresql {
	postgresql := &v1.Postgresql{
		TypeMeta: k8sapimachineryv1.TypeMeta{
			APIVersion: "acid.zalan.do/v1",
			Kind:       "postgresql",
		},
		ObjectMeta: k8sapimachineryv1.ObjectMeta{
			Name:      GetDatabaseName(),
			Namespace: input.NamespaceName,
			Labels:    input.Labels,
		},
		Spec: v1.PostgresSpec{
			PodAnnotations: map[string]string{"postgres-cluster-id": input.PostgresKubernetesStackInput.ResourceInput.Metadata.Name},
			PostgresqlParam: v1.PostgresqlParam{
				PgVersion:  PostgresVersion,
				Parameters: getPostgresParameters(defaultPostgresParameters, input.PostgresKubernetesStackInput.ResourceInput.PostgresParameters),
			},
			Volume: v1.Volume{
				Size: input.PostgresKubernetesKubernetesStackInput.ResourceInput.PostgresKubernetes.Spec.Kubernetes.PostgresContainer.DiskSize,
			},
			Patroni:           v1.Patroni{},
			Resources:         getResources(input.PostgresKubernetesKubernetesStackInput.ResourceInput.PostgresKubernetes.Spec.Kubernetes.PostgresContainer.Resources),
			TeamID:            TeamId,
			Users:             getUsers(input.PostgresKubernetesKubernetesStackInput.ResourceInput.PostgresKubernetesConfig),
			NumberOfInstances: input.PostgresKubernetesKubernetesStackInput.ResourceInput.PostgresKubernetes.Spec.Kubernetes.PostgresContainer.Replicas,
			Databases:         getDatabases(input.PostgresKubernetesKubernetesStackInput.ResourceInput.PostgresKubernetesConfig),
			AdditionalVolumes: []v1.AdditionalVolume{
				{
					Name:      "stunnel-ca",
					MountPath: StunnelCertMountPath,
					SubPath:   "tls-combined.pem",
					TargetContainers: []string{
						StunnelContainerName.String(),
					},
					VolumeSource: k8scorev1.VolumeSource{
						Secret: &k8scorev1.SecretVolumeSource{
							SecretName: cert.GetCertSecretName(cert.Name),
						},
					},
				},
			},
		},
	}

	addSidecars(input, postgresql, stunnelSidecarImage)

	return postgresql
}

func addSidecars(input *Input, postgresql *v1.Postgresql, stunnelSidecarImage string) {
	ingressSpec := input.PostgresKubernetesKubernetesStackInput.ResourceInput.Spec.Ingress
	if ingressSpec != nil || !ingressSpec.IsEnabled ||
		ingressSpec.IngressType != kubernetesworkloadingresstype.KubernetesWorkloadIngressType_ingress_controller {
		return
	}
	postgresql.Spec.Sidecars = []v1.Sidecar{
		{
			Name:        StunnelContainerName.String(),
			Resources:   getStunnelSidecarResources(),
			DockerImage: stunnelSidecarImage,
			Ports: []k8scorev1.ContainerPort{
				{
					Name:          englishword.EnglishWord_postgres.String(),
					ContainerPort: StunnelContainerPort,
					Protocol:      "TCP",
				},
			},
			Env: []k8scorev1.EnvVar{
				{
					Name:  "STUNNEL_MODE",
					Value: englishword.EnglishWord_server.String(),
				}, {
					Name:  "STUNNEL_LOG_LEVEL",
					Value: englishword.EnglishWord_debug.String(),
				}, {
					Name:  "STUNNEL_ACCEPT_PORT",
					Value: strconv.Itoa(StunnelContainerPort),
				}, {
					Name:  "STUNNEL_FORWARD_HOST",
					Value: englishword.EnglishWord_localhost.String(),
				}, {
					Name:  "STUNNEL_FORWARD_PORT",
					Value: strconv.Itoa(PostgresContainerPort),
				},
			},
		},
	}
}

func getStunnelSidecarResources() *v1.Resources {
	return &v1.Resources{
		ResourceRequests: v1.ResourceDescription{
			CPU:    "100m",
			Memory: "100Mi",
		},
		ResourceLimits: v1.ResourceDescription{
			CPU:    "500m",
			Memory: "1Gi",
		},
	}
}

func getResources(inputResources *kubernetesv1model.ContainerResources) *v1.Resources {
	return &v1.Resources{
		ResourceRequests: v1.ResourceDescription{
			CPU:    inputResources.Requests.Cpu,
			Memory: inputResources.Requests.Memory,
		},
		ResourceLimits: v1.ResourceDescription{
			CPU:    inputResources.Limits.Cpu,
			Memory: inputResources.Limits.Memory,
		},
	}
}

func getUsers(clusterConfig *code2cloudv1deploypgcstackk8smodel.PostgresKubernetesConfig) map[string]v1.UserFlags {
	users := make(map[string]v1.UserFlags, 0)
	for _, u := range clusterConfig.Users {
		users[u.Name] = defaultUserFlags
	}
	return users
}

func getDatabases(clusterConfig *code2cloudv1deploypgcstackk8smodel.PostgresKubernetesConfig) map[string]string {
	databases := make(map[string]string, 0)
	for _, d := range clusterConfig.Databases {
		databases[d.Name] = d.Owner.Name
	}
	return databases
}

func getPostgresParameters(defaultParams, inputParams map[string]string) map[string]string {
	parameters := make(map[string]string, 0)
	for k, v := range defaultParams {
		parameters[k] = v
	}
	for k, v := range inputParams {
		parameters[k] = v
	}
	return parameters
}

// GetDatabaseName returns input for zalando operator for name of the database.
// for zolando operator the name is required to be always prefixed by teamId
// a kubernetes service with the same name is created by the operator
func GetDatabaseName() string {
	return fmt.Sprintf("%s-server", TeamId)
}

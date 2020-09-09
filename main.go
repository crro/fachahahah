package main

import (
	"context"
	"fmt"

	//"github.com/rancher/wrangler/pkg/kubeconfig"
	// "github.com/docker/docker/pkg/reexec"
	// projectv1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/admin.rio.cattle.io/v1"
	// "github.com/rancher/wrangler/pkg/kubeconfig"
	// "k8s.io/apimachinery/pkg/api/errors"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/client-go/kubernetes"
	// "k8s.io/client-go/rest"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	rest "k8s.io/client-go/rest"
)

type GenericCondition struct {
	// Type of cluster condition.
	Type string `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime string `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition
	Message string `json:"message,omitempty"`
}

type AutoscaleConfig struct {
	// ContainerConcurrency specifies the maximum allowed in-flight (concurrent) requests per container of the Revision. Defaults to 0 which means unlimited concurrency.
	Concurrency int `json:"concurrency,omitempty"`

	// The minimal number of replicas Service can be scaled
	MinReplicas *int32 `json:"minReplicas,omitempty" mapper:"alias=minScale|min"`

	// The maximum number of replicas Service can be scaled
	MaxReplicas *int32 `json:"maxReplicas,omitempty" mapper:"alias=maxScale|max"`
}

// RolloutConfig specifies the configuration when promoting a new revision
type RolloutConfig struct {
	// Increment Value each Rollout can scale up or down, always a positive number
	Increment int `json:"increment,omitempty"`

	// Interval between each Rollout in seconds
	IntervalSeconds int `json:"intervalSeconds,omitempty" mapper:"alias=interval"`

	// Pause if true the rollout will stop in place until set to false.
	Pause bool `json:"pause,omitempty"`
}

// ServiceSpec represents spec for Service
type ServiceSpec struct {
	PodConfig

	// This service is a template for new versions to be created based on changes
	// from the build.repo
	Template bool `json:"template,omitempty"`

	// Whether to only stage services that are generated through the template from build.repo
	StageOnly bool `json:"stageOnly,omitempty"`

	// Version of this service
	Version string `json:"version,omitempty"`

	// The exposed app name, if no value is set, then metadata.name of the Service is used
	App string `json:"app,omitempty"`

	// The weight among services with matching app field to determine how much traffic is load balanced
	// to this service.  If rollout is set, the weight becomes the target weight of the rollout.
	Weight *int `json:"weight,omitempty"`

	// Number of desired pods. This is a pointer to distinguish between explicit zero and not specified. Defaults to 1 in deployment.
	Replicas *int `json:"replicas,omitempty" mapper:"alias=scale"`

	// The maximum number of pods that can be unavailable during the update.
	// The value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// An absolute number is calculated from percentage by rounding down.
	// This cannot be 0 if MaxSurge is 0.
	// Defaults to 25%.
	// Example: when this is set to 30%, the old ReplicaSet can be scaled down to 70% of desired pods
	// immediately when the rolling update starts. Once new pods are ready, the old ReplicaSet
	// can be scaled down further, followed by scaling up the new ReplicaSet, ensuring
	// that the total number of pods available at all times during the update is at
	// least 70% of desired pods.
	// +optional
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`

	// The maximum number of pods that can be scheduled above the desired number of
	// pods.
	// The value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// This can not be 0 if MaxUnavailable is 0.
	// An absolute number is calculated from percentage by rounding up.
	// Defaults to 25%.
	// Example: when this is set to 30%, the new ReplicaSet can be scaled up immediately when
	// the rolling update starts, such that the total number of old and new pods do not exceed
	// 130% of desired pods. Once the old pods have been killed,
	// the new ReplicaSet can be scaled up further, ensuring that total number of pods running
	// at any time during the update is at most 130% of desired pods.
	// +optional
	MaxSurge *intstr.IntOrString `json:"maxSurge,omitempty"`

	// Autoscale the replicas based on the amount of traffic received by this service
	Autoscale *AutoscaleConfig `json:"autoscale,omitempty"`

	// RolloutDuration specifies time for template service to reach 100% weight, used to set rollout config
	RolloutDuration *metav1.Duration `json:"rolloutDuration,omitempty" mapper:"duration"`

	// RolloutConfig controls how each service is allocated ComputedWeight
	RolloutConfig *RolloutConfig `json:"rollout,omitempty"`

	// Place one pod per node that matches the scheduling rules
	Global bool `json:"global,omitempty"`

	// Whether to disable Service mesh for the Service. If true, no mesh sidecar will be deployed along with the Service
	ServiceMesh *bool `json:"serviceMesh,omitempty"`

	// RequestTimeoutSeconds specifies the timeout set on api gateway for each individual service
	RequestTimeoutSeconds *int `json:"requestTimeoutSeconds,omitempty"`

	// Permissions to the Services. It will create corresponding ServiceAccounts, Roles and RoleBinding.
	Permissions []Permission `json:"permissions,omitempty" mapper:"permissions,alias=permission"`

	// GlobalPermissions to the Services. It will create corresponding ServiceAccounts, ClusterRoles and ClusterRoleBinding.
	GlobalPermissions []Permission `json:"globalPermissions,omitempty" mapper:"permissions,alias=globalPermission"`
}

type PodDNSConfigOption struct {
	Name  string  `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`
}

// ContainerSecurityContext holds pod-level security attributes and common container constants. Optional: Defaults to empty. See type description for default values of each field.
type ContainerSecurityContext struct {
	// The UID to run the entrypoint of the container process. Defaults to user specified in image metadata if unspecified. May also be set in SecurityContext.
	// If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence for that container
	RunAsUser *int64 `json:"runAsUser,omitempty" mapper:"alias=user"`

	// The GID to run the entrypoint of the container process. Uses runtime default if unset. May also be set in SecurityContext.
	// If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence for that container.
	RunAsGroup *int64 `json:"runAsGroup,omitempty" mapper:"alias=user"`

	// Whether this container has a read-only root filesystem. Default is false.
	ReadOnlyRootFilesystem *bool `json:"readOnlyRootFilesystem,omitempty"`

	// Run container in privileged mode.
	// Processes in privileged containers are essentially equivalent to root on the host.
	// Defaults to false.
	// +optional
	Privileged *bool `json:"privileged,omitempty"`
}

type NamedContainer struct {
	// The name of the container
	Name string `json:"name,omitempty"`

	// List of initialization containers belonging to the pod.
	// Init containers are executed in order prior to containers being started.
	// If any init container fails, the pod is considered to have failed and is handled according to its restartPolicy.
	// The name for an init container or normal container must be unique among all containers.
	// Init containers may not have Lifecycle actions, Readiness probes, or Liveness probes.
	// The resourceRequirements of an init container are taken into account during scheduling by finding the highest request/limit for each resource type, and then using the max of of that value or the sum of the normal containers.
	// Limits are applied to init containers in a similar fashion. Init containers cannot currently be added or removed. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	Init bool `json:"init,omitempty"`

	Container
}

type Container struct {
	// Docker image name. More info: https://kubernetes.io/docs/concepts/containers/images This field is optional to allow higher level config management to default or override container images in workload controllers like Deployments and StatefulSets.
	Image string `json:"image,omitempty"`

	// ImageBuild specifies how to build this image
	ImageBuild *ImageBuildSpec `json:"build,omitempty"`

	// Entrypoint array. Not executed within a shell. The docker image's ENTRYPOINT is used if this is not provided.
	// Variable references $(VAR_NAME) are expanded using the container's environment. If a variable cannot be resolved, the reference in the input string will be unchanged.
	// The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not.
	// Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	Command []string `json:"command,omitempty" mapper:"shlex"`

	// Arguments to the entrypoint. The docker image's CMD is used if this is not provided.
	// Variable references $(VAR_NAME) are expanded using the container's environment.
	// If a variable cannot be resolved, the reference in the input string will be unchanged.
	// The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not.
	// Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	Args []string `json:"args,omitempty" mapper:"shlex,alias=arg"`

	// Container's working directory. If not specified, the container runtime's default will be used, which might be configured in the container image. Cannot be updated.
	WorkingDir string `json:"workingDir,omitempty"`

	// List of ports to expose from the container. Exposing a port here gives the system additional information about the network connections a container uses, but is primarily informational. Not specifying a port here DOES NOT prevent that port from being exposed.
	// Any port which is listening on the default "0.0.0.0" address inside a container will be accessible from the network. Cannot be updated.
	Ports []ContainerPort `json:"ports,omitempty" mapper:"ports,alias=port"`

	// List of environment variables to set in the container. Cannot be updated.
	Env []EnvVar `json:"env,omitempty" mapper:"env,envmap=sep==,alias=environment"`

	// CPU, in milliCPU (e.g. 500 = .5 CPU cores)
	CPUMillis *int64 `json:"cpuMillis,omitempty" mapper:"quantity,alias=cpu|cpus"`

	// Memory, in bytes
	MemoryBytes *int64 `json:"memoryBytes,omitempty" mapper:"quantity,alias=mem|memory"`

	// Secrets Mounts
	Secrets []DataMount `json:"secrets,omitempty" mapper:"secrets,envmap=sep=:,alias=secret"`

	// Configmaps Mounts
	Configs []DataMount `json:"configs,omitempty" mapper:"configs,envmap=sep=:,alias=config"`

	// Periodic probe of container liveness. Container will be restarted if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	LivenessProbe *v1.Probe `json:"livenessProbe,omitempty" mapper:"alias=liveness"`

	// Periodic probe of container service readiness. Container will be removed from service endpoints if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	ReadinessProbe *v1.Probe `json:"readinessProbe,omitempty" mapper:"alias=readiness"`

	// Image pull policy. One of Always, Never, IfNotPresent. Defaults to Always if tag is does not start with v[0-9] or [0-9], or IfNotPresent otherwise. Cannot be updated. More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty" mapper:"enum=Always|IfNotPresent|Never,alias=pullPolicy"`

	// Whether this container should allocate a buffer for stdin in the container runtime. If this is not set, reads from stdin in the container will always result in EOF. Default is false.
	Stdin bool `json:"stdin,omitempty" mapper:"alias=stdinOpen"`

	// Whether the container runtime should close the stdin channel after it has been opened by a single attach. When stdin is true the stdin stream will remain open across multiple attach sessions.
	// If stdinOnce is set to true, stdin is opened on container start, is empty until the first client attaches to stdin, and then remains open and accepts data until the client disconnects, at which time stdin is closed and remains closed until the container is restarted. If this flag is false, a container processes that reads from stdin will never receive an EOF. Default is false
	StdinOnce bool `json:"stdinOnce,omitempty"`

	// Whether this container should allocate a TTY for itself, also requires 'stdin' to be true. Default is false.
	TTY bool `json:"tty,omitempty"`

	// Pod volumes to mount into the container's filesystem
	Volumes []Volume `json:"volumes,omitempty" mapper:"volumes,envmap=sep=:,alias=volume"`

	*ContainerSecurityContext
}

type DataMount struct {
	// The directory or file to mount the value to in the container
	Target string `json:"target,omitempty"`
	// The name of the ConfigMap or Secret to mount
	Name string `json:"name,omitempty"`
	// The key in the data of the ConfigMap or Secret to mount to a file.
	// If Key is set the Target must be a file.  If key is set the target must be a directory and will
	// contain one file per key from the Secret/ConfigMap data field.
	Key string `json:"key,omitempty"`
}

type VolumeTemplate struct {
	// Labels to be applied to the created PVC
	Labels map[string]string `json:"labels,omitempty"`
	// Annotations to be applied to the created PVC
	Annotations map[string]string `json:"annotations,omitempty"`

	// Name of the VolumeTemplate. A volume entry will use this name to refer to the created volume
	Name string

	// AccessModes contains the desired access modes the volume should have.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1
	// +optional
	AccessModes []v1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`
	// Resources represents the minimum resources the volume should have.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources
	// +optional
	StorageRequest int64 `json:"storage,omitempty" mapper:"quantity"`
	// Name of the StorageClass required by the claim.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1
	StorageClassName string `json:"storageClassName,omitempty"`
	// volumeMode defines what type of volume is required by the claim.
	// Value of Filesystem is implied when not included in claim spec.
	// This is a beta feature.
	// +optional
	VolumeMode *v1.PersistentVolumeMode `json:"volumeMode,omitempty"`
}

type Volume struct {
	// Name is the name of the volume. If multiple Volumes in the same pod share the same name they
	// will be the same underlying storage. If persistent is set to true Name is required and will be
	// used to reference a PersistentVolumeClaim in the current namespace.
	//
	// If Name matches the name of a VolumeTemplate on this service then the VolumeTemplate will be used as the
	// source of the volume.
	Name string `json:"name,omitempty"`

	// That path within the container to mount the volume to
	Path string `json:"path,omitempty"`

	// That path on the host to mount into this container
	HostPath string `json:"hostpath,omitempty"`

	// HostPathType specify HostPath type
	HostPathType *v1.HostPathType `json:"hostPathType,omitempty" protobuf:"bytes,2,opt,name=type"`

	// If Persistent is true then this volume refers to a PersistentVolumeClaim in this namespace. The
	// Name field is used to reference PersistentVolumeClaim.  If the Name of this Volume matches a VolumeTemplate
	// then Persistent is assumed to be true
	Persistent bool `json:"persistent,omitempty"`
}

type EnvVar struct {
	Name          string `json:"name,omitempty"`
	Value         string `json:"value,omitempty"`
	SecretName    string `json:"secretName,omitempty"`
	ConfigMapName string `json:"configMapName,omitempty"`
	Key           string `json:"key,omitempty"`
}

type DNS struct {
	// Set DNS policy for the pod. Defaults to "ClusterFirst". Valid values are 'ClusterFirstWithHostNet', 'ClusterFirst', 'Default' or 'None'.
	// DNS parameters given in DNSConfig will be merged with the policy selected with DNSPolicy.
	// To have DNS options set along with hostNetwork, you have to specify DNS policy explicitly to 'ClusterFirstWithHostNet'.
	Policy v1.DNSPolicy `json:"policy,omitempty" mapper:"enum=ClusterFirst|ClusterFirstWithHostNet|Default|None|host=Default"`

	// A list of DNS name server IP addresses. This will be appended to the base nameservers generated from DNSPolicy. Duplicated nameservers will be removed.
	Nameservers []string `json:"nameservers,omitempty"`

	// A list of DNS search domains for host-name lookup. This will be appended to the base search paths generated from DNSPolicy. Duplicated search paths will be removed.
	Searches []string `json:"searches,omitempty"`

	// A list of DNS resolver options. This will be merged with the base options generated from DNSPolicy.
	// Duplicated entries will be removed. Resolution options given in Options will override those that appear in the base DNSPolicy.
	Options []PodDNSConfigOption `json:"options,omitempty" mapper:"dnsOptions"`
}

type PodConfig struct {
	// List of containers belonging to the pod. Containers cannot currently be added or removed. There must be at least one container in a Pod. Cannot be updated.
	Sidecars []NamedContainer `json:"containers,omitempty"`

	// Specifies the hostname of the Pod If not specified, the pod's hostname will be set to a system-defined value.
	Hostname string `json:"hostname,omitempty"`

	// HostAliases is an optional list of hosts and IPs that will be injected into the pod's hosts file if specified. This is only valid for non-hostNetwork pods.
	HostAliases []v1.HostAlias `json:"hostAliases,omitempty" mapper:"hosts,envmap=sep==,alias=hosts"`

	// Host networking requested for this pod. Use the host's network namespace. If this option is set, the ports that will be used must be specified. Default to false.
	HostNetwork bool `json:"hostNetwork,omitempty" mapper:"hostNetwork"`

	// Image pull secret
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty" mapper:"alias=pullSecrets"`

	// Volumes to create per replica
	VolumeTemplates []VolumeTemplate `json:"volumeTemplates,omitempty"`

	// DNS settings for this Pod
	DNS *DNS `json:"dns,omitempty"`

	*v1.Affinity `json:",inline"`

	Container
}

type Protocol string

const (
	ProtocolTCP   Protocol = "TCP"
	ProtocolUDP   Protocol = "UDP"
	ProtocolSCTP  Protocol = "SCTP"
	ProtocolHTTP  Protocol = "HTTP"
	ProtocolHTTP2 Protocol = "HTTP2"
	ProtocolGRPC  Protocol = "GRPC"
)

type ContainerPort struct {
	Name string `json:"name,omitempty"`
	// Expose will make the port available outside the cluster. All http/https ports will be set to true by default
	// if Expose is nil.  All other protocols are set to false by default
	Expose     *bool    `json:"expose,omitempty"`
	Protocol   Protocol `json:"protocol,omitempty"`
	Port       int32    `json:"port"`
	TargetPort int32    `json:"targetPort,omitempty"`
	HostPort   bool     `json:"hostport,omitempty"`
}

func (in ContainerPort) IsHTTP() bool {
	return in.Protocol == "" || in.Protocol == ProtocolHTTP || in.Protocol == ProtocolHTTP2
}

func (in ContainerPort) IsExposed() bool {
	if in.Expose != nil {
		return *in.Expose
	}
	return in.IsHTTP()
}

type ServiceStatus struct {
	// DeploymentReady for ready status on deployment
	DeploymentReady bool `json:"deploymentReady,omitempty"`

	// ScaleStatus for the Service
	ScaleStatus *ScaleStatus `json:"scaleStatus,omitempty"`

	// ComputedApp is the calculated value of Spec.App if not set
	ComputedApp string `json:"computedApp,omitempty"`

	// ComputedVersion is the calculated value of Spec.Version if not set
	ComputedVersion string `json:"computedVersion,omitempty"`

	// ComputedReplicas is calculated from autoscaling component to make sure pod has the desired load
	ComputedReplicas *int `json:"computedReplicas,omitempty"`

	// ComputedWeight is the weight calculated from the rollout revision
	ComputedWeight *int `json:"computedWeight,omitempty"`

	// ContainerRevision are populated from builds to store commits for each repo
	ContainerRevision map[string]BuildRevision `json:"containerRevision,omitempty"`

	// GeneratedServices contains all the service names are generated from build template
	GeneratedServices map[string]bool `json:"generatedServices,omitempty"`

	// GitCommits contains all git commits that triggers template update
	GitCommits []string `json:"gitCommits,omitempty"`

	// ShouldGenerate contains the serviceName that should be generated on the next controller run
	ShouldGenerate string `json:"shouldGenerate,omitempty"`

	// ShouldClean contains all the services that are generated from template but should be cleaned up.
	ShouldClean map[string]bool `json:"shouldClean,omitempty"`

	// Represents the latest available observations of a deployment's current state.
	Conditions []GenericCondition `json:"conditions,omitempty"`

	// The Endpoints to access this version directly
	Endpoints []string `json:"endpoints,omitempty" column:"name=Endpoint,type=string,jsonpath=.status.endpoints[0]"`

	// The Endpoints to access this service as part of an app
	AppEndpoints []string `json:"appEndpoints,omitempty"`

	// log token to access build log
	BuildLogToken string `json:"buildLogToken,omitempty"`

	// Watch represents if a service should creates git watcher to watch git changes
	Watch bool `json:"watch,omitempty"`
}

type ImageBuildSpec struct {
	// Repository url
	Repo string `json:"repo,omitempty"`

	// Repo Revision. Can be a git commit or tag
	Revision string `json:"revision,omitempty"`

	// Repo Branch. If specified, a gitmodule will be created to watch the repo and creating new revision if new commit or tag is pushed.
	Branch string `json:"branch,omitempty"`

	// Specify the name of the Dockerfile in the Repo. This is the full path relative to the repo root. Defaults to `Dockerfile`.
	Dockerfile string `json:"dockerfile,omitempty"`

	// Specify build context. Defaults to "."
	Context string `json:"context,omitempty"`

	// Specify build args
	Args []string `json:"args,omitempty" mapper:"alias=arg"`

	// Specify the build template. Defaults to `buildkit`.
	Template string `json:"template,omitempty"`

	// Specify the github secret name. Used to create Github webhook, the secret key has to be `accessToken`
	WebhookSecretName string `json:"webhookSecretName,omitempty"`

	// Specify secret name for checking our git resources
	CloneSecretName string `json:"cloneSecretName,omitempty"`

	// Specify custom registry to push the image instead of built-in one
	PushRegistry string `json:"pushRegistry,omitempty"`

	// Specify secret for pushing to custom registry
	PushRegistrySecretName string `json:"pushRegistrySecretName,omitempty"`

	// Specify image name instead of the one generated from service name, format: $registry/$imageName:$revision
	ImageName string `json:"imageName,omitempty"`

	// Whether to enable builds for pull requests
	PR bool `json:"pr,omitempty" mapper:"alias=onPR"`

	// Whether to enable builds for tags
	Tag bool `json:"tag,omitempty" mapper:"alias=onTag"`

	// Match string that includes tags that match
	TagIncludeRegexp string `json:"tagInclude,omitempty"`

	// Match string that excludes tags which match
	TagExcludeRegexp string `json:"tagExclude,omitempty"`

	// Build image with no cache
	NoCache bool `json:"noCache,omitempty"`

	// TimeoutSeconds describes how long the build can run
	TimeoutSeconds *int `json:"timeout,omitempty" mapper:"duration"`
}

type Permission struct {
	Role         string   `json:"role,omitempty"`
	Verbs        []string `json:"verbs,omitempty"`
	APIGroup     string   `json:"apiGroup,omitempty"`
	Resource     string   `json:"resource,omitempty"`
	URL          string   `json:"url,omitempty"`
	ResourceName string   `json:"resourceName,omitempty"`
}

type ScaleStatus struct {
	// Total number of unavailable pods targeted by this deployment. This is the total number of pods that are still required for the deployment to have 100% available capacity.
	// They may either be pods that are running but not yet available or pods that still have not been created.
	Unavailable int `json:"unavailable,omitempty"`

	// Total number of available pods (ready for at least minReadySeconds) targeted by this deployment.
	Available int `json:"available,omitempty"`
}

type BuildRevision struct {
	Commits []string `json:"commits,omitempty"`
}

type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceSpec   `json:"spec,omitempty"`
	Status ServiceStatus `json:"status,omitempty"`
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AutoscaleConfig) DeepCopyInto(out *AutoscaleConfig) {
	*out = *in
	if in.MinReplicas != nil {
		in, out := &in.MinReplicas, &out.MinReplicas
		*out = new(int32)
		**out = **in
	}
	if in.MaxReplicas != nil {
		in, out := &in.MaxReplicas, &out.MaxReplicas
		*out = new(int32)
		**out = **in
	}
	return
}

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AutoscaleConfig.
// func (in *AutoscaleConfig) DeepCopy() *AutoscaleConfig {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(AutoscaleConfig)
// 	in.DeepCopyInto(out)
// 	return out
// }

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BuildRevision) DeepCopyInto(out *BuildRevision) {
	*out = *in
	if in.Commits != nil {
		in, out := &in.Commits, &out.Commits
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BuildRevision.
func (in *BuildRevision) DeepCopy() *BuildRevision {
	if in == nil {
		return nil
	}
	out := new(BuildRevision)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Container) DeepCopyInto(out *Container) {
	*out = *in
	if in.ImageBuild != nil {
		in, out := &in.ImageBuild, &out.ImageBuild
		*out = new(ImageBuildSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Command != nil {
		in, out := &in.Command, &out.Command
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Args != nil {
		in, out := &in.Args, &out.Args
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Ports != nil {
		in, out := &in.Ports, &out.Ports
		*out = make([]ContainerPort, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]EnvVar, len(*in))
		copy(*out, *in)
	}
	if in.CPUMillis != nil {
		in, out := &in.CPUMillis, &out.CPUMillis
		*out = new(int64)
		**out = **in
	}
	if in.MemoryBytes != nil {
		in, out := &in.MemoryBytes, &out.MemoryBytes
		*out = new(int64)
		**out = **in
	}
	if in.Secrets != nil {
		in, out := &in.Secrets, &out.Secrets
		*out = make([]DataMount, len(*in))
		copy(*out, *in)
	}
	if in.Configs != nil {
		in, out := &in.Configs, &out.Configs
		*out = make([]DataMount, len(*in))
		copy(*out, *in)
	}
	if in.LivenessProbe != nil {
		in, out := &in.LivenessProbe, &out.LivenessProbe
		*out = new(v1.Probe)
		(*in).DeepCopyInto(*out)
	}
	if in.ReadinessProbe != nil {
		in, out := &in.ReadinessProbe, &out.ReadinessProbe
		*out = new(v1.Probe)
		(*in).DeepCopyInto(*out)
	}
	if in.Volumes != nil {
		in, out := &in.Volumes, &out.Volumes
		*out = make([]Volume, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ContainerSecurityContext != nil {
		in, out := &in.ContainerSecurityContext, &out.ContainerSecurityContext
		*out = new(ContainerSecurityContext)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Container.
func (in *Container) DeepCopy() *Container {
	if in == nil {
		return nil
	}
	out := new(Container)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ContainerPort) DeepCopyInto(out *ContainerPort) {
	*out = *in
	if in.Expose != nil {
		in, out := &in.Expose, &out.Expose
		*out = new(bool)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ContainerPort.
func (in *ContainerPort) DeepCopy() *ContainerPort {
	if in == nil {
		return nil
	}
	out := new(ContainerPort)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ContainerSecurityContext) DeepCopyInto(out *ContainerSecurityContext) {
	*out = *in
	if in.RunAsUser != nil {
		in, out := &in.RunAsUser, &out.RunAsUser
		*out = new(int64)
		**out = **in
	}
	if in.RunAsGroup != nil {
		in, out := &in.RunAsGroup, &out.RunAsGroup
		*out = new(int64)
		**out = **in
	}
	if in.ReadOnlyRootFilesystem != nil {
		in, out := &in.ReadOnlyRootFilesystem, &out.ReadOnlyRootFilesystem
		*out = new(bool)
		**out = **in
	}
	if in.Privileged != nil {
		in, out := &in.Privileged, &out.Privileged
		*out = new(bool)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ContainerSecurityContext.
func (in *ContainerSecurityContext) DeepCopy() *ContainerSecurityContext {
	if in == nil {
		return nil
	}
	out := new(ContainerSecurityContext)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DNS) DeepCopyInto(out *DNS) {
	*out = *in
	if in.Nameservers != nil {
		in, out := &in.Nameservers, &out.Nameservers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Searches != nil {
		in, out := &in.Searches, &out.Searches
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Options != nil {
		in, out := &in.Options, &out.Options
		*out = make([]PodDNSConfigOption, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DNS.
func (in *DNS) DeepCopy() *DNS {
	if in == nil {
		return nil
	}
	out := new(DNS)
	in.DeepCopyInto(out)
	return out
}

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *DataMount) DeepCopyInto(out *DataMount) {
// 	*out = *in
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataMount.
// func (in *DataMount) DeepCopy() *DataMount {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(DataMount)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *Destination) DeepCopyInto(out *Destination) {
// 	*out = *in
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Destination.
// func (in *Destination) DeepCopy() *Destination {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(Destination)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *EnvVar) DeepCopyInto(out *EnvVar) {
// 	*out = *in
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvVar.
// func (in *EnvVar) DeepCopy() *EnvVar {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(EnvVar)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *ExternalService) DeepCopyInto(out *ExternalService) {
// 	*out = *in
// 	out.TypeMeta = in.TypeMeta
// 	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
// 	in.Spec.DeepCopyInto(&out.Spec)
// 	in.Status.DeepCopyInto(&out.Status)
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExternalService.
// func (in *ExternalService) DeepCopy() *ExternalService {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(ExternalService)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
// func (in *ExternalService) DeepCopyObject() runtime.Object {
// 	if c := in.DeepCopy(); c != nil {
// 		return c
// 	}
// 	return nil
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *ExternalServiceList) DeepCopyInto(out *ExternalServiceList) {
// 	*out = *in
// 	out.TypeMeta = in.TypeMeta
// 	in.ListMeta.DeepCopyInto(&out.ListMeta)
// 	if in.Items != nil {
// 		in, out := &in.Items, &out.Items
// 		*out = make([]ExternalService, len(*in))
// 		for i := range *in {
// 			(*in)[i].DeepCopyInto(&(*out)[i])
// 		}
// 	}
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExternalServiceList.
// func (in *ExternalServiceList) DeepCopy() *ExternalServiceList {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(ExternalServiceList)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
// func (in *ExternalServiceList) DeepCopyObject() runtime.Object {
// 	if c := in.DeepCopy(); c != nil {
// 		return c
// 	}
// 	return nil
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *ExternalServiceSpec) DeepCopyInto(out *ExternalServiceSpec) {
// 	*out = *in
// 	if in.IPAddresses != nil {
// 		in, out := &in.IPAddresses, &out.IPAddresses
// 		*out = make([]string, len(*in))
// 		copy(*out, *in)
// 	}
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExternalServiceSpec.
// func (in *ExternalServiceSpec) DeepCopy() *ExternalServiceSpec {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(ExternalServiceSpec)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *ExternalServiceStatus) DeepCopyInto(out *ExternalServiceStatus) {
// 	*out = *in
// 	if in.Conditions != nil {
// 		in, out := &in.Conditions, &out.Conditions
// 		*out = make([]genericcondition.GenericCondition, len(*in))
// 		copy(*out, *in)
// 	}
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExternalServiceStatus.
// func (in *ExternalServiceStatus) DeepCopy() *ExternalServiceStatus {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(ExternalServiceStatus)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *Fault) DeepCopyInto(out *Fault) {
// 	*out = *in
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Fault.
// func (in *Fault) DeepCopy() *Fault {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(Fault)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *HeaderMatch) DeepCopyInto(out *HeaderMatch) {
// 	*out = *in
// 	if in.Value != nil {
// 		in, out := &in.Value, &out.Value
// 		*out = new(StringMatch)
// 		**out = **in
// 	}
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HeaderMatch.
// func (in *HeaderMatch) DeepCopy() *HeaderMatch {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(HeaderMatch)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *HeaderOperations) DeepCopyInto(out *HeaderOperations) {
// 	*out = *in
// 	if in.Add != nil {
// 		in, out := &in.Add, &out.Add
// 		*out = make([]NameValue, len(*in))
// 		copy(*out, *in)
// 	}
// 	if in.Set != nil {
// 		in, out := &in.Set, &out.Set
// 		*out = make([]NameValue, len(*in))
// 		copy(*out, *in)
// 	}
// 	if in.Remove != nil {
// 		in, out := &in.Remove, &out.Remove
// 		*out = make([]string, len(*in))
// 		copy(*out, *in)
// 	}
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HeaderOperations.
// func (in *HeaderOperations) DeepCopy() *HeaderOperations {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(HeaderOperations)
// 	in.DeepCopyInto(out)
// 	return out
// }

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ImageBuildSpec) DeepCopyInto(out *ImageBuildSpec) {
	*out = *in
	if in.Args != nil {
		in, out := &in.Args, &out.Args
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.TimeoutSeconds != nil {
		in, out := &in.TimeoutSeconds, &out.TimeoutSeconds
		*out = new(int)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ImageBuildSpec.
func (in *ImageBuildSpec) DeepCopy() *ImageBuildSpec {
	if in == nil {
		return nil
	}
	out := new(ImageBuildSpec)
	in.DeepCopyInto(out)
	return out
}

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *Match) DeepCopyInto(out *Match) {
// 	*out = *in
// 	if in.Path != nil {
// 		in, out := &in.Path, &out.Path
// 		*out = new(StringMatch)
// 		**out = **in
// 	}
// 	if in.Schema != nil {
// 		in, out := &in.Schema, &out.Schema
// 		*out = new(StringMatch)
// 		**out = **in
// 	}
// 	if in.Methods != nil {
// 		in, out := &in.Methods, &out.Methods
// 		*out = make([]string, len(*in))
// 		copy(*out, *in)
// 	}
// 	if in.Headers != nil {
// 		in, out := &in.Headers, &out.Headers
// 		*out = make([]HeaderMatch, len(*in))
// 		for i := range *in {
// 			(*in)[i].DeepCopyInto(&(*out)[i])
// 		}
// 	}
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Match.
// func (in *Match) DeepCopy() *Match {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(Match)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *NameValue) DeepCopyInto(out *NameValue) {
// 	*out = *in
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NameValue.
// func (in *NameValue) DeepCopy() *NameValue {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(NameValue)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NamedContainer) DeepCopyInto(out *NamedContainer) {
	*out = *in
	in.Container.DeepCopyInto(&out.Container)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NamedContainer.
func (in *NamedContainer) DeepCopy() *NamedContainer {
	if in == nil {
		return nil
	}
	out := new(NamedContainer)
	in.DeepCopyInto(out)
	return out
}

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Permission) DeepCopyInto(out *Permission) {
	*out = *in
	if in.Verbs != nil {
		in, out := &in.Verbs, &out.Verbs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Permission.
func (in *Permission) DeepCopy() *Permission {
	if in == nil {
		return nil
	}
	out := new(Permission)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodConfig) DeepCopyInto(out *PodConfig) {
	*out = *in
	if in.Sidecars != nil {
		in, out := &in.Sidecars, &out.Sidecars
		*out = make([]NamedContainer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.HostAliases != nil {
		in, out := &in.HostAliases, &out.HostAliases
		*out = make([]v1.HostAlias, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ImagePullSecrets != nil {
		in, out := &in.ImagePullSecrets, &out.ImagePullSecrets
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.VolumeTemplates != nil {
		in, out := &in.VolumeTemplates, &out.VolumeTemplates
		*out = make([]VolumeTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.DNS != nil {
		in, out := &in.DNS, &out.DNS
		*out = new(DNS)
		(*in).DeepCopyInto(*out)
	}
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(v1.Affinity)
		(*in).DeepCopyInto(*out)
	}
	in.Container.DeepCopyInto(&out.Container)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodConfig.
func (in *PodConfig) DeepCopy() *PodConfig {
	if in == nil {
		return nil
	}
	out := new(PodConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodDNSConfigOption) DeepCopyInto(out *PodDNSConfigOption) {
	*out = *in
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodDNSConfigOption.
func (in *PodDNSConfigOption) DeepCopy() *PodDNSConfigOption {
	if in == nil {
		return nil
	}
	out := new(PodDNSConfigOption)
	in.DeepCopyInto(out)
	return out
}

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *Question) DeepCopyInto(out *Question) {
// 	*out = *in
// 	if in.Options != nil {
// 		in, out := &in.Options, &out.Options
// 		*out = make([]string, len(*in))
// 		copy(*out, *in)
// 	}
// 	if in.Subquestions != nil {
// 		in, out := &in.Subquestions, &out.Subquestions
// 		*out = make([]SubQuestion, len(*in))
// 		for i := range *in {
// 			(*in)[i].DeepCopyInto(&(*out)[i])
// 		}
// 	}
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Question.
// func (in *Question) DeepCopy() *Question {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(Question)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *Redirect) DeepCopyInto(out *Redirect) {
// 	*out = *in
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Redirect.
// func (in *Redirect) DeepCopy() *Redirect {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(Redirect)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *Retry) DeepCopyInto(out *Retry) {
// 	*out = *in
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Retry.
// func (in *Retry) DeepCopy() *Retry {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(Retry)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *Rewrite) DeepCopyInto(out *Rewrite) {
// 	*out = *in
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Rewrite.
// func (in *Rewrite) DeepCopy() *Rewrite {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(Rewrite)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *RolloutConfig) DeepCopyInto(out *RolloutConfig) {
// 	*out = *in
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RolloutConfig.
// func (in *RolloutConfig) DeepCopy() *RolloutConfig {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(RolloutConfig)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *RouteSpec) DeepCopyInto(out *RouteSpec) {
// 	*out = *in
// 	in.Match.DeepCopyInto(&out.Match)
// 	if in.To != nil {
// 		in, out := &in.To, &out.To
// 		*out = make([]WeightedDestination, len(*in))
// 		copy(*out, *in)
// 	}
// 	if in.Redirect != nil {
// 		in, out := &in.Redirect, &out.Redirect
// 		*out = new(Redirect)
// 		**out = **in
// 	}
// 	if in.Rewrite != nil {
// 		in, out := &in.Rewrite, &out.Rewrite
// 		*out = new(Rewrite)
// 		**out = **in
// 	}
// 	if in.Retry != nil {
// 		in, out := &in.Retry, &out.Retry
// 		*out = new(Retry)
// 		**out = **in
// 	}
// 	if in.Headers != nil {
// 		in, out := &in.Headers, &out.Headers
// 		*out = new(HeaderOperations)
// 		(*in).DeepCopyInto(*out)
// 	}
// 	if in.Fault != nil {
// 		in, out := &in.Fault, &out.Fault
// 		*out = new(Fault)
// 		**out = **in
// 	}
// 	if in.Mirror != nil {
// 		in, out := &in.Mirror, &out.Mirror
// 		*out = new(Destination)
// 		**out = **in
// 	}
// 	if in.TimeoutSeconds != nil {
// 		in, out := &in.TimeoutSeconds, &out.TimeoutSeconds
// 		*out = new(int)
// 		**out = **in
// 	}
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouteSpec.
// func (in *RouteSpec) DeepCopy() *RouteSpec {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(RouteSpec)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *Router) DeepCopyInto(out *Router) {
// 	*out = *in
// 	out.TypeMeta = in.TypeMeta
// 	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
// 	in.Spec.DeepCopyInto(&out.Spec)
// 	in.Status.DeepCopyInto(&out.Status)
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Router.
// func (in *Router) DeepCopy() *Router {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(Router)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
// func (in *Router) DeepCopyObject() runtime.Object {
// 	if c := in.DeepCopy(); c != nil {
// 		return c
// 	}
// 	return nil
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *RouterList) DeepCopyInto(out *RouterList) {
// 	*out = *in
// 	out.TypeMeta = in.TypeMeta
// 	in.ListMeta.DeepCopyInto(&out.ListMeta)
// 	if in.Items != nil {
// 		in, out := &in.Items, &out.Items
// 		*out = make([]Router, len(*in))
// 		for i := range *in {
// 			(*in)[i].DeepCopyInto(&(*out)[i])
// 		}
// 	}
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouterList.
// func (in *RouterList) DeepCopy() *RouterList {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(RouterList)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
// func (in *RouterList) DeepCopyObject() runtime.Object {
// 	if c := in.DeepCopy(); c != nil {
// 		return c
// 	}
// 	return nil
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *RouterSpec) DeepCopyInto(out *RouterSpec) {
// 	*out = *in
// 	if in.Routes != nil {
// 		in, out := &in.Routes, &out.Routes
// 		*out = make([]RouteSpec, len(*in))
// 		for i := range *in {
// 			(*in)[i].DeepCopyInto(&(*out)[i])
// 		}
// 	}
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouterSpec.
// func (in *RouterSpec) DeepCopy() *RouterSpec {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(RouterSpec)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *RouterStatus) DeepCopyInto(out *RouterStatus) {
// 	*out = *in
// 	if in.Endpoints != nil {
// 		in, out := &in.Endpoints, &out.Endpoints
// 		*out = make([]string, len(*in))
// 		copy(*out, *in)
// 	}
// 	if in.Conditions != nil {
// 		in, out := &in.Conditions, &out.Conditions
// 		*out = make([]genericcondition.GenericCondition, len(*in))
// 		copy(*out, *in)
// 	}
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouterStatus.
// func (in *RouterStatus) DeepCopy() *RouterStatus {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(RouterStatus)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *ScaleStatus) DeepCopyInto(out *ScaleStatus) {
// 	*out = *in
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScaleStatus.
// func (in *ScaleStatus) DeepCopy() *ScaleStatus {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(ScaleStatus)
// 	in.DeepCopyInto(out)
// 	return out
// }

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Service) DeepCopyInto(out *Service) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Service.
func (in *Service) DeepCopy() *Service {
	if in == nil {
		return nil
	}
	out := new(Service)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Service) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *ServiceList) DeepCopyInto(out *ServiceList) {
// 	*out = *in
// 	out.TypeMeta = in.TypeMeta
// 	in.ListMeta.DeepCopyInto(&out.ListMeta)
// 	if in.Items != nil {
// 		in, out := &in.Items, &out.Items
// 		*out = make([]Service, len(*in))
// 		for i := range *in {
// 			(*in)[i].DeepCopyInto(&(*out)[i])
// 		}
// 	}
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceList.
// func (in *ServiceList) DeepCopy() *ServiceList {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(ServiceList)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
// func (in *ServiceList) DeepCopyObject() runtime.Object {
// 	if c := in.DeepCopy(); c != nil {
// 		return c
// 	}
// 	return nil
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceSpec) DeepCopyInto(out *ServiceSpec) {
	*out = *in
	in.PodConfig.DeepCopyInto(&out.PodConfig)
	if in.Weight != nil {
		in, out := &in.Weight, &out.Weight
		*out = new(int)
		**out = **in
	}
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int)
		**out = **in
	}
	if in.MaxUnavailable != nil {
		in, out := &in.MaxUnavailable, &out.MaxUnavailable
		*out = new(intstr.IntOrString)
		**out = **in
	}
	if in.MaxSurge != nil {
		in, out := &in.MaxSurge, &out.MaxSurge
		*out = new(intstr.IntOrString)
		**out = **in
	}
	if in.Autoscale != nil {
		in, out := &in.Autoscale, &out.Autoscale
		*out = new(AutoscaleConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.RolloutDuration != nil {
		in, out := &in.RolloutDuration, &out.RolloutDuration
		*out = new(metav1.Duration)
		**out = **in
	}
	if in.RolloutConfig != nil {
		in, out := &in.RolloutConfig, &out.RolloutConfig
		*out = new(RolloutConfig)
		**out = **in
	}
	if in.ServiceMesh != nil {
		in, out := &in.ServiceMesh, &out.ServiceMesh
		*out = new(bool)
		**out = **in
	}
	if in.RequestTimeoutSeconds != nil {
		in, out := &in.RequestTimeoutSeconds, &out.RequestTimeoutSeconds
		*out = new(int)
		**out = **in
	}
	if in.Permissions != nil {
		in, out := &in.Permissions, &out.Permissions
		*out = make([]Permission, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.GlobalPermissions != nil {
		in, out := &in.GlobalPermissions, &out.GlobalPermissions
		*out = make([]Permission, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceSpec.
func (in *ServiceSpec) DeepCopy() *ServiceSpec {
	if in == nil {
		return nil
	}
	out := new(ServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceStatus) DeepCopyInto(out *ServiceStatus) {
	*out = *in
	if in.ScaleStatus != nil {
		in, out := &in.ScaleStatus, &out.ScaleStatus
		*out = new(ScaleStatus)
		**out = **in
	}
	if in.ComputedReplicas != nil {
		in, out := &in.ComputedReplicas, &out.ComputedReplicas
		*out = new(int)
		**out = **in
	}
	if in.ComputedWeight != nil {
		in, out := &in.ComputedWeight, &out.ComputedWeight
		*out = new(int)
		**out = **in
	}
	if in.ContainerRevision != nil {
		in, out := &in.ContainerRevision, &out.ContainerRevision
		*out = make(map[string]BuildRevision, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.GeneratedServices != nil {
		in, out := &in.GeneratedServices, &out.GeneratedServices
		*out = make(map[string]bool, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.GitCommits != nil {
		in, out := &in.GitCommits, &out.GitCommits
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ShouldClean != nil {
		in, out := &in.ShouldClean, &out.ShouldClean
		*out = make(map[string]bool, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]GenericCondition, len(*in))
		copy(*out, *in)
	}
	if in.Endpoints != nil {
		in, out := &in.Endpoints, &out.Endpoints
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AppEndpoints != nil {
		in, out := &in.AppEndpoints, &out.AppEndpoints
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceStatus.
func (in *ServiceStatus) DeepCopy() *ServiceStatus {
	if in == nil {
		return nil
	}
	out := new(ServiceStatus)
	in.DeepCopyInto(out)
	return out
}

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *Stack) DeepCopyInto(out *Stack) {
// 	*out = *in
// 	out.TypeMeta = in.TypeMeta
// 	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
// 	in.Spec.DeepCopyInto(&out.Spec)
// 	in.Status.DeepCopyInto(&out.Status)
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Stack.
// func (in *Stack) DeepCopy() *Stack {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(Stack)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
// func (in *Stack) DeepCopyObject() runtime.Object {
// 	if c := in.DeepCopy(); c != nil {
// 		return c
// 	}
// 	return nil
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *StackBuild) DeepCopyInto(out *StackBuild) {
// 	*out = *in
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackBuild.
// func (in *StackBuild) DeepCopy() *StackBuild {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(StackBuild)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *StackList) DeepCopyInto(out *StackList) {
// 	*out = *in
// 	out.TypeMeta = in.TypeMeta
// 	in.ListMeta.DeepCopyInto(&out.ListMeta)
// 	if in.Items != nil {
// 		in, out := &in.Items, &out.Items
// 		*out = make([]Stack, len(*in))
// 		for i := range *in {
// 			(*in)[i].DeepCopyInto(&(*out)[i])
// 		}
// 	}
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackList.
// func (in *StackList) DeepCopy() *StackList {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(StackList)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
// func (in *StackList) DeepCopyObject() runtime.Object {
// 	if c := in.DeepCopy(); c != nil {
// 		return c
// 	}
// 	return nil
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *StackSpec) DeepCopyInto(out *StackSpec) {
// 	*out = *in
// 	if in.Build != nil {
// 		in, out := &in.Build, &out.Build
// 		*out = new(StackBuild)
// 		**out = **in
// 	}
// 	if in.Permissions != nil {
// 		in, out := &in.Permissions, &out.Permissions
// 		*out = make([]Permission, len(*in))
// 		for i := range *in {
// 			(*in)[i].DeepCopyInto(&(*out)[i])
// 		}
// 	}
// 	if in.AdditionalGroupVersionKinds != nil {
// 		in, out := &in.AdditionalGroupVersionKinds, &out.AdditionalGroupVersionKinds
// 		*out = make([]schema.GroupVersionKind, len(*in))
// 		copy(*out, *in)
// 	}
// 	if in.Answers != nil {
// 		in, out := &in.Answers, &out.Answers
// 		*out = make(map[string]string, len(*in))
// 		for key, val := range *in {
// 			(*out)[key] = val
// 		}
// 	}
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackSpec.
// func (in *StackSpec) DeepCopy() *StackSpec {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(StackSpec)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *StackStatus) DeepCopyInto(out *StackStatus) {
// 	*out = *in
// 	if in.Conditions != nil {
// 		in, out := &in.Conditions, &out.Conditions
// 		*out = make([]genericcondition.GenericCondition, len(*in))
// 		copy(*out, *in)
// 	}
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackStatus.
// func (in *StackStatus) DeepCopy() *StackStatus {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(StackStatus)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *StringMatch) DeepCopyInto(out *StringMatch) {
// 	*out = *in
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StringMatch.
// func (in *StringMatch) DeepCopy() *StringMatch {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(StringMatch)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *SubQuestion) DeepCopyInto(out *SubQuestion) {
// 	*out = *in
// 	if in.Options != nil {
// 		in, out := &in.Options, &out.Options
// 		*out = make([]string, len(*in))
// 		copy(*out, *in)
// 	}
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubQuestion.
// func (in *SubQuestion) DeepCopy() *SubQuestion {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(SubQuestion)
// 	in.DeepCopyInto(out)
// 	return out
// }

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *TemplateMeta) DeepCopyInto(out *TemplateMeta) {
// 	*out = *in
// 	if in.Questions != nil {
// 		in, out := &in.Questions, &out.Questions
// 		*out = make([]Question, len(*in))
// 		for i := range *in {
// 			(*in)[i].DeepCopyInto(&(*out)[i])
// 		}
// 	}
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TemplateMeta.
// func (in *TemplateMeta) DeepCopy() *TemplateMeta {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(TemplateMeta)
// 	in.DeepCopyInto(out)
// 	return out
// }

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Volume) DeepCopyInto(out *Volume) {
	*out = *in
	if in.HostPathType != nil {
		in, out := &in.HostPathType, &out.HostPathType
		*out = new(v1.HostPathType)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Volume.
func (in *Volume) DeepCopy() *Volume {
	if in == nil {
		return nil
	}
	out := new(Volume)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeTemplate) DeepCopyInto(out *VolumeTemplate) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.AccessModes != nil {
		in, out := &in.AccessModes, &out.AccessModes
		*out = make([]v1.PersistentVolumeAccessMode, len(*in))
		copy(*out, *in)
	}
	if in.VolumeMode != nil {
		in, out := &in.VolumeMode, &out.VolumeMode
		*out = new(v1.PersistentVolumeMode)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeTemplate.
func (in *VolumeTemplate) DeepCopy() *VolumeTemplate {
	if in == nil {
		return nil
	}
	out := new(VolumeTemplate)
	in.DeepCopyInto(out)
	return out
}

// // DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
// func (in *WeightedDestination) DeepCopyInto(out *WeightedDestination) {
// 	*out = *in
// 	out.Destination = in.Destination
// 	return
// }

// // DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WeightedDestination.
// func (in *WeightedDestination) DeepCopy() *WeightedDestination {
// 	if in == nil {
// 		return nil
// 	}
// 	out := new(WeightedDestination)
// 	in.DeepCopyInto(out)
// 	return out
// }

// type WeightedDestination struct {
// 	Destination

// 	// Weight for the Destination
// 	Weight int `json:"weight,omitempty"`
// }

// type Destination struct {
// 	// Destination Service
// 	App string `json:"app,omitempty"`

// 	// Destination Revision
// 	Version string `json:"version,omitempty"`

// 	// Destination Port
// 	Port uint32 `json:"port,omitempty"`
// }

// // ExternalService creates a DNS record and route rules for any Service outside of the cluster, can be IPs or FQDN outside the mesh
// type ExternalService struct {
// 	metav1.TypeMeta   `json:",inline"`
// 	metav1.ObjectMeta `json:"metadata,omitempty"`
// 	Spec              ExternalServiceSpec   `json:"spec,omitempty"`
// 	Status            ExternalServiceStatus `json:"status,omitempty"`
// }

// type ExternalServiceSpec struct {
// 	// External service located outside the mesh, represented by IPs
// 	IPAddresses []string `json:"ipAddresses,omitempty"`

// 	// External service located outside the mesh, represented by DNS
// 	FQDN string `json:"fqdn,omitempty"`

// 	// In-Mesh app in another namespace
// 	TargetApp string `json:"targetApp,omitempty"`

// 	// In-Mesh version in another namespace
// 	TargetVersion string `json:"targetVersion,omitempty"`

// 	// In-Mesh router in another namespace
// 	TargetRouter string `json:"targetRouter,omitempty"`

// 	// Namespace of in-mesh service in another namespace
// 	TargetServiceNamespace string `json:"targetServiceNamespace,omitempty"`
// }

// type ExternalServiceStatus struct {
// 	// Represents the latest available observations of a ExternalService's current state.
// 	Conditions []GenericCondition `json:"conditions,omitempty"`
// }

// // ExternalServiceList is a list of ExternalService resources
// type ExternalServiceList struct {
// 	metav1.TypeMeta `json:",inline"`
// 	metav1.ListMeta `json:"metadata"`

// 	Items []ExternalService `json:"items"`
// }

func setConfigDefaults(config *rest.Config) error {
	var SchemeGroupVersion = schema.GroupVersion{Group: "rio.cattle.io", Version: "v1"}
	gv := SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

func Get(name string, options metav1.GetOptions, client rest.RESTClient) (result *Service, err error) {
	result = &Service{}
	ctx, _ := context.WithCancel(context.Background())
	err = client.Get().
		Namespace("default").
		Resource("services").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

func main() {
	loader := GetInteractiveClientConfig("/Users/crro/.kube/config")

	defaultNs, _, err := loader.Namespace()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(defaultNs)

	restConfig, err := loader.ClientConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	restConfig.QPS = 500
	restConfig.Burst = 100
	if err := setConfigDefaults(restConfig); err != nil {
		fmt.Println(err.Error())
		return
	}
	client, err := rest.RESTClientFor(restConfig)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	s, err := Get("kubernetes", metav1.GetOptions{}, *client)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("s is", s)
}

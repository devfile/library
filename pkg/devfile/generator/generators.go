package generator

import (
	"fmt"

	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"

	"github.com/devfile/library/pkg/devfile/parser"
)

const (
	// DevfileSourceVolumeMount is the default directory to mount the volume in the container
	DevfileSourceVolumeMount = "/projects"

	// EnvProjectsRoot is the env defined for project mount in a component container when component's mountSources=true
	EnvProjectsRoot = "PROJECTS_ROOT"

	// EnvProjectsSrc is the env defined for path to the project source in a component container
	EnvProjectsSrc = "PROJECT_SOURCE"

	deploymentKind       = "Deployment"
	deploymentAPIVersion = "apps/v1"
)

// GetTypeMeta gets a type meta of the specified kind and version
func GetTypeMeta(kind string, APIVersion string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       kind,
		APIVersion: APIVersion,
	}
}

// GetObjectMeta gets an object meta with the parameters
func GetObjectMeta(name, namespace string, labels, annotations map[string]string) metav1.ObjectMeta {

	objectMeta := metav1.ObjectMeta{
		Name:        name,
		Namespace:   namespace,
		Labels:      labels,
		Annotations: annotations,
	}

	return objectMeta
}

// GetContainers iterates through the devfile components and returns a slice of the corresponding containers
func GetContainers(devfileObj parser.DevfileObj) ([]corev1.Container, error) {
	var containers []corev1.Container
	for _, comp := range devfileObj.Data.GetDevfileContainerComponents() {
		envVars := convertEnvs(comp.Container.Env)
		resourceReqs := getResourceReqs(comp)
		ports := convertPorts(comp.Container.Endpoints)
		containerParams := containerParams{
			Name:         comp.Name,
			Image:        comp.Container.Image,
			IsPrivileged: false,
			Command:      comp.Container.Command,
			Args:         comp.Container.Args,
			EnvVars:      envVars,
			ResourceReqs: resourceReqs,
			Ports:        ports,
		}
		container := getContainer(containerParams)

		// If `mountSources: true` was set PROJECTS_ROOT & PROJECT_SOURCE env
		if comp.Container.MountSources == nil || *comp.Container.MountSources {
			syncRootFolder := addSyncRootFolder(container, comp.Container.SourceMapping)

			err := addSyncFolder(container, syncRootFolder, devfileObj.Data.GetProjects())
			if err != nil {
				return nil, err
			}
		}
		containers = append(containers, *container)
	}
	return containers, nil
}

// PodTemplateSpecParams is a struct that contains the required data to create a pod template spec object
type PodTemplateSpecParams struct {
	ObjectMeta     metav1.ObjectMeta
	InitContainers []corev1.Container
	Containers     []corev1.Container
	Volumes        []corev1.Volume
}

// GetPodTemplateSpec gets a pod template spec that can be used to create a deployment spec
func GetPodTemplateSpec(podTemplateSpecParams PodTemplateSpecParams) *corev1.PodTemplateSpec {
	podTemplateSpec := &corev1.PodTemplateSpec{
		ObjectMeta: podTemplateSpecParams.ObjectMeta,
		Spec: corev1.PodSpec{
			InitContainers: podTemplateSpecParams.InitContainers,
			Containers:     podTemplateSpecParams.Containers,
			Volumes:        podTemplateSpecParams.Volumes,
		},
	}

	return podTemplateSpec
}

// DeploymentSpecParams is a struct that contains the required data to create a deployment spec object
type DeploymentSpecParams struct {
	PodTemplateSpec   corev1.PodTemplateSpec
	PodSelectorLabels map[string]string
}

// GetDeploymentSpec gets a deployment spec
func GetDeploymentSpec(deploySpecParams DeploymentSpecParams) *appsv1.DeploymentSpec {
	deploymentSpec := &appsv1.DeploymentSpec{
		Strategy: appsv1.DeploymentStrategy{
			Type: appsv1.RecreateDeploymentStrategyType,
		},
		Selector: &metav1.LabelSelector{
			MatchLabels: deploySpecParams.PodSelectorLabels,
		},
		Template: deploySpecParams.PodTemplateSpec,
	}

	return deploymentSpec
}

// DeploymentParams is a struct that contains the required data to create a deployment object
type DeploymentParams struct {
	TypeMeta       metav1.TypeMeta
	ObjectMeta     metav1.ObjectMeta
	DeploymentSpec appsv1.DeploymentSpec
}

// GetDeployment gets a deployment object
func GetDeployment(deployParams DeploymentParams) *appsv1.Deployment {

	deployment := &appsv1.Deployment{
		TypeMeta:   deployParams.TypeMeta,
		ObjectMeta: deployParams.ObjectMeta,
		Spec:       deployParams.DeploymentSpec,
	}

	return deployment
}

// GetPVCSpec gets a RWO pvc spec
func GetPVCSpec(quantity resource.Quantity) *corev1.PersistentVolumeClaimSpec {

	pvcSpec := &corev1.PersistentVolumeClaimSpec{
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: quantity,
			},
		},
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
	}

	return pvcSpec
}

// GetServiceSpec iterates through the devfile components and returns a ServiceSpec
func GetServiceSpec(devfileObj parser.DevfileObj, selectorLabels map[string]string) (*corev1.ServiceSpec, error) {

	var containerPorts []corev1.ContainerPort
	portExposureMap := getPortExposure(devfileObj)
	containers, err := GetContainers(devfileObj)
	if err != nil {
		return nil, err
	}
	for _, c := range containers {
		for _, port := range c.Ports {
			portExist := false
			for _, entry := range containerPorts {
				if entry.ContainerPort == port.ContainerPort {
					portExist = true
					break
				}
			}
			// if Exposure == none, should not create a service for that port
			if !portExist && portExposureMap[int(port.ContainerPort)] != v1.NoneEndpointExposure {
				port.Name = fmt.Sprintf("port-%v", port.ContainerPort)
				containerPorts = append(containerPorts, port)
			}
		}
	}
	serviceSpecParams := serviceSpecParams{
		ContainerPorts: containerPorts,
		SelectorLabels: selectorLabels,
	}

	return getServiceSpec(serviceSpecParams), nil
}

// ServiceParams is a struct that contains the required data to create a service object
type ServiceParams struct {
	TypeMeta    metav1.TypeMeta
	ObjectMeta  metav1.ObjectMeta
	ServiceSpec corev1.ServiceSpec
}

// GetService gets the service
func GetService(serviceParams ServiceParams) *corev1.Service {
	service := &corev1.Service{
		TypeMeta:   serviceParams.TypeMeta,
		ObjectMeta: serviceParams.ObjectMeta,
		Spec:       serviceParams.ServiceSpec,
	}

	return service
}

// IngressSpecParams struct for function GenerateIngressSpec
// serviceName is the name of the service for the target reference
// ingressDomain is the ingress domain to use for the ingress
// portNumber is the target port of the ingress
// Path is the path of the ingress
// TLSSecretName is the target TLS Secret name of the ingress
type IngressSpecParams struct {
	ServiceName   string
	IngressDomain string
	PortNumber    intstr.IntOrString
	TLSSecretName string
	Path          string
}

// GetIngressSpec gets an ingress spec
func GetIngressSpec(ingressSpecParams IngressSpecParams) *extensionsv1.IngressSpec {
	path := "/"
	if ingressSpecParams.Path != "" {
		path = ingressSpecParams.Path
	}
	ingressSpec := &extensionsv1.IngressSpec{
		Rules: []extensionsv1.IngressRule{
			{
				Host: ingressSpecParams.IngressDomain,
				IngressRuleValue: extensionsv1.IngressRuleValue{
					HTTP: &extensionsv1.HTTPIngressRuleValue{
						Paths: []extensionsv1.HTTPIngressPath{
							{
								Path: path,
								Backend: extensionsv1.IngressBackend{
									ServiceName: ingressSpecParams.ServiceName,
									ServicePort: ingressSpecParams.PortNumber,
								},
							},
						},
					},
				},
			},
		},
	}
	secretNameLength := len(ingressSpecParams.TLSSecretName)
	if secretNameLength != 0 {
		ingressSpec.TLS = []extensionsv1.IngressTLS{
			{
				Hosts: []string{
					ingressSpecParams.IngressDomain,
				},
				SecretName: ingressSpecParams.TLSSecretName,
			},
		}
	}

	return ingressSpec
}

// IngressParams is a struct that contains the required data to create an ingress object
type IngressParams struct {
	TypeMeta    metav1.TypeMeta
	ObjectMeta  metav1.ObjectMeta
	IngressSpec extensionsv1.IngressSpec
}

// GetIngress gets an ingress
func GetIngress(ingressParams IngressParams) *extensionsv1.Ingress {
	ingress := &extensionsv1.Ingress{
		TypeMeta:   ingressParams.TypeMeta,
		ObjectMeta: ingressParams.ObjectMeta,
		Spec:       ingressParams.IngressSpec,
	}

	return ingress
}

// RouteSpecParams struct for function GenerateRouteSpec
// serviceName is the name of the service for the target reference
// portNumber is the target port of the ingress
// Path is the path of the route
type RouteSpecParams struct {
	ServiceName string
	PortNumber  intstr.IntOrString
	Path        string
	Secure      bool
}

// GetRouteSpec gets a route spec
func GetRouteSpec(routeParams RouteSpecParams) *routev1.RouteSpec {
	routePath := "/"
	if routeParams.Path != "" {
		routePath = routeParams.Path
	}
	routeSpec := &routev1.RouteSpec{
		To: routev1.RouteTargetReference{
			Kind: "Service",
			Name: routeParams.ServiceName,
		},
		Port: &routev1.RoutePort{
			TargetPort: routeParams.PortNumber,
		},
		Path: routePath,
	}

	if routeParams.Secure {
		routeSpec.TLS = &routev1.TLSConfig{
			Termination:                   routev1.TLSTerminationEdge,
			InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
		}
	}

	return routeSpec
}

// RouteParams is a struct that contains the required data to create a route object
type RouteParams struct {
	TypeMeta   metav1.TypeMeta
	ObjectMeta metav1.ObjectMeta
	RouteSpec  routev1.RouteSpec
}

// GetRoute gets a route
func GetRoute(routeParams RouteParams) *routev1.Route {
	route := &routev1.Route{
		TypeMeta:   routeParams.TypeMeta,
		ObjectMeta: routeParams.ObjectMeta,
		Spec:       routeParams.RouteSpec,
	}

	return route
}

// GetOwnerReference generates an ownerReference  from the deployment which can then be set as
// owner for various Kubernetes objects and ensure that when the owner object is deleted from the
// cluster, all other objects are automatically removed by Kubernetes garbage collector
func GetOwnerReference(deployment *appsv1.Deployment) metav1.OwnerReference {

	ownerReference := metav1.OwnerReference{
		APIVersion: deploymentAPIVersion,
		Kind:       deploymentKind,
		Name:       deployment.Name,
		UID:        deployment.UID,
	}

	return ownerReference
}

// BuildConfigSpecParams is a struct to create build config spec
type BuildConfigSpecParams struct {
	CommonObjectMeta metav1.ObjectMeta
	GitURL           string
	GitRef           string
	BuildStrategy    buildv1.BuildStrategy
}

// GetBuildConfigSpec gets the build config spec and outputs the build to the image stream
func GetBuildConfigSpec(buildConfigSpecParams BuildConfigSpecParams) *buildv1.BuildConfigSpec {

	return &buildv1.BuildConfigSpec{
		CommonSpec: buildv1.CommonSpec{
			Output: buildv1.BuildOutput{
				To: &corev1.ObjectReference{
					Kind: "ImageStreamTag",
					Name: buildConfigSpecParams.CommonObjectMeta.Name + ":latest",
				},
			},
			Source: buildv1.BuildSource{
				Git: &buildv1.GitBuildSource{
					URI: buildConfigSpecParams.GitURL,
					Ref: buildConfigSpecParams.GitRef,
				},
				Type: buildv1.BuildSourceGit,
			},
			Strategy: buildConfigSpecParams.BuildStrategy,
		},
	}
}

// BuildConfigParams is a struct that contains the required data to create a build config object
type BuildConfigParams struct {
	TypeMeta        metav1.TypeMeta
	ObjectMeta      metav1.ObjectMeta
	BuildConfigSpec buildv1.BuildConfigSpec
}

// GetBuildConfig gets a build config
func GetBuildConfig(buildConfigParams BuildConfigParams) *buildv1.BuildConfig {
	buildConfig := &buildv1.BuildConfig{
		TypeMeta:   buildConfigParams.TypeMeta,
		ObjectMeta: buildConfigParams.ObjectMeta,
		Spec:       buildConfigParams.BuildConfigSpec,
	}

	return buildConfig
}

// GetSourceBuildStrategy gets the source build strategy
func GetSourceBuildStrategy(imageName, imageNamespace string) buildv1.BuildStrategy {
	return buildv1.BuildStrategy{
		SourceStrategy: &buildv1.SourceBuildStrategy{
			From: corev1.ObjectReference{
				Kind:      "ImageStreamTag",
				Name:      imageName,
				Namespace: imageNamespace,
			},
		},
	}
}

// ImageStreamParams is a struct that contains the required data to create an image stream object
type ImageStreamParams struct {
	TypeMeta        metav1.TypeMeta
	ObjectMeta      metav1.ObjectMeta
	ImageStreamSpec imagev1.ImageStreamSpec
}

// GetImageStream is a function to return the image stream
func GetImageStream(imageStreamParams ImageStreamParams) imagev1.ImageStream {
	imageStream := imagev1.ImageStream{
		TypeMeta:   imageStreamParams.TypeMeta,
		ObjectMeta: imageStreamParams.ObjectMeta,
		Spec:       imageStreamParams.ImageStreamSpec,
	}
	return imageStream
}

package lotaprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	lotaproviderv1alpha1 "github.com/mcanevet/lota-operator/pkg/apis/lotaprovider/v1alpha1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_lotaprovider")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

type providerSchemas struct {
	FormatVersion string                    `json:"format_version"`
	Schemas       map[string]providerSchema `json:"provider_schemas"`
}

type providerSchema struct {
	Provider          provider                  `json:"provider,omitempty"`
	ResourceSchemas   map[string]resourceSchema `json:"resource_schemas,omitempty"`
	DataSourceSchemas map[string]interface{}    `json:"data_source_schemas,omitempty"`
}

type provider struct {
	Block   block `json:"block"`
	Version int   `json:"version"`
}

type resourceSchema struct {
	Block   block `json:"block"`
	Version int   `json:"version"`
}

type block struct {
	Attributes map[string]attribute   `json:"attributes,omitempty"`
	BlockTypes map[string]interface{} `json:"block_types,omitempty"`
}

// LotaController hold the controller instance and method for a LotaProvider
type LotaController struct {
	Controller controller.Controller
	Client     dynamic.Interface
}

// from command/jsonprovider/attribute.go
type attribute struct {
	AttributeType json.RawMessage `json:"type,omitempty"`
	Description   string          `json:"description,omitempty"`
	Required      bool            `json:"required,omitempty"`
	Optional      bool            `json:"optional,omitempty"`
	Computed      bool            `json:"computed,omitempty"`
	Sensitive     bool            `json:"sensitive,omitempty"`
}

// Add creates a new LotaProvider Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	client, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return err
	}
	return add(mgr, newReconciler(mgr), client)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileLotaProvider{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// CRDWatcherMapper creates a EventHandler interface to map CustomResourceDefinition objects back to
// controller and add given GVK to watch list.
type CRDWatcherMapper struct {
	controller *LotaController
}

// Map requests directed to CRD objects and extract related GVK to trigger another watch on
// controller instance.
func (c *CRDWatcherMapper) Map(obj handler.MapObject) []reconcile.Request {
	// use c.controller to add a watch to the new CRD that obj refers to. Use dynamic client!
	mapperLogger := log.WithValues(
		"Object.Namespace", obj.Meta.GetNamespace(),
		"Object.Name", obj.Meta.GetName(),
	)
	mapperLogger.Info("Call Map")
	c.controller.Controller.Watch(&source.Kind{Type: obj.Object.GetObjectKind()}, NewCreateWatchEventHandler(c.controller))

	return []reconcile.Request{}
}

// NewCreateWatchEventHandler creates a new instance of handler.EventHandler interface with
// CRDWatcherMapper as map-func.
func NewCreateWatchEventHandler(controller *LotaController) handler.EventHandler {
	return &handler.EnqueueRequestsFromMapFunc{
		ToRequests: &CRDWatcherMapper{controller: controller},
	}
}

func (s *LotaController) addCRDWatch() error {
	err := s.Controller.Watch(&source.Kind{Type: &apiextensionsv1beta1.CustomResourceDefinition{}}, NewCreateWatchEventHandler(s))
	if err != nil {
		return err
	}

	return nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler, client dynamic.Interface) error {
	// Create a new controller
	c, err := controller.New("lotaprovider-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	myController := &LotaController{
		Controller: c,
		Client:     client,
	}

	// Watch for changes to primary resource LotaProvider
	err = myController.Controller.Watch(&source.Kind{Type: &lotaproviderv1alpha1.LotaProvider{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to GVKs relevant for LotaProvider
	err = myController.addCRDWatch()
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileLotaProvider implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileLotaProvider{}

// ReconcileLotaProvider reconciles a LotaProvider object
type ReconcileLotaProvider struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a LotaProvider object and makes changes based on the state read
// and what is in the LotaProvider.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileLotaProvider) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling LotaProvider")

	// Fetch the LotaProvider instance
	instance := &lotaproviderv1alpha1.LotaProvider{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	providerSchema := make(map[string]string)
	for i := 0; i < len(instance.Spec.Schema); i++ {
		providerSchema[instance.Spec.Schema[i].Name] = instance.Spec.Schema[i].Value
	}
	providerSchema["version"] = instance.Spec.Version

	provider := map[string]map[string]string{}
	provider[instance.Spec.Name] = providerSchema

	data := map[string]map[string]map[string]string{}
	data["provider"] = provider

	reqLogger.Info("DEBUG", "Terraform code to launch", data)

	dir, err := ioutil.TempDir("/tmp", "lota-operator")
	if err != nil {
		return reconcile.Result{}, err
	}
	defer os.RemoveAll(dir)

	tfCode, err := json.Marshal(data)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/providers.tf.json", dir), tfCode, 0644)
	if err != nil {
		return reconcile.Result{}, err
	}

	cmd := exec.Command("terraform", "init", "-no-color")
	cmd.Dir = dir
	err = cmd.Run()
	if err != nil {
		return reconcile.Result{}, err
	}

	cmd = exec.Command("terraform", "providers", "schema", "-json")
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return reconcile.Result{}, err
	}

	var ps providerSchemas
	if err = json.Unmarshal(stdout.Bytes(), &ps); err != nil {
		return reconcile.Result{}, err
	}

	var config *rest.Config
	kubeConfigPath := os.Getenv("KUBECONFIG")
	if kubeConfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	crdClient, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		return reconcile.Result{}, err
	}

	for k, v := range ps.Schemas[instance.Spec.Name].ResourceSchemas {
		// Define a new CRD object
		crd := newCRDForCR(instance, k, v.Block.Attributes)
		reqLogger.Info("DEBUG", "CRD", crd)
		reqLogger.Info("DEBUG", "CRD.Spec", crd.Spec)

		// Set LotaProvider instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, crd, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		reqLogger.Info("DEBUG", "Name", crd.Name, "Namespace", crd.Namespace)

		// Check if this CRD already exists
		found := &apiextensionsv1beta1.CustomResourceDefinition{}
		_, err = crdClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crd.Name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				reqLogger.Info("Creating a new CRD", "CRD.Namespace", crd.Namespace, "CRD.Name", crd.Name)
				_, err = crdClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
				if err != nil {
					return reconcile.Result{}, err
				}

				// CRD created successfully - don't requeue
			} else {
				return reconcile.Result{}, err
			}
		} else {
			// CRD already exists - don't requeue
			reqLogger.Info("Skip reconcile: CRD already exists", "CRD.Namespace", found.Namespace, "CRD.Name", found.Name)
		}
	}

	return reconcile.Result{}, nil
}

func snakeCaseToCamelCase(inputUnderScoreStr string) (camelCase string) {
	//snake_case to camelCase

	isToUpper := false

	for k, v := range inputUnderScoreStr {
		if k == 0 {
			camelCase = strings.ToUpper(string(inputUnderScoreStr[0]))
		} else {
			if isToUpper {
				camelCase += strings.ToUpper(string(v))
				isToUpper = false
			} else {
				if v == '_' {
					isToUpper = true
				} else {
					camelCase += string(v)
				}
			}
		}
	}
	return
}

// newCRDForCR returns a CustomResourceDefinition for the ResourceSchemas
func newCRDForCR(cr *lotaproviderv1alpha1.LotaProvider, resource string, attributes map[string]attribute) *apiextensionsv1beta1.CustomResourceDefinition {
	camelCased := snakeCaseToCamelCase(resource)
	singular := strings.ToLower(camelCased)
	plural := fmt.Sprintf("%ss", singular)

	properties := make(map[string]apiextensionsv1beta1.JSONSchemaProps)

	properties["provider"] = apiextensionsv1beta1.JSONSchemaProps{
		Type: "string",
	}

	for k, v := range attributes {
		// FIXME: only supports string for now
		if string(v.AttributeType) == "\"string\"" {
			properties[k] = apiextensionsv1beta1.JSONSchemaProps{
				Type: "string",
			}
		}
	}

	validationSchema := &apiextensionsv1beta1.JSONSchemaProps{
		Properties: map[string]apiextensionsv1beta1.JSONSchemaProps{
			"spec": apiextensionsv1beta1.JSONSchemaProps{
				Type:       "object",
				Properties: properties,
			},
		},
		Type: "object",
	}

	return &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s.%s.lota-operator.io", plural, cr.Spec.Name),
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   fmt.Sprintf("%s.lota-operator.io", cr.Spec.Name),
			Version: fmt.Sprintf("v%s-alpha1", strings.ReplaceAll(cr.Spec.Version, ".", "-")),
			Scope:   apiextensionsv1beta1.ClusterScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   plural,
				Singular: singular,
				Kind:     camelCased,
			},
			Validation: &apiextensionsv1beta1.CustomResourceValidation{
				OpenAPIV3Schema: validationSchema,
			},
		},
	}
}

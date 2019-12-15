package lotaresource

import (
	"strings"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_lotaresource")

// Controller hold the controller instance and method for a LotaResource
type Controller struct {
	Controller      controller.Controller
	Client          dynamic.Interface
	watchingObjects map[string]bool
}

// Add creates a new LotaResource Controller and adds it to the Manager. The Manager will set fields on the Controller
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
	return &ReconcileLotaResource{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler, client dynamic.Interface) error {
	// Create a new controller
	c, err := controller.New("lotaresource-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	s := &Controller{
		Controller:      c,
		Client:          client,
		watchingObjects: make(map[string]bool),
	}

	// Watch for changes to primary resource CustomResourceDefinition
	err = s.Controller.Watch(
		&source.Kind{Type: &apiextensionsv1beta1.CustomResourceDefinition{}},
		NewCreateWatchEventHandler(s),
	)
	if err != nil {
		return err
	}

	return nil
}

// CRDWatcherMapper creates a EventHandler interface to map CustomResourceDefinition objects back to
// controller and add given GVK to watch list.
type CRDWatcherMapper struct {
	controller *Controller
}

// Map requests directed to CRD objects and extract related GVK to trigger another watch on
// controller instance.
func (c *CRDWatcherMapper) Map(obj handler.MapObject) []reconcile.Request {
	// use c.controller to add a watch to the new CRD that obj refers to. Use dynamic client!
	mapperLogger := log.WithValues(
		"Object.Namespace", obj.Meta.GetNamespace(),
		"Object.Name", obj.Meta.GetName(),
	)
	mapperLogger.Info("Mapping LotaResource")

	if strings.HasSuffix(obj.Meta.GetName(), ".lota-operator.io") && !strings.HasSuffix(obj.Meta.GetName(), "lotaprovider.lota-operator.io") {
		if _, exists := c.controller.watchingObjects[obj.Meta.GetName()]; exists {
			mapperLogger.Info("Skip Watching: Already under watch")
		} else {
			c.controller.Controller.Watch(&source.Kind{Type: obj.Object}, NewCreateWatchEventHandler(c.controller))
			c.controller.watchingObjects[obj.Meta.GetName()] = true
		}
	} else {
		mapperLogger.Info("Skip Watching: Not a lota resource")
	}

	return []reconcile.Request{}
}

// NewCreateWatchEventHandler creates a new instance of handler.EventHandler interface with
// CRDWatcherMapper as map-func.
func NewCreateWatchEventHandler(controller *Controller) handler.EventHandler {
	return &handler.EnqueueRequestsFromMapFunc{
		ToRequests: &CRDWatcherMapper{controller: controller},
	}
}

// blank assignment to verify that ReconcileLotaResource implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileLotaResource{}

// ReconcileLotaResource reconciles a LotaResource object
type ReconcileLotaResource struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a LotaResource object and makes changes based on the state read
// and what is in the LotaResource.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileLotaResource) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling LotaResource")

	return reconcile.Result{}, nil
}

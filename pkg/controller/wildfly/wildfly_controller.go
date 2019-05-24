package wildfly

import (
	"context"
	"log"

	wildflyv1alpha1 "github.com/giannisalinetti/wildfly-operator/pkg/apis/wildfly/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Define constant and defaults for the deployment
const (
	containerNameString = "wildfly"
	imageDefault        = "docker.io/jboss/wildfly"
)

// Slices and maps cannot be initialized as constants in Go
var (
	commandDefault = []string{"/opt/jboss/wildfly/bin/standalone.sh", "-b", "0.0.0.0"}
)

// Add creates a new Wildfly Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileWildfly{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("wildfly-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Wildfly
	err = c.Watch(&source.Kind{Type: &wildflyv1alpha1.Wildfly{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Wildfly
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &wildflyv1alpha1.Wildfly{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileWildfly{}

// ReconcileWildfly reconciles a Wildfly object
type ReconcileWildfly struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Wildfly object and makes changes based on the state read
// and what is in the Wildfly.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileWildfly) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling Wildfly %s/%s\n", request.Namespace, request.Name)

	// Fetch the Wildfly instance
	instance := &wildflyv1alpha1.Wildfly{}
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

	// Deployment reconciliation
	foundDep := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, foundDep)
	if err != nil && errors.IsNotFound(err) {
		// Define new Wildfly Deployment
		dep := r.newWildflyDeployment(instance)
		log.Printf("Creating a new Wildfly Deployment: %s/%s\n", dep.Namespace, dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			log.Printf("Failed to create new Wildfly Deployment: %v\n", err)
			return reconcile.Result{}, err
		}
		// After successful deployment return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		log.Printf("Failed to get Deployment: %v\n", err)
		return reconcile.Result{}, err
	}

	// Reconcile deployment size
	size := instance.Spec.Size
	if *foundDep.Spec.Replicas != size {
		foundDep.Spec.Replicas = &size
		err = r.client.Update(context.TODO(), foundDep)
		if err != nil {
			log.Printf("Failed to update Wildfly Deployment: %v\n", err)
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Service reconciliation
	foundSvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, foundSvc)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Wildfly Service
		svc := r.newWildflyService(instance)
		log.Printf("Creating a new Wildfly Service: %s/%s\n", svc.Namespace, svc.Name)
		err = r.client.Create(context.TODO(), svc)
		if err != nil {
			log.Printf("Failed to create new Wildfly Service: %v\n", err)
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		log.Printf("Failed to get Service: %v\n", err)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// newWildflyDeployment manages the creation of a wildfly Deployment
func (r *ReconcileWildfly) newWildflyDeployment(cr *wildflyv1alpha1.Wildfly) *appsv1.Deployment {
	// cr variables declaration
	var replicas int32
	var imageString string
	var imageTag string
	var commandSlice []string

	labels := map[string]string{
		"app": cr.Name,
	}

	// Don' accept negative replicas
	if cr.Spec.Size < 0 {
		replicas = 1
	} else {
		replicas = cr.Spec.Size
	}

	// If no image name is assigned we default to docker.io/jboss/wildfly
	if cr.Spec.Image == "" {
		imageString = imageDefault
	} else {
		imageString = cr.Spec.Image
	}

	// Use latest tag if version is an empty string
	if cr.Spec.Version == "" {
		imageTag = "latest"
	} else {
		imageTag = cr.Spec.Version
	}

	// Pass a default command slice if nothing is provided
	if cr.Spec.Cmd == nil {
		commandSlice = commandDefault
	} else {
		commandSlice = cr.Spec.Cmd
	}

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:    containerNameString,
						Image:   imageString + ":" + imageTag,
						Command: commandSlice,
					}},
				},
			},
		},
	}
	controllerutil.SetControllerReference(cr, dep, r.scheme)
	return dep
}

func (r *ReconcileWildfly) newWildflyService(cr *wildflyv1alpha1.Wildfly) *corev1.Service {
	labels := map[string]string{
		"app": cr.Name,
	}
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:     "default",
					Port:     8080,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}

	controllerutil.SetControllerReference(cr, svc, r.scheme)
	return svc
}

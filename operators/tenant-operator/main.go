// Tenant Operator
// This operator manages Tenant custom resources and creates the necessary
// namespace, ResourceQuota, LimitRange, NetworkPolicies, and RBAC

package main

import (
	"context"
	"flag"
	"os"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	// Add custom API types to scheme
	// utilruntime.Must(platformv1alpha1.AddToScheme(scheme))
}

// TenantSpec defines the desired state of Tenant
type TenantSpec struct {
	Owner               string              `json:"owner"`
	CostCenter          string              `json:"costCenter,omitempty"`
	Quota               TenantQuota         `json:"quota,omitempty"`
	AllowedIntegrations []string            `json:"allowedIntegrations,omitempty"`
	Contacts            map[string]string   `json:"contacts,omitempty"`
}

type TenantQuota struct {
	CPU      string `json:"cpu,omitempty"`
	Memory   string `json:"memory,omitempty"`
	Pods     int    `json:"pods,omitempty"`
	PVCs     int    `json:"pvcs,omitempty"`
	Services int    `json:"services,omitempty"`
}

// TenantStatus defines the observed state of Tenant
type TenantStatus struct {
	Phase                  string `json:"phase,omitempty"`
	NamespaceCreated       bool   `json:"namespaceCreated,omitempty"`
	QuotaApplied           bool   `json:"quotaApplied,omitempty"`
	NetworkPolicyApplied   bool   `json:"networkPolicyApplied,omitempty"`
	RBACApplied            bool   `json:"rbacApplied,omitempty"`
}

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile handles the reconciliation loop for Tenant resources
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Reconciling Tenant", "name", req.Name)

	// This is a simplified example - in production, you would:
	// 1. Fetch the Tenant CR
	// 2. Create namespace if not exists
	// 3. Apply ResourceQuota
	// 4. Apply LimitRange
	// 5. Apply NetworkPolicies
	// 6. Apply RBAC
	// 7. Update status

	// For now, we'll create resources based on the tenant name
	tenantName := req.Name

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: tenantName,
			Labels: map[string]string{
				"platform.xyz.com/tenant":                     tenantName,
				"istio-injection":                             "enabled",
				"pod-security.kubernetes.io/enforce":          "restricted",
			},
		},
	}

	if err := r.Create(ctx, ns); err != nil {
		if !errors.IsAlreadyExists(err) {
			log.Error(err, "Failed to create namespace")
			return ctrl.Result{}, err
		}
	}
	log.Info("Namespace created/exists", "namespace", tenantName)

	// Create ResourceQuota
	quota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-quota",
			Namespace: tenantName,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourceRequestsCPU:    resource.MustParse("10"),
				corev1.ResourceRequestsMemory: resource.MustParse("20Gi"),
				corev1.ResourceLimitsCPU:      resource.MustParse("20"),
				corev1.ResourceLimitsMemory:   resource.MustParse("40Gi"),
				corev1.ResourcePods:           resource.MustParse("100"),
			},
		},
	}

	if err := r.Create(ctx, quota); err != nil {
		if !errors.IsAlreadyExists(err) {
			log.Error(err, "Failed to create ResourceQuota")
			return ctrl.Result{}, err
		}
	}
	log.Info("ResourceQuota created/exists", "namespace", tenantName)

	// Create default deny NetworkPolicy
	netpol := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-deny-ingress",
			Namespace: tenantName,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
			},
		},
	}

	if err := r.Create(ctx, netpol); err != nil {
		if !errors.IsAlreadyExists(err) {
			log.Error(err, "Failed to create NetworkPolicy")
			return ctrl.Result{}, err
		}
	}
	log.Info("NetworkPolicy created/exists", "namespace", tenantName)

	// Create RoleBinding for tenant team
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenantName + "-developers",
			Namespace: tenantName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "Group",
				Name:     tenantName + "-team",
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "edit",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	if err := r.Create(ctx, roleBinding); err != nil {
		if !errors.IsAlreadyExists(err) {
			log.Error(err, "Failed to create RoleBinding")
			return ctrl.Result{}, err
		}
	}
	log.Info("RoleBinding created/exists", "namespace", tenantName)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// For(&platformv1alpha1.Tenant{}).  // Uncomment when CRD is registered
		For(&corev1.Namespace{}). // Temporary: watch namespaces instead
		Complete(r)
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "tenant-operator.platform.xyz.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&TenantReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Tenant")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}


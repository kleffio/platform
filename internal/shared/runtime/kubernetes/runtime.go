// Package kubernetes implements RuntimeAdapter using the Kubernetes API.
// Requires the platform Pod to have a ServiceAccount with the RBAC permissions
// documented in PLUGIN_RUNTIME.md.
//
// This is a structural stub. To enable it:
//  1. Add k8s.io/client-go and k8s.io/api to go.mod.
//  2. Uncomment the import block and implement the methods below.
package kubernetes

import (
	"context"
	"fmt"

	"github.com/kleffio/platform/internal/shared/runtime"
)

// Runtime implements runtime.RuntimeAdapter for Kubernetes.
type Runtime struct {
	namespace string
	// client    *kubernetes.Clientset  // uncomment when k8s.io/client-go is added
}

// New creates a KubernetesRuntime.
// namespace is the k8s namespace plugins are deployed into (default: "kleff").
func New(namespace string) (*Runtime, error) {
	if namespace == "" {
		namespace = "kleff"
	}
	// cfg, err := rest.InClusterConfig()
	// if err != nil { return nil, fmt.Errorf("kubernetes: in-cluster config: %w", err) }
	// cli, err := kubernetes.NewForConfig(cfg)
	// if err != nil { return nil, fmt.Errorf("kubernetes: create client: %w", err) }
	return &Runtime{namespace: namespace}, nil
}

var _ runtime.RuntimeAdapter = (*Runtime)(nil)

func (r *Runtime) Deploy(_ context.Context, spec runtime.ContainerSpec) error {
	return fmt.Errorf("kubernetes runtime not yet fully implemented — add k8s.io/client-go and implement")
}

func (r *Runtime) Remove(_ context.Context, id string) error {
	return fmt.Errorf("kubernetes runtime not yet fully implemented")
}

func (r *Runtime) Start(_ context.Context, id string) error {
	return fmt.Errorf("kubernetes runtime not yet fully implemented")
}

func (r *Runtime) Stop(_ context.Context, id string) error {
	return fmt.Errorf("kubernetes runtime not yet fully implemented")
}

func (r *Runtime) Status(_ context.Context, id string) (runtime.ContainerStatus, error) {
	return runtime.ContainerStatus{ID: id, State: runtime.StateUnknown}, nil
}

// Endpoint returns the ClusterIP Service DNS name for a plugin.
// Pattern: {id}.{namespace}.svc.cluster.local:{port}
func (r *Runtime) Endpoint(_ context.Context, id string, port int) (string, error) {
	return fmt.Sprintf("%s.%s.svc.cluster.local:%d", id, r.namespace, port), nil
}

func (r *Runtime) Logs(_ context.Context, _ string, _ int) ([]string, error) {
	return nil, fmt.Errorf("kubernetes runtime not yet fully implemented")
}

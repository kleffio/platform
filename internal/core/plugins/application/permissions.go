package application

import pluginsv1 "github.com/kleffio/plugin-sdk-go/v1"

// Layer tag constants mirror pluginsv1.Tag* — defined locally so they compile
// against the published SDK v0.1.1 which does not yet export them.
const (
	tagFrontend = "frontend"
	tagBackend  = "backend"
	tagIdentity = "identity"
	tagDevOps   = "devops"
)

// tagCapabilities maps each layer tag to the set of capabilities it permits.
// A plugin may only exercise capabilities that its declared tags allow.
var tagCapabilities = map[string][]string{
	tagFrontend: {pluginsv1.CapabilityUIManifest},
	tagBackend:  {pluginsv1.CapabilityAPIMiddleware, pluginsv1.CapabilityAPIRoutes},
	tagIdentity: {pluginsv1.CapabilityIdentityProvider},
	tagDevOps:   {}, // reserved for future daemon/k8s capabilities
}

// permittedCapabilities returns the set of capabilities a plugin is allowed to
// declare based on its manifest tags. Returns nil if no layer tag is present,
// meaning no restrictions apply (backwards-compatible behaviour).
func permittedCapabilities(tags []string) map[string]bool {
	permitted := make(map[string]bool)
	hasLayerTag := false
	for _, tag := range tags {
		caps, ok := tagCapabilities[tag]
		if !ok {
			continue
		}
		hasLayerTag = true
		for _, c := range caps {
			permitted[c] = true
		}
	}
	if !hasLayerTag {
		return nil
	}
	return permitted
}

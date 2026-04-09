package application

import (
	"testing"

	pluginsv1 "github.com/kleffio/plugin-sdk-go/v1"
)

func TestPermittedCapabilities(t *testing.T) {
	tests := []struct {
		name    string
		tags    []string
		wantNil bool
		allowed []string
		denied  []string
	}{
		{
			name:    "no tags — backwards compat, no restrictions",
			tags:    []string{},
			wantNil: true,
		},
		{
			name:    "unrecognised tags only — no restrictions",
			tags:    []string{"storage", "networking"},
			wantNil: true,
		},
		{
			name:    "frontend tag only",
			tags:    []string{"frontend"},
			wantNil: false,
			allowed: []string{pluginsv1.CapabilityUIManifest},
			denied:  []string{pluginsv1.CapabilityAPIMiddleware, pluginsv1.CapabilityAPIRoutes, pluginsv1.CapabilityIdentityProvider},
		},
		{
			name:    "backend tag only",
			tags:    []string{"backend"},
			wantNil: false,
			allowed: []string{pluginsv1.CapabilityAPIMiddleware, pluginsv1.CapabilityAPIRoutes},
			denied:  []string{pluginsv1.CapabilityUIManifest, pluginsv1.CapabilityIdentityProvider},
		},
		{
			name:    "identity tag only",
			tags:    []string{"identity"},
			wantNil: false,
			allowed: []string{pluginsv1.CapabilityIdentityProvider},
			denied:  []string{pluginsv1.CapabilityUIManifest, pluginsv1.CapabilityAPIMiddleware, pluginsv1.CapabilityAPIRoutes},
		},
		{
			name:    "devops tag only — no capabilities yet",
			tags:    []string{"devops"},
			wantNil: false,
			denied:  []string{pluginsv1.CapabilityUIManifest, pluginsv1.CapabilityAPIMiddleware, pluginsv1.CapabilityAPIRoutes, pluginsv1.CapabilityIdentityProvider},
		},
		{
			name:    "frontend + backend",
			tags:    []string{"frontend", "backend"},
			wantNil: false,
			allowed: []string{pluginsv1.CapabilityUIManifest, pluginsv1.CapabilityAPIMiddleware, pluginsv1.CapabilityAPIRoutes},
			denied:  []string{pluginsv1.CapabilityIdentityProvider},
		},
		{
			name:    "backend + identity (IDP with custom routes)",
			tags:    []string{"backend", "identity"},
			wantNil: false,
			allowed: []string{pluginsv1.CapabilityAPIMiddleware, pluginsv1.CapabilityAPIRoutes, pluginsv1.CapabilityIdentityProvider},
			denied:  []string{pluginsv1.CapabilityUIManifest},
		},
		{
			name:    "frontend + backend + identity (full-stack plugin)",
			tags:    []string{"frontend", "backend", "identity"},
			wantNil: false,
			allowed: []string{pluginsv1.CapabilityUIManifest, pluginsv1.CapabilityAPIMiddleware, pluginsv1.CapabilityAPIRoutes, pluginsv1.CapabilityIdentityProvider},
		},
		{
			name:    "layer tag mixed with unrecognised tags",
			tags:    []string{"frontend", "oss", "cool-plugin"},
			wantNil: false,
			allowed: []string{pluginsv1.CapabilityUIManifest},
			denied:  []string{pluginsv1.CapabilityAPIMiddleware, pluginsv1.CapabilityAPIRoutes, pluginsv1.CapabilityIdentityProvider},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := permittedCapabilities(tc.tags)

			if tc.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("expected non-nil map, got nil")
			}

			for _, cap := range tc.allowed {
				if !got[cap] {
					t.Errorf("expected capability %q to be allowed, but it was not", cap)
				}
			}

			for _, cap := range tc.denied {
				if got[cap] {
					t.Errorf("expected capability %q to be denied, but it was allowed", cap)
				}
			}
		})
	}
}

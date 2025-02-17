package e2e

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	kudov1alpha1 "github.com/jlevesy/kudo/pkg/apis/k8s.kudo.dev/v1alpha1"
)

type escalationWaitCondSpec struct {
	state         kudov1alpha1.EscalationState
	grantStatuses []kudov1alpha1.GrantStatus
}

func condEscalationStatusMatchesSpec(spec escalationWaitCondSpec) updateCondition {
	return func(obj runtime.Object) bool {
		esc, ok := obj.(*kudov1alpha1.Escalation)
		if !ok {
			return false
		}

		return escalationStatusMatchesSpec(spec, esc.Status)
	}
}

func escalationStatusMatchesSpec(want escalationWaitCondSpec, got kudov1alpha1.EscalationStatus) bool {
	if want.state != got.State {
		return false
	}

	if len(want.grantStatuses) != len(got.GrantRefs) {
		return false
	}

	for i, refStatus := range want.grantStatuses {
		if refStatus != got.GrantRefs[i].Status {
			return false
		}
	}

	return true
}

func assertGrantedK8sResourcesCreated(t *testing.T, esc kudov1alpha1.Escalation, resource string) {
	for _, ref := range esc.Status.GrantRefs {
		assertObjectCreated(
			t,
			admin.k8s.RbacV1().RESTClient(),
			resourceNameNamespace{
				resource:  resource,
				name:      ref.Name,
				namespace: ref.Namespace,
			},
			30*time.Second,
		)
	}
}

func assertGrantedK8sResourcesDeleted(t *testing.T, esc kudov1alpha1.Escalation, resource string) {
	for _, ref := range esc.Status.GrantRefs {
		assertObjectDeleted(
			t,
			admin.k8s.RbacV1().RESTClient(),
			resourceNameNamespace{
				resource:  resource,
				name:      ref.Name,
				namespace: ref.Namespace,
			},
			30*time.Second,
		)
	}
}

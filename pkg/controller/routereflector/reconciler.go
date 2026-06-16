// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package routereflector holds the route reflector lifecycle controller for the calico
// networking extension.
//
// SCAFFOLD ONLY. This file establishes the types and the reconciliation entrypoint, but
// the reconciliation logic is intentionally not implemented. Before this controller can be
// wired into the extension manager, the following design points need to be settled:
//
//   - Where does the controller run? Today the extension actuator runs against the seed
//     and applies the calico chart to the shoot via the gardener resource manager. The
//     route reflector reconciler needs a client for the SHOOT cluster, since it reads/
//     writes Node and Calico Node/BGPPeer CRs there. This requires either:
//       a) Embedding the controller in calico-kube-controllers (deployed into the shoot),
//          which keeps it close to Calico but bloats that image.
//       b) Running it in a managed sidecar/pod the extension deploys to the shoot,
//          analogous to how cluster-autoscaler's worker side runs.
//       c) Running it from the seed via the shoot's kubeconfig (simplest, least HA).
//     Option (c) is the cheapest path and matches the rest of the extension; (b) is the
//     long-term answer.
//
//   - Migration sequencing. Flipping nodeToNodeMeshEnabled from true to false BEFORE the
//     RR peerings are established and converged will black-hole pod traffic. The
//     controller MUST:
//       1. Label countPerZone nodes per zone with route-reflector=true.
//       2. Set spec.bgp.routeReflectorClusterID on the corresponding Calico Node CR.
//       3. Wait for client->RR and RR<->RR sessions to reach Established state
//          (observable via calico-node BIRD status or the BGPPeer status).
//       4. Only then patch BGPConfiguration to nodeToNodeMeshEnabled=false.
//     Step 3 is the hard part and is not implemented in this scaffold.
//
//   - RR replacement on node deletion. When an RR node is removed (autoscaler scale-down,
//     rolling update), the controller must promote a replacement node BEFORE the old RR
//     is gone, again to avoid a window without working route reflectors.
//
//   - Failure-domain awareness. countPerZone is the user-facing knob, but the controller
//     should also avoid colocating all RRs of a zone on a single instance type or
//     spot-priced pool, where applicable.
//
// None of the above is implemented. This package compiles, exposes a Reconciler type,
// and is intended to be filled in iteratively. It is NOT registered with the manager.
package routereflector

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apiscalico "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico"
)

// Config controls how the route reflector controller picks and manages RRs.
type Config struct {
	// CountPerZone is the desired number of route reflectors per failure domain (zone).
	// Recommended: 2 for HA in production. 1 is acceptable for test clusters.
	CountPerZone int
	// LabelKey is the node label set on selected RRs (default: "route-reflector").
	LabelKey string
	// LabelValue is the value of the RR label (default: "true").
	LabelValue string
	// ClusterID is the BGP route reflector cluster ID set on every selected RR.
	// Defaults to "224.0.0.1".
	ClusterID string
}

// Reconciler reconciles the desired route reflector population on the shoot cluster.
//
// This is a SCAFFOLD: the Reconcile method is unimplemented and returns an error so that
// accidentally registering this controller is a loud failure rather than silent
// black-hole-causing dataplane changes.
type Reconciler struct {
	// Client is a client for the SHOOT cluster (NOT the seed). The reconciler reads
	// core/v1 Nodes and writes:
	//   - core/v1 Node labels on selected RR nodes.
	//   - crd.projectcalico.org/v1 Node CRs (sets spec.bgp.routeReflectorClusterID).
	//   - crd.projectcalico.org/v1 BGPConfiguration (sets nodeToNodeMeshEnabled=false
	//     once peers are converged).
	Client client.Client

	Config Config

	// BGP is the providerConfig BGP block. Used to derive the desired NodeSelector and to
	// enforce the migration sequencing precondition (do not disable mesh until peers are up).
	BGP *apiscalico.BGP
}

// Reconcile is intentionally unimplemented in this scaffold. See the package doc comment
// for the design points that need to be resolved before this can be filled in.
func (r *Reconciler) Reconcile(_ context.Context, _ reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, fmt.Errorf("route reflector lifecycle controller is not implemented yet; see pkg/controller/routereflector package doc for the open design points")
}

// SelectRouteReflectors is the placeholder for the per-zone selection algorithm.
// It will pick Config.CountPerZone nodes per zone with stable identity (so the same nodes
// remain RRs across reconciles when possible), preferring nodes that are not currently
// candidates for scale-down.
//
// Unimplemented.
func (r *Reconciler) SelectRouteReflectors(_ context.Context) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

// EnsurePeersConverged is the placeholder for waiting until BGP sessions to/between RRs
// are Established. Required precondition before disabling the full mesh.
//
// Unimplemented.
func (r *Reconciler) EnsurePeersConverged(_ context.Context) error {
	return fmt.Errorf("not implemented")
}

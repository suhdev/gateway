// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/test/benchmark/suite"
)

func init() {
	BenchmarkTests = append(BenchmarkTests, ScaleHTTPRoutes30K)
}

var ScaleHTTPRoutes30K = suite.BenchmarkTest{
	ShortName:   "ScaleHTTPRoute30K",
	Description: "Test gateway scalability with over 30K HTTPRoutes to verify it can handle large-scale deployments.",
	Test: func(t *testing.T, bSuite *suite.BenchmarkTestSuite) (reports []*suite.BenchmarkReport) {
		var (
			ctx               = context.Background()
			ns                = "benchmark-test"
			totalHosts uint16 = 10
			err        error
		)

		gatewayNN := types.NamespacedName{Name: "benchmark-30k", Namespace: ns}
		gateway := bSuite.GatewayTemplate.DeepCopy()
		gateway.SetName(gatewayNN.Name)
		err = bSuite.CreateResource(ctx, gateway)
		require.NoError(t, err)

		routeNameFormat := "benchmark-30k-route-%d"
		routeHostnameFormat := "www.benchmark-30k-%d.com"
		// Test with 30K+ routes - scaling up to 30,000 routes
		routeScales := []uint16{30000}
		routeScalesN := len(routeScales)
		routeNNs := make([]types.NamespacedName, 0, routeScales[routeScalesN-1])

		bSuite.RegisterCleanup(t, ctx, gateway, &gwapiv1.HTTPRoute{})

		t.Run("scaling up to 30K httproutes", func(t *testing.T) {
			var start, batch uint16 = 0, 0
			for _, scale := range routeScales {
				routePerHost := scale / totalHosts
				testName := fmt.Sprintf("scaling up httproutes to %d with %d routes per hostname", scale, routePerHost)

				t.Run(testName, func(t *testing.T) {
					tlog.Logf(t, "Start scaling up HTTPRoutes to %d with %d routes per hostname", scale, routePerHost)
					startTime := time.Now()

					// For this test, we only scale once (30000 routes), so batch starts at 0
					// routePerHost-batch = 3000-0 = 3000 routes per host
					err = bSuite.ScaleUpHTTPRoutes(ctx, [2]uint16{start, scale}, routeNameFormat, routeHostnameFormat, gatewayNN.Name, routePerHost-batch,
						func(route *gwapiv1.HTTPRoute, applyAt time.Time) {
							routeNN := types.NamespacedName{Name: route.Name, Namespace: route.Namespace}
							routeNNs = append(routeNNs, routeNN)

							// Log progress every 1000 routes
							if len(routeNNs)%1000 == 0 {
								t.Logf("Created %d HTTPRoutes so far", len(routeNNs))
							}
						})
					require.NoError(t, err)
					start = scale
					batch = routePerHost

					duration := time.Since(startTime)
					t.Logf("Successfully created %d HTTPRoutes in %s", scale, duration)

					// Verify that the gateway accepts the routes
					t.Logf("Verifying gateway accepts all routes...")
					gatewayAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, bSuite.Client, bSuite.TimeoutConfig,
						bSuite.ControllerName, kubernetes.NewGatewayRef(gatewayNN), routeNNs...)

					t.Logf("Gateway address: %s, all %d routes accepted", gatewayAddr, len(routeNNs))

					// Run benchmark test at this scale
					jobName := fmt.Sprintf("scale-up-httproutes-%d", scale)
					report, err := bSuite.Benchmark(t, ctx, jobName, testName, gatewayAddr, routeHostnameFormat, int(totalHosts), startTime)
					require.NoError(t, err)

					reports = append(reports, report)
				})
			}
		})

		t.Run("scaling down from 30K httproutes", func(t *testing.T) {
			// Scale down from 30000 to 1 using the suite method
			// The remaining routes will be cleaned up by RegisterCleanup
			start := routeScales[routeScalesN-1]
			scale := uint16(1)
			testName := fmt.Sprintf("scaling down httproutes from %d to %d", start, scale)

			t.Run(testName, func(t *testing.T) {
				startTime := time.Now()

				// Use the suite's ScaleDownHTTPRoutes method for consistency
				// Track deletion count for logging
				deletedCount := 0
				err = bSuite.ScaleDownHTTPRoutes(ctx, [2]uint16{start, scale}, routeNameFormat, gatewayNN.Name,
					func(route *gwapiv1.HTTPRoute) {
						deletedCount++
						// Log progress every 1000 routes
						if deletedCount%1000 == 0 {
							t.Logf("Deleted %d HTTPRoutes so far", deletedCount)
						}
					})
				require.NoError(t, err)

				duration := time.Since(startTime)
				t.Logf("Successfully deleted %d HTTPRoutes in %s (1 route remaining for cleanup)", deletedCount, duration)
			})
		})

		return
	},
}

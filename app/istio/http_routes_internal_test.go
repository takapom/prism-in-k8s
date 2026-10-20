package istio

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gold-kou/prism-in-k8s/app/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
)

func TestBuildHTTPRoutesDelaySerialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		delayNanos  int32
		wantSeconds int64
		wantNanos   int32
	}{
		{
			name:        "100ms",
			delayNanos:  100_000_000,
			wantSeconds: 0,
			wantNanos:   100_000_000,
		},
		{
			name:        "just below one second",
			delayNanos:  999_999_999,
			wantSeconds: 0,
			wantNanos:   999_999_999,
		},
		{
			name:        "one second",
			delayNanos:  1_000_000_000,
			wantSeconds: 1,
			wantNanos:   0,
		},
		{
			name:        "one and a half seconds",
			delayNanos:  1_500_000_000,
			wantSeconds: 1,
			wantNanos:   500_000_000,
		},
		{
			name:        "two seconds",
			delayNanos:  2_000_000_000,
			wantSeconds: 2,
			wantNanos:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			const host = "sample-prism-mock.sample-prism-mock.svc.cluster.local"
			routes := buildHTTPRoutes(host, []params.VirtualServiceRoute{
				{Name: "delayed", DelayNanos: tt.delayNanos, DelayPercentage: 100},
			})
			require.NotEmpty(t, routes)
			delay := routes[0].GetFault().GetDelay().GetFixedDelay()
			require.NotNil(t, delay)

			// Exercise the same protobuf JSON marshaler used for the API request.
			spec := &networkingv1alpha3.VirtualService{
				Hosts: []string{host},
				Http:  routes,
			}
			_, err := json.Marshal(spec)
			require.NoError(t, err)

			assert.Equal(t, tt.wantSeconds, delay.GetSeconds())
			assert.Equal(t, tt.wantNanos, delay.GetNanos())
			assert.Equal(t, time.Duration(tt.delayNanos), delay.AsDuration())
		})
	}
}

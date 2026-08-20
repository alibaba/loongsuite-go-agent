// Copyright (c) 2026 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sdkconfig

import (
	"strings"
	"testing"
)

func TestNewSpanSampler(t *testing.T) {
	const parentBasedAlwaysOn = "ParentBased{root:AlwaysOnSampler"

	tests := []struct {
		name     string
		legacy   string
		standard string
		arg      string
		want     string
	}{
		// Nothing configured keeps the pre-existing default.
		{"unset", "", "", "", parentBasedAlwaysOn},

		// OTEL_TRACE_SAMPLER, unchanged behaviour.
		{"legacy never", "0", "", "", "AlwaysOffSampler"},
		{"legacy negative", "-1", "", "", "AlwaysOffSampler"},
		{"legacy always", "1", "", "", "AlwaysOnSampler"},
		{"legacy ratio", "0.1", "", "", "ParentBased{root:TraceIDRatioBased{0.1}"},
		{"legacy invalid", "bogus", "", "", parentBasedAlwaysOn},

		// OTEL_TRACE_SAMPLER wins when both are set.
		{"legacy over standard", "0.1", "always_off", "", "ParentBased{root:TraceIDRatioBased{0.1}"},

		// Standard values.
		{"always_on", "", "always_on", "", "AlwaysOnSampler"},
		{"always_off", "", "always_off", "", "AlwaysOffSampler"},
		{"traceidratio", "", "traceidratio", "0.25", "TraceIDRatioBased{0.25}"},
		{"parentbased_always_on", "", "parentbased_always_on", "", parentBasedAlwaysOn},
		{"parentbased_always_off", "", "parentbased_always_off", "", "ParentBased{root:AlwaysOffSampler"},
		{"parentbased_traceidratio", "", "parentbased_traceidratio", "0.1", "ParentBased{root:TraceIDRatioBased{0.1}"},

		// Names are matched case-insensitively and trimmed.
		{"uppercase", "", "PARENTBASED_TRACEIDRATIO", "0.1", "ParentBased{root:TraceIDRatioBased{0.1}"},
		{"padded", "", "  always_off  ", "", "AlwaysOffSampler"},

		// Samplers that are not linked into the agent fall back.
		{"jaeger_remote", "", "jaeger_remote", "", parentBasedAlwaysOn},
		{"parentbased_jaeger_remote", "", "parentbased_jaeger_remote", "", parentBasedAlwaysOn},
		{"xray", "", "xray", "", parentBasedAlwaysOn},
		{"unknown", "", "not_a_sampler", "", parentBasedAlwaysOn},

		// An unusable ratio falls back to 1.0, which is AlwaysOn.
		{"missing arg", "", "traceidratio", "", "AlwaysOnSampler"},
		{"invalid arg", "", "traceidratio", "abc", "AlwaysOnSampler"},
		{"arg above range", "", "traceidratio", "5", "AlwaysOnSampler"},
		{"arg below range", "", "traceidratio", "-1", "AlwaysOnSampler"},
		{"arg NaN", "", "traceidratio", "NaN", "AlwaysOnSampler"},
		{"arg Inf", "", "traceidratio", "Inf", "AlwaysOnSampler"},

		// The same fallbacks reach the parent-based variant.
		{"parentbased missing arg", "", "parentbased_traceidratio", "", parentBasedAlwaysOn},
		{"parentbased invalid arg", "", "parentbased_traceidratio", "abc", parentBasedAlwaysOn},
		{"parentbased arg above range", "", "parentbased_traceidratio", "5", parentBasedAlwaysOn},
		{"parentbased arg below range", "", "parentbased_traceidratio", "-1", parentBasedAlwaysOn},
		{"parentbased arg NaN", "", "parentbased_traceidratio", "NaN", parentBasedAlwaysOn},
		{"parentbased arg Inf", "", "parentbased_traceidratio", "Inf", parentBasedAlwaysOn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(trace_sampler, tt.legacy)
			t.Setenv(traces_sampler, tt.standard)
			t.Setenv(traces_sampler_arg, tt.arg)

			got := NewSpanSampler().Description()
			if !strings.HasPrefix(got, tt.want) {
				t.Errorf("NewSpanSampler() = %s, want it to start with %s", got, tt.want)
			}
		})
	}
}

func TestParseSamplerRatio(t *testing.T) {
	tests := []struct {
		arg  string
		want float64
	}{
		{"0.1", 0.1},
		{"0", 0},
		{"1", 1},
		{"  0.5  ", 0.5},
		{"", default_sampler_ratio},
		{"abc", default_sampler_ratio},
		{"1.5", default_sampler_ratio},
		{"-0.1", default_sampler_ratio},
		// Every comparison against NaN is false, so a range check alone would
		// let it reach the sampler.
		{"NaN", default_sampler_ratio},
		{"Inf", default_sampler_ratio},
		{"-Inf", default_sampler_ratio},
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			if got := parseSamplerRatio(tt.arg); got != tt.want {
				t.Errorf("parseSamplerRatio(%q) = %v, want %v", tt.arg, got, tt.want)
			}
		})
	}
}

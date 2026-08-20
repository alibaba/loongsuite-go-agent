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

// Package sdkconfig turns the OpenTelemetry SDK environment variables into
// SDK objects. It deliberately holds no references to the symbols the agent
// injects at build time, so it can be exercised by ordinary unit tests.
package sdkconfig

import (
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/sdk/trace"
)

// Agent-specific sampler variable, kept for deployments that already use it.
const trace_sampler = "OTEL_TRACE_SAMPLER"

// Standard OpenTelemetry SDK sampler configuration.
// https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/
const traces_sampler = "OTEL_TRACES_SAMPLER"
const traces_sampler_arg = "OTEL_TRACES_SAMPLER_ARG"

// Ratio used when a *_traceidratio sampler is selected without a usable
// OTEL_TRACES_SAMPLER_ARG, as required by the specification.
const default_sampler_ratio = 1.0

// NewSpanSampler builds the sampler described by the environment.
func NewSpanSampler() trace.Sampler {
	legacy := strings.TrimSpace(os.Getenv(trace_sampler))
	standard := strings.TrimSpace(os.Getenv(traces_sampler))

	// OTEL_TRACE_SAMPLER is specific to this agent and predates the standard
	// variables, so it keeps precedence for deployments already using it.
	if legacy != "" {
		if standard != "" {
			log.Printf("Both %s and %s are set, %s takes precedence and %s=%s is ignored",
				trace_sampler, traces_sampler, trace_sampler, traces_sampler, standard)
		}

		return newRatioSampler(legacy)
	}

	if standard != "" {
		return newStandardSampler(standard, os.Getenv(traces_sampler_arg))
	}

	// Equivalent to the specification default, parentbased_always_on.
	return trace.ParentBased(trace.AlwaysSample())
}

// newRatioSampler handles OTEL_TRACE_SAMPLER, which takes a bare ratio.
func newRatioSampler(samplerStr string) trace.Sampler {
	sampler, err := strconv.ParseFloat(samplerStr, 64)
	if err != nil {
		log.Printf("Invalid OTEL_TRACE_SAMPLER value: %s, fallback to parent based sampler", samplerStr)
		return trace.ParentBased(trace.AlwaysSample())
	}

	if sampler <= 0 {
		return trace.NeverSample()
	} else if sampler >= 1 {
		return trace.AlwaysSample()
	} else {
		return trace.ParentBased(trace.TraceIDRatioBased(sampler))
	}
}

// newStandardSampler handles OTEL_TRACES_SAMPLER / OTEL_TRACES_SAMPLER_ARG as
// defined by the OpenTelemetry SDK environment variable specification.
func newStandardSampler(name, arg string) trace.Sampler {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "always_on":
		return trace.AlwaysSample()
	case "always_off":
		return trace.NeverSample()
	case "traceidratio":
		return trace.TraceIDRatioBased(parseSamplerRatio(arg))
	case "parentbased_always_on":
		return trace.ParentBased(trace.AlwaysSample())
	case "parentbased_always_off":
		return trace.ParentBased(trace.NeverSample())
	case "parentbased_traceidratio":
		return trace.ParentBased(trace.TraceIDRatioBased(parseSamplerRatio(arg)))
	case "jaeger_remote", "parentbased_jaeger_remote", "xray":
		// Valid names, but they need samplers that are not linked into the
		// agent.
		log.Printf("OTEL_TRACES_SAMPLER=%s is not supported (remote and X-Ray samplers are not linked into the agent), fallback to parent based sampler", name)
		return trace.ParentBased(trace.AlwaysSample())
	default:
		log.Printf("Unknown OTEL_TRACES_SAMPLER value: %s, fallback to parent based sampler", name)
		return trace.ParentBased(trace.AlwaysSample())
	}
}

// parseSamplerRatio reads OTEL_TRACES_SAMPLER_ARG for the *_traceidratio
// samplers. Anything the specification does not allow is logged and falls back
// to the default ratio.
func parseSamplerRatio(arg string) float64 {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		log.Printf("Missing OTEL_TRACES_SAMPLER_ARG, fallback to ratio %v", default_sampler_ratio)
		return default_sampler_ratio
	}

	ratio, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		log.Printf("Invalid OTEL_TRACES_SAMPLER_ARG value: %s, fallback to ratio %v", arg, default_sampler_ratio)
		return default_sampler_ratio
	}

	// NaN has to be rejected explicitly: every comparison against it is false,
	// so a range check alone lets it through.
	if math.IsNaN(ratio) || math.IsInf(ratio, 0) || ratio < 0 || ratio > 1 {
		log.Printf("Out of range OTEL_TRACES_SAMPLER_ARG value: %s, fallback to ratio %v", arg, default_sampler_ratio)
		return default_sampler_ratio
	}

	return ratio
}

package frontend

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/sudorandom/fauxrpc/private/registry"
	dashboardv1 "github.com/sudorandom/fauxrpc/proto/gen/dashboard/v1"
	"github.com/sudorandom/fauxrpc/proto/gen/dashboard/v1/dashboardv1connect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var startupTime = time.Now()

// MetricsCollector defines an interface for gathering server metrics.
// This allows the dashboard handler to be decoupled from the specific metrics
// implementation.
type MetricsCollector interface {
	TotalRequests() uint64
	TotalErrors() uint64
	RequestsPerSecond() float64
}

var _ dashboardv1connect.DashboardServiceHandler = (*Handler)(nil)

type Handler struct {
	registry registry.ServiceRegistry
	metrics  MetricsCollector
	version  string
	httpHost string
}

func NewHandler(
	registry registry.ServiceRegistry,
	metrics MetricsCollector,
	version string,
	httpHost string,
) *Handler {
	return &Handler{
		registry: registry,
		metrics:  metrics,
		version:  version,
		httpHost: httpHost,
	}
}

// GetDashboardSummary implements dashboardv1connect.DashboardServiceHandler.
func (h *Handler) GetDashboardSummary(
	ctx context.Context,
	req *connect.Request[dashboardv1.GetDashboardSummaryRequest],
	stream *connect.ServerStream[dashboardv1.GetDashboardSummaryResponse],
) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// totalRequests := h.metrics.TotalRequests()
			// totalErrors := h.metrics.TotalErrors()
			// var errorRate float64
			// if totalRequests > 0 {
			// 	errorRate = (float64(totalErrors) / float64(totalRequests)) * 100
			// }

			uniqueMethods := int64(0)
			serviceCount := int64(0)
			h.registry.ForEachService(func(sd protoreflect.ServiceDescriptor) bool {
				serviceCount++
				uniqueMethods += int64(sd.Methods().Len())
				return true
			})

			resp := &dashboardv1.GetDashboardSummaryResponse{
				// TotalRequests:     proto.Int64(int64(totalRequests)),
				// RequestsPerSecond: proto.Float64(h.metrics.RequestsPerSecond()),
				// TotalErrors:       proto.Int64(int64(totalErrors)),
				// ErrorRate:      proto.String(fmt.Sprintf("%.3f%%", errorRate)),
				UniqueServices: proto.Int64(int64(serviceCount)),
				UniqueMethods:  proto.Int64(int64(uniqueMethods)),
				Uptime:         proto.String(formatUptime(time.Since(startupTime))),
				GoVersion:      proto.String(runtime.Version()),
				FauxrpcVersion: proto.String(h.version),
				HttpHost:       proto.String(h.httpHost),
			}

			if err := stream.Send(resp); err != nil {
				// The client has likely disconnected.
				return err
			}
		}
	}
}

// formatUptime converts a duration into a human-readable string like "2d 14h 32m".
func formatUptime(d time.Duration) string {
	d = d.Round(time.Minute)
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	return strings.Join(parts, " ")
}

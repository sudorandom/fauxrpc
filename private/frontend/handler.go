package frontend

//go:generate go tool templ generate
//go:generate bash -c "./generate_assets.sh"

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/jhump/protoreflect/desc" //nolint:staticcheck
	"github.com/jhump/protoreflect/desc/protoprint"
	"github.com/sudorandom/fauxrpc/private/frontend/templates"
	"github.com/sudorandom/fauxrpc/private/frontend/templates/browser"
	"github.com/sudorandom/fauxrpc/private/frontend/templates/partials"
	templates_stubs "github.com/sudorandom/fauxrpc/private/frontend/templates/stubs"
	"github.com/sudorandom/fauxrpc/private/log"
	"github.com/sudorandom/fauxrpc/private/metrics"
	"github.com/sudorandom/fauxrpc/private/registry"
	"github.com/sudorandom/fauxrpc/private/stubs"
	"github.com/sudorandom/fauxrpc/protocel"
	"go.yaml.in/yaml/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"

	stubsv1 "github.com/sudorandom/fauxrpc/private/gen/stubs/v1"
)

//go:embed assets
var embeddedAssets embed.FS

// Provider defines the interface for providing server statistics and the logger.
type Provider interface {
	GetStats() *metrics.Stats
	GetLogger() *log.Logger
	registry.ServiceRegistry
	GetStubDB() stubs.StubDatabase
	HasProtobufServices() bool
	HasOpenAPIRoutes() bool
}

func buildStubGroups(p Provider, selectedTarget, selectedID string) ([]templates_stubs.StubGroup, *stubsv1.Stub, error) {
	allStubEntries := p.GetStubDB().GetStubs()
	pbStubs, err := stubs.StubsToProto(allStubEntries)
	if err != nil {
		return nil, nil, err
	}

	groupedStubs := make(map[string][]*stubsv1.Stub)
	for _, stub := range pbStubs {
		target := stub.GetRef().GetTarget()
		groupedStubs[target] = append(groupedStubs[target], stub)
	}

	var targets []string
	for target := range groupedStubs {
		targets = append(targets, target)
	}
	sort.Strings(targets)

	groups := make([]templates_stubs.StubGroup, 0, len(targets))
	var selected *stubsv1.Stub

	for _, target := range targets {
		list := groupedStubs[target]
		sort.Slice(list, func(i, j int) bool {
			if list[i].GetPriority() != list[j].GetPriority() {
				return list[i].GetPriority() > list[j].GetPriority()
			}
			return list[i].GetRef().GetId() < list[j].GetRef().GetId()
		})
		for _, s := range list {
			if selected == nil && selectedTarget != "" &&
				s.GetRef().GetTarget() == selectedTarget && s.GetRef().GetId() == selectedID {
				selected = s
			}
		}
		groups = append(groups, templates_stubs.StubGroup{
			Target: target,
			Stubs:  list,
		})
	}

	if selected == nil && len(groups) > 0 && len(groups[0].Stubs) > 0 {
		selected = groups[0].Stubs[0]
	}

	return groups, selected, nil
}

func printProtoFile(fd protoreflect.FileDescriptor) (string, error) {
	wrapped, err := desc.WrapFile(fd)
	if err != nil {
		return "", err
	}
	printer := protoprint.Printer{}
	return printer.PrintProtoToString(wrapped)
}

func buildSchemaCatalog(p Provider, requestedPath string) ([]browser.SchemaPackageGroup, *browser.SchemaFileEntry, string, error) {
	files := p.Files()
	byPkg := make(map[string][]browser.SchemaFileEntry)
	pkgHasServices := make(map[string]bool)
	fdByPath := make(map[string]protoreflect.FileDescriptor)

	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		pkg := string(fd.Package())
		if pkg == "" {
			pkg = "(default)"
		}
		filePath := fd.Path()
		fileName := filePath
		if lastSlash := strings.LastIndex(filePath, "/"); lastSlash != -1 {
			fileName = filePath[lastSlash+1:]
		}
		svcCount := fd.Services().Len()
		if svcCount > 0 {
			pkgHasServices[pkg] = true
		}
		entry := browser.SchemaFileEntry{
			Path:         filePath,
			FileName:     fileName,
			Package:      pkg,
			ServiceCount: svcCount,
			MessageCount: fd.Messages().Len(),
			EnumCount:    fd.Enums().Len(),
		}
		byPkg[pkg] = append(byPkg[pkg], entry)
		fdByPath[filePath] = fd
		return true
	})

	pkgs := make([]string, 0, len(byPkg))
	for pkg := range byPkg {
		pkgs = append(pkgs, pkg)
	}
	sort.Slice(pkgs, func(i, j int) bool {
		if pkgHasServices[pkgs[i]] != pkgHasServices[pkgs[j]] {
			return pkgHasServices[pkgs[i]]
		}
		return pkgs[i] < pkgs[j]
	})

	groups := make([]browser.SchemaPackageGroup, 0, len(pkgs))
	var selected *browser.SchemaFileEntry

	for _, pkg := range pkgs {
		entries := byPkg[pkg]
		sort.Slice(entries, func(i, j int) bool {
			if (entries[i].ServiceCount > 0) != (entries[j].ServiceCount > 0) {
				return entries[i].ServiceCount > 0
			}
			return entries[i].Path < entries[j].Path
		})
		for idx := range entries {
			e := &entries[idx]
			if requestedPath != "" && (e.Path == requestedPath || e.Package == requestedPath) && selected == nil {
				copied := *e
				selected = &copied
			}
		}
		groups = append(groups, browser.SchemaPackageGroup{
			Package: pkg,
			Files:   entries,
		})
	}

	if selected == nil && len(groups) > 0 && len(groups[0].Files) > 0 {
		copied := groups[0].Files[0]
		selected = &copied
	}

	var content string
	if selected != nil {
		if fd, ok := fdByPath[selected.Path]; ok {
			var err error
			content, err = printProtoFile(fd)
			if err != nil {
				return nil, nil, "", err
			}
		}
	}

	return groups, selected, content, nil
}

type stubFormUpdatePayload struct {
	ID           string `json:"id"`
	Target       string `json:"target"`
	Priority     int32  `json:"priority"`
	ActiveIf     string `json:"active_if"`
	JSONContent  string `json:"json_content"`
	CELContent   string `json:"cel_content"`
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

func parseStubUpdateBody(body []byte, contentType, defaultTarget, defaultID string) (stubs.StubFile, error) {
	if strings.Contains(strings.ToLower(contentType), "application/json") {
		var payload stubFormUpdatePayload
		if err := json.Unmarshal(body, &payload); err != nil {
			return stubs.StubFile{}, fmt.Errorf("invalid form payload: %w", err)
		}
		var parsedContent any
		if rawJSON := strings.TrimSpace(payload.JSONContent); rawJSON != "" {
			if err := json.Unmarshal([]byte(rawJSON), &parsedContent); err != nil {
				return stubs.StubFile{}, fmt.Errorf("invalid JSON response payload: %w", err)
			}
		}
		entry := stubs.StubFileEntry{
			ID:           strings.TrimSpace(payload.ID),
			Target:       strings.TrimSpace(payload.Target),
			Priority:     payload.Priority,
			ActiveIf:     strings.TrimSpace(payload.ActiveIf),
			Content:      parsedContent,
			CelContent:   strings.TrimSpace(payload.CELContent),
			ErrorCode:    payload.ErrorCode,
			ErrorMessage: strings.TrimSpace(payload.ErrorMessage),
		}
		if entry.ID == "" {
			entry.ID = defaultID
		}
		if entry.Target == "" {
			entry.Target = defaultTarget
		}
		return stubs.StubFile{Stubs: []stubs.StubFileEntry{entry}}, nil
	}

	var stubFile stubs.StubFile
	decoder := yaml.NewDecoder(bytes.NewReader(body))
	decoder.KnownFields(true)
	err := decoder.Decode(&stubFile)
	if err != nil || len(stubFile.Stubs) == 0 {
		var singleEntry stubs.StubFileEntry
		singleDecoder := yaml.NewDecoder(bytes.NewReader(body))
		singleDecoder.KnownFields(true)
		if singleErr := singleDecoder.Decode(&singleEntry); singleErr == nil && (singleEntry.Target != "" || defaultTarget != "") {
			stubFile = stubs.StubFile{Stubs: []stubs.StubFileEntry{singleEntry}}
			err = nil
		}
	}
	if err != nil {
		return stubs.StubFile{}, fmt.Errorf("invalid stub YAML: %w", err)
	}
	if len(stubFile.Stubs) == 0 {
		return stubs.StubFile{}, fmt.Errorf("no stubs found in YAML payload")
	}
	if stubFile.Stubs[0].ID == "" && defaultID != "" {
		stubFile.Stubs[0].ID = defaultID
	}
	if stubFile.Stubs[0].Target == "" && defaultTarget != "" {
		stubFile.Stubs[0].Target = defaultTarget
	}
	return stubFile, nil
}

type stubValidationResponse struct {
	Valid         bool   `json:"valid"`
	TargetError   string `json:"target_error,omitempty"`
	ActiveIfError string `json:"active_if_error,omitempty"`
	JSONError     string `json:"json_error,omitempty"`
	CELError      string `json:"cel_error,omitempty"`
	YAMLError     string `json:"yaml_error,omitempty"`
}

func validateStubFormPayload(p Provider, payload stubFormUpdatePayload) stubValidationResponse {
	res := stubValidationResponse{Valid: true}

	rawJSON := strings.TrimSpace(payload.JSONContent)
	if rawJSON != "" {
		var parsed any
		if err := json.Unmarshal([]byte(rawJSON), &parsed); err != nil {
			res.JSONError = fmt.Sprintf("Invalid JSON syntax: %v", err)
			res.Valid = false
		}
	}

	rawTarget := strings.TrimSpace(payload.Target)
	normalizedTarget := strings.ReplaceAll(rawTarget, "/", ".")
	var md protoreflect.MessageDescriptor
	var methodDesc protoreflect.MethodDescriptor

	if normalizedTarget != "" {
		if strings.Count(rawTarget, "/") > 1 {
			res.TargetError = fmt.Sprintf("target name has %d slashes when at most one is acceptable", strings.Count(rawTarget, "/"))
			res.Valid = false
		} else {
			desc, err := p.FindDescriptorByName(protoreflect.FullName(normalizedTarget))
			if err != nil {
				res.TargetError = fmt.Sprintf("unable to find object named %s", normalizedTarget)
				res.Valid = false
			} else {
				switch t := desc.(type) {
				case protoreflect.MethodDescriptor:
					methodDesc = t
					md = t.Output()
				case protoreflect.MessageDescriptor:
					md = t
				default:
					res.TargetError = fmt.Sprintf("target %s is not a method or message", normalizedTarget)
					res.Valid = false
				}
			}
		}
	}

	activeIf := strings.TrimSpace(payload.ActiveIf)
	if activeIf != "" {
		if methodDesc != nil {
			if _, err := stubs.NewActiveIf(methodDesc, activeIf); err != nil {
				res.ActiveIfError = err.Error()
				res.Valid = false
			}
		} else if md != nil {
			res.ActiveIfError = "active_if condition requires an RPC method target"
			res.Valid = false
		} else if normalizedTarget == "" {
			res.ActiveIfError = "Specify an RPC method target above to validate active_if against its request type"
			res.Valid = false
		}
	}

	if rawJSON != "" && res.JSONError == "" && md != nil {
		jsonOpts := protojson.UnmarshalOptions{Resolver: p.Resolver()}
		msg := registry.NewMessage(md).Interface()
		if err := jsonOpts.Unmarshal([]byte(rawJSON), msg); err != nil {
			res.JSONError = err.Error()
			res.Valid = false
		}
	}

	celContent := strings.TrimSpace(payload.CELContent)
	if celContent != "" {
		if md != nil {
			compiler, err := protocel.NewCompiler(p.Files())
			if err != nil {
				res.CELError = err.Error()
				res.Valid = false
			} else if _, err := compiler.Compile(md, celContent); err != nil {
				res.CELError = err.Error()
				res.Valid = false
			}
		} else if normalizedTarget == "" {
			res.CELError = "Specify a target above to validate CEL against its response message schema"
			res.Valid = false
		}
	}

	return res
}

func matchesLogFilter(entry *log.LogEntry, filter string) bool {
	if filter == "" {
		return true
	}
	switch {
	case strings.Contains(strings.ToLower(entry.Service), filter):
		return true
	case strings.Contains(strings.ToLower(entry.Method), filter):
		return true
	case strings.Contains(fmt.Sprintf("%d", entry.Status), filter):
		return true
	case strings.Contains(strings.ToLower(string(entry.RequestBody)), filter):
		return true
	case strings.Contains(strings.ToLower(string(entry.ResponseBody)), filter):
		return true
	case strings.Contains(strings.ToLower(string(entry.RequestHeaders)), filter):
		return true
	case strings.Contains(strings.ToLower(string(entry.ResponseHeaders)), filter):
		return true
	default:
		return false
	}
}

func DashboardHandler(p Provider) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/fauxrpc/sse/logs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := p.GetLogger()
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		history, ch, unsubscribe := logger.SubscribeWithHistory()
		defer unsubscribe()

		ctx := r.Context()
		if _, err := fmt.Fprintf(w, "event: Connected\ndata: ok\n\n"); err != nil {
			slog.Error("failed to write new entry to SSE stream", "err", err)
			return
		}

		filter := strings.ToLower(r.URL.Query().Get("filter"))

		// Send captured history first
		for _, entry := range history {
			if !matchesLogFilter(entry, filter) {
				continue
			}

			var buf bytes.Buffer
			if err := partials.LogEntry(entry, p).Render(ctx, &buf); err != nil {
				slog.Error("failed to render new log entry", "err", err)
				continue
			}

			html := buf.String()
			formattedHTML := strings.ReplaceAll(html, "\n", "\ndata: ")

			if _, err := fmt.Fprintf(w, "event: Request\ndata: %s\n\n", formattedHTML); err != nil {
				slog.Error("failed to write new entry to SSE stream", "err", err)
				return
			}
		}
		flusher.Flush()

		for {
			select {
			case <-ctx.Done():
				return // Client disconnected
			case entry, ok := <-ch:
				if !ok || entry == nil {
					return
				}
				if !matchesLogFilter(entry, filter) {
					continue
				}

				var buf bytes.Buffer
				if err := partials.LogEntry(entry, p).Render(ctx, &buf); err != nil {
					slog.Error("failed to render new log entry", "err", err)
					continue
				}

				html := buf.String()
				formattedHTML := strings.ReplaceAll(html, "\n", "\ndata: ")

				if _, err := fmt.Fprintf(w, "event: Request\ndata: %s\n\n", formattedHTML); err != nil {
					slog.Error("failed to write new entry to SSE stream", "err", err)
					return
				}
				flusher.Flush()
			}
		}
	}))

	mux.Handle("/fauxrpc/sse/summary", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return // Client disconnected
			case <-ticker.C:
				stats := p.GetStats()
				var buf bytes.Buffer
				if err := partials.Summary(stats).Render(ctx, &buf); err != nil {
					slog.Error("failed to render summary partial", "err", err)
					continue
				}
				formattedHTML := strings.ReplaceAll(buf.String(), "\n", "\ndata: ")
				if _, err := fmt.Fprintf(w, "event: Summary\ndata: %s\n\n", formattedHTML); err != nil {
					slog.Error("failed to write summary to SSE stream", "err", err)
					return
				}
				flusher.Flush()
			}
		}
	}))

	mux.Handle("/fauxrpc/stats/reset", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		p.GetStats().Reset()
		templ.Handler(partials.Summary(p.GetStats())).ServeHTTP(w, r)
	}))

	mux.Handle("/fauxrpc/logs/clear", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		p.GetLogger().Clear()
		w.WriteHeader(http.StatusNoContent)
	}))

	mux.Handle("/fauxrpc/logs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") == "true" {
			templ.Handler(partials.RequestLog(nil, p)).ServeHTTP(w, r)
		} else {
			templ.Handler(templates.Index(partials.RequestLog(nil, p), p)).ServeHTTP(w, r)
		}
	}))

	mux.Handle("/fauxrpc/browser/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/fauxrpc/browser")
		path = strings.TrimPrefix(path, "/")

		groups, selected, content, err := buildSchemaCatalog(p, path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Target") == "schema-viewer-pane" && selected != nil {
			templ.Handler(browser.FileContent(selected.Package, selected.Path, selected.FileName, content)).ServeHTTP(w, r)
			return
		}

		view := browser.Browser(groups, selected, content)
		if r.Header.Get("HX-Request") == "true" {
			templ.Handler(view).ServeHTTP(w, r)
		} else {
			templ.Handler(templates.Index(view, p)).ServeHTTP(w, r)
		}
	}))

	mux.Handle("/fauxrpc/stubs/register", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read YAML body: %v", err), http.StatusBadRequest)
			return
		}
		var stubFile stubs.StubFile
		decoder := yaml.NewDecoder(bytes.NewReader(body))
		decoder.KnownFields(true)
		if err := decoder.Decode(&stubFile); err != nil {
			http.Error(w, fmt.Sprintf("Invalid stub YAML: %v", err), http.StatusBadRequest)
			return
		}
		if len(stubFile.Stubs) == 0 {
			http.Error(w, "No stubs found in YAML payload", http.StatusBadRequest)
			return
		}
		req, err := stubFile.ToRequest()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to build stub request: %v", err), http.StatusBadRequest)
			return
		}
		h := stubs.NewHandler(p, p.GetStubDB())
		resp, err := h.AddStubs(r.Context(), req)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to register stub: %v", err), http.StatusBadRequest)
			return
		}
		registeredID := ""
		if len(resp.GetStubs()) > 0 {
			registeredID = resp.GetStubs()[0].GetRef().GetId()
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "ID: %s", registeredID)
	}))

	mux.Handle("/fauxrpc/stubs/validate", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read body: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		ct := strings.ToLower(r.Header.Get("Content-Type"))
		if strings.Contains(ct, "application/json") {
			var payload stubFormUpdatePayload
			if err := json.Unmarshal(body, &payload); err != nil {
				_ = json.NewEncoder(w).Encode(stubValidationResponse{
					Valid:     false,
					JSONError: fmt.Sprintf("invalid form payload: %v", err),
				})
				return
			}
			_ = json.NewEncoder(w).Encode(validateStubFormPayload(p, payload))
			return
		}

		stubFile, err := parseStubUpdateBody(body, ct, "", "")
		if err != nil {
			_ = json.NewEncoder(w).Encode(stubValidationResponse{
				Valid:     false,
				YAMLError: err.Error(),
			})
			return
		}
		req, err := stubFile.ToRequest()
		if err != nil {
			_ = json.NewEncoder(w).Encode(stubValidationResponse{
				Valid:     false,
				YAMLError: err.Error(),
			})
			return
		}
		tempHandler := stubs.NewHandler(p, stubs.NewStubDatabase())
		if _, err := tempHandler.AddStubs(r.Context(), req); err != nil {
			_ = json.NewEncoder(w).Encode(stubValidationResponse{
				Valid:     false,
				YAMLError: err.Error(),
			})
			return
		}
		_ = json.NewEncoder(w).Encode(stubValidationResponse{Valid: true})
	}))

	mux.Handle("/fauxrpc/stubs/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/fauxrpc/stubs")
		path = strings.TrimPrefix(path, "/")

		if r.Method == http.MethodDelete {
			if path == "" {
				p.GetStubDB().RemoveAllStubs()
			} else if lastSlash := strings.LastIndex(path, "/"); lastSlash != -1 {
				target := strings.ReplaceAll(path[:lastSlash], "/", ".")
				id := path[lastSlash+1:]
				p.GetStubDB().RemoveStub(stubs.StubKey{
					Name: protoreflect.FullName(target),
					ID:   id,
				})
			}
			groups, selected, err := buildStubGroups(p, "", "")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			templ.Handler(templates_stubs.List(groups, selected)).ServeHTTP(w, r)
			return
		}

		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			var oldTarget, oldID string
			if lastSlash := strings.LastIndex(path, "/"); lastSlash != -1 {
				oldTarget = strings.ReplaceAll(path[:lastSlash], "/", ".")
				oldID = path[lastSlash+1:]
			}

			body, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
				return
			}

			stubFile, err := parseStubUpdateBody(body, r.Header.Get("Content-Type"), oldTarget, oldID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			req, err := stubFile.ToRequest()
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to build stub request: %v", err), http.StatusBadRequest)
				return
			}

			h := stubs.NewHandler(p, p.GetStubDB())
			resp, err := h.AddStubs(r.Context(), req)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to update stub: %v", err), http.StatusBadRequest)
				return
			}

			newTarget := strings.ReplaceAll(stubFile.Stubs[0].Target, "/", ".")
			newID := stubFile.Stubs[0].ID
			if len(resp.GetStubs()) > 0 && resp.GetStubs()[0].GetRef().GetId() != "" {
				newID = resp.GetStubs()[0].GetRef().GetId()
			}

			if oldTarget != "" && oldID != "" && (oldTarget != newTarget || oldID != newID) {
				p.GetStubDB().RemoveStub(stubs.StubKey{
					Name: protoreflect.FullName(oldTarget),
					ID:   oldID,
				})
			}

			groups, selected, err := buildStubGroups(p, newTarget, newID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			templ.Handler(templates_stubs.List(groups, selected)).ServeHTTP(w, r)
			return
		}

		var selectedTarget, selectedID string
		if path != "" {
			if lastSlash := strings.LastIndex(path, "/"); lastSlash != -1 {
				selectedTarget = strings.ReplaceAll(path[:lastSlash], "/", ".")
				selectedID = path[lastSlash+1:]
			}
		}

		groups, selected, err := buildStubGroups(p, selectedTarget, selectedID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Target") == "stub-inspector-pane" && selected != nil {
			templ.Handler(templates_stubs.Single(selected)).ServeHTTP(w, r)
			return
		}

		view := templates_stubs.List(groups, selected)
		if r.Header.Get("HX-Request") == "true" {
			templ.Handler(view).ServeHTTP(w, r)
		} else {
			templ.Handler(templates.Index(view, p)).ServeHTTP(w, r)
		}
	}))

	mux.Handle("/fauxrpc/summary", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") == "true" {
			templ.Handler(partials.SummaryPage(p.GetStats())).ServeHTTP(w, r)
		} else {
			templ.Handler(templates.Index(partials.SummaryPage(p.GetStats()), p)).ServeHTTP(w, r)
		}
	}))

	mux.Handle("/fauxrpc/about", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") == "true" {
			templ.Handler(templates.About()).ServeHTTP(w, r)
		} else {
			templ.Handler(templates.Index(templates.About(), p)).ServeHTTP(w, r)
		}
	}))

	mux.Handle("/fauxrpc/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(templates.Index(partials.RequestLog(nil, p), p)).ServeHTTP(w, r)
	}))

	// Serve static assets from the embedded file system
	assetFS, err := fs.Sub(embeddedAssets, "assets")
	if err != nil {
		slog.Error("failed to create sub-filesystem for assets", "err", err)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Internal server error: asset loading failed", http.StatusInternalServerError)
		})
	}
	mux.Handle("/fauxrpc/assets/", http.StripPrefix("/fauxrpc/assets/", http.FileServer(http.FS(assetFS))))

	return mux
}

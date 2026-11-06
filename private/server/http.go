package server

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"connectrpc.com/connect/v2"
	"connectrpc.com/vanguard"
	"github.com/go-chi/chi/v5"
	"github.com/sudorandom/fauxrpc"
	"github.com/sudorandom/fauxrpc/private/registry"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	rootPathSentinel      = "__fauxrpc_root__"
	trailingSlashSentinel = "__fauxrpc_slash__"
)

type methodDescWithCustomOptions struct {
	protoreflect.MethodDescriptor
	opts protoreflect.ProtoMessage
}

func (m methodDescWithCustomOptions) Options() protoreflect.ProtoMessage {
	return m.opts
}

func wrapMethodWithNormalizedHTTP(md protoreflect.MethodDescriptor) protoreflect.MethodDescriptor {
	rawRule := fauxrpc.ExtractHTTPRule(md)
	if rawRule == nil {
		return md
	}
	opts := md.Options()
	if opts == nil {
		return md
	}
	rule := proto.Clone(rawRule).(*annotations.HttpRule)
	meta := &httpTranscoderMeta{
		customMethods: make(map[string]struct{}),
		verbs:         make(map[string]struct{}),
	}
	meta.normalizeRule(rule)

	cloned := proto.Clone(opts)
	extDesc := annotations.E_Http.TypeDescriptor()
	refl := cloned.ProtoReflect()
	var toClear []protoreflect.FieldDescriptor
	refl.Range(func(fd protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
		if fd.IsExtension() && fd.Number() == extDesc.Number() {
			toClear = append(toClear, fd)
		}
		return true
	})
	for _, fd := range toClear {
		refl.Clear(fd)
	}

	validRule := validateMethodHTTPRule(md, cloned, rule)
	if validRule != nil {
		proto.SetExtension(cloned, annotations.E_Http, validRule)
	}
	return methodDescWithCustomOptions{
		MethodDescriptor: md,
		opts:             cloned,
	}
}

func validateMethodHTTPRule(
	md protoreflect.MethodDescriptor,
	emptyOpts protoreflect.ProtoMessage,
	rule *annotations.HttpRule,
) *annotations.HttpRule {
	tryMount := func(candidate *annotations.HttpRule) error {
		optsCopy := proto.Clone(emptyOpts)
		proto.SetExtension(optsCopy, annotations.E_Http, candidate)
		wrapped := methodDescWithCustomOptions{MethodDescriptor: md, opts: optsCopy}
		srv := connect.NewServer()
		srv.Register(connect.Method{
			Spec: connect.Spec{
				Procedure: "/" + string(md.Parent().FullName()) + "/" + string(md.Name()),
				Schema:    wrapped,
			},
			Handler: func(_ context.Context, _ connect.Spec, _ connect.ServerStream) error {
				return nil
			},
		})
		return vanguard.Mount(&restMuxCapture{}, srv)
	}

	if err := tryMount(rule); err == nil {
		return rule
	} else if len(rule.GetAdditionalBindings()) == 0 {
		slog.Warn("Skipping invalid google.api.http rule", "method", md.FullName(), "error", err)
		return nil
	}

	// Try keeping primary binding and filtering individual additional_bindings.
	filtered := proto.Clone(rule).(*annotations.HttpRule)
	filtered.AdditionalBindings = nil
	if err := tryMount(filtered); err != nil {
		slog.Warn("Skipping invalid google.api.http rule", "method", md.FullName(), "error", err)
		return nil
	}
	for _, binding := range rule.GetAdditionalBindings() {
		filtered.AdditionalBindings = append(filtered.AdditionalBindings, binding)
		if err := tryMount(filtered); err != nil {
			slog.Warn("Skipping invalid google.api.http additional_binding", "method", md.FullName(), "error", err)
			filtered.AdditionalBindings = filtered.AdditionalBindings[:len(filtered.AdditionalBindings)-1]
		}
	}
	return filtered
}

type trailingSlashPattern struct {
	segments []string // segments before the trailing slash
	verb     string   // canonical verb, or ""
}

type httpTranscoderMeta struct {
	customMethods         map[string]struct{}
	verbs                 map[string]struct{}
	hasRoot               bool
	trailingSlashPatterns []trailingSlashPattern
}

func collectHTTPTranscoderMeta(s *server) *httpTranscoderMeta {
	meta := &httpTranscoderMeta{
		customMethods: make(map[string]struct{}),
		verbs:         make(map[string]struct{}),
	}
	s.ForEachService(func(sd protoreflect.ServiceDescriptor) bool {
		methods := sd.Methods()
		for i := 0; i < methods.Len(); i++ {
			md := methods.Get(i)
			rawRule := fauxrpc.ExtractHTTPRule(md)
			if rawRule == nil {
				continue
			}
			rule := proto.Clone(rawRule).(*annotations.HttpRule)
			meta.normalizeRule(rule)
		}
		return true
	})
	for m := range meta.customMethods {
		if m != "" && m != "*" {
			chi.RegisterMethod(m)
		}
	}
	return meta
}

func (m *httpTranscoderMeta) normalizeRule(rule *annotations.HttpRule) {
	if rule == nil {
		return
	}
	switch p := rule.GetPattern().(type) {
	case *annotations.HttpRule_Get:
		p.Get = m.normalizeTemplate(p.Get)
	case *annotations.HttpRule_Put:
		p.Put = m.normalizeTemplate(p.Put)
	case *annotations.HttpRule_Post:
		p.Post = m.normalizeTemplate(p.Post)
	case *annotations.HttpRule_Delete:
		p.Delete = m.normalizeTemplate(p.Delete)
	case *annotations.HttpRule_Patch:
		p.Patch = m.normalizeTemplate(p.Patch)
	case *annotations.HttpRule_Custom:
		if p.Custom != nil {
			if kind := p.Custom.GetKind(); kind != "" {
				m.customMethods[kind] = struct{}{}
			}
			p.Custom.Path = m.normalizeTemplate(p.Custom.GetPath())
		}
	}
	for _, sub := range rule.GetAdditionalBindings() {
		m.normalizeRule(sub)
	}
}

func (m *httpTranscoderMeta) normalizeTemplate(tpl string) string {
	if tpl == "" {
		return tpl
	}
	if tpl == "/" {
		m.hasRoot = true
		return "/" + rootPathSentinel
	}
	pathPart, verb := splitTemplateVerb(tpl)
	canonVerb := verb
	if verb != "" {
		if cv, ok := canonicalizeLiteralSegment(verb); ok {
			canonVerb = cv
			m.verbs[cv] = struct{}{}
		}
	}
	if len(pathPart) > 1 && strings.HasSuffix(pathPart, "/") {
		trimmed := strings.TrimPrefix(strings.TrimSuffix(pathPart, "/"), "/")
		m.trailingSlashPatterns = append(m.trailingSlashPatterns, trailingSlashPattern{
			segments: canonicalTemplateSegments(trimmed),
			verb:     canonVerb,
		})
		pathPart += trailingSlashSentinel
		if verb != "" {
			return pathPart + ":" + verb
		}
		return pathPart
	}
	return tpl
}

func canonicalTemplateSegments(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	i := 0
	for i < len(s) {
		if s[i] == '/' {
			i++
			continue
		}
		if s[i] == '{' {
			i++
			for i < len(s) && s[i] != '=' && s[i] != '}' {
				i++
			}
			if i < len(s) && s[i] == '}' {
				i++
				out = append(out, "*")
				continue
			}
			if i < len(s) && s[i] == '=' {
				i++
				start := i
				depth := 1
				for i < len(s) && depth > 0 {
					switch s[i] {
					case '{':
						depth++
					case '}':
						depth--
					}
					if depth > 0 {
						i++
					}
				}
				inner := s[start:i]
				if i < len(s) {
					i++
				}
				out = append(out, canonicalTemplateSegments(inner)...)
				continue
			}
		}
		start := i
		for i < len(s) && s[i] != '/' {
			i++
		}
		tok := s[start:i]
		if tok == "*" || tok == "**" {
			out = append(out, tok)
		} else if canon, ok := canonicalizeLiteralSegment(tok); ok {
			out = append(out, canon)
		} else {
			out = append(out, tok)
		}
	}
	return out
}

func (m *httpTranscoderMeta) matchesTrailingSlash(pathPart, verb string) bool {
	if len(m.trailingSlashPatterns) == 0 {
		return false
	}
	canonVerb := verb
	if verb != "" {
		if cv, ok := canonicalizeLiteralSegment(verb); ok {
			canonVerb = cv
		}
	}
	trimmed := strings.TrimPrefix(strings.TrimSuffix(pathPart, "/"), "/")
	var reqSegs []string
	if trimmed != "" {
		reqSegs = strings.Split(trimmed, "/")
		for i, seg := range reqSegs {
			if cv, ok := canonicalizeLiteralSegment(seg); ok {
				reqSegs[i] = cv
			}
		}
	}
	for _, pat := range m.trailingSlashPatterns {
		if pat.verb != canonVerb {
			continue
		}
		if matchTemplateSegments(pat.segments, reqSegs) {
			return true
		}
	}
	return false
}

func matchTemplateSegments(pat, req []string) bool {
	for i := 0; i < len(pat); i++ {
		if pat[i] == "**" {
			return true
		}
		if i >= len(req) {
			return false
		}
		if pat[i] != "*" && pat[i] != req[i] {
			return false
		}
	}
	return len(pat) == len(req)
}

func splitTemplateVerb(tpl string) (pathPart, verb string) {
	lastColon := strings.LastIndexByte(tpl, ':')
	lastBrace := strings.LastIndexByte(tpl, '}')
	lastSlash := strings.LastIndexByte(tpl, '/')
	if lastColon >= 0 && lastColon > lastBrace && lastColon > lastSlash {
		return tpl[:lastColon], tpl[lastColon+1:]
	}
	return tpl, ""
}

type restMuxCapture struct {
	handler http.Handler
}

func (c *restMuxCapture) Handle(_ string, h http.Handler) {
	c.handler = h
}

func mountVanguardHandler(
	rpcServer *connect.Server,
	resolver *registry.Resolver,
) (http.Handler, error) {
	capture := &restMuxCapture{}
	if err := vanguard.Mount(capture, rpcServer, vanguard.WithTypeResolver(resolver)); err != nil {
		return nil, err
	}
	if capture.handler == nil {
		return http.NotFoundHandler(), nil
	}
	return capture.handler, nil
}

func newTranscoderDispatcher(
	connectMux http.Handler,
	restHandler http.Handler,
	serviceNames []string,
	meta *httpTranscoderMeta,
) http.Handler {
	servicePrefixes := make(map[string]struct{}, len(serviceNames)+2)
	for _, name := range serviceNames {
		servicePrefixes["/"+name+"/"] = struct{}{}
	}
	servicePrefixes["/grpc.reflection.v1.ServerReflection/"] = struct{}{}
	servicePrefixes["/grpc.reflection.v1alpha.ServerReflection/"] = struct{}{}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if idx := strings.LastIndexByte(r.URL.Path, '/'); idx > 0 {
			if _, ok := servicePrefixes[r.URL.Path[:idx+1]]; ok {
				connectMux.ServeHTTP(w, r)
				return
			}
		}

		escaped := r.URL.EscapedPath()
		rewritten := escaped
		if meta.hasRoot && rewritten == "/" {
			rewritten = "/" + rootPathSentinel
		} else if len(meta.trailingSlashPatterns) > 0 && len(rewritten) > 1 {
			pathPart, verb := splitEscapedVerb(rewritten)
			if len(pathPart) > 1 && strings.HasSuffix(pathPart, "/") && meta.matchesTrailingSlash(pathPart, verb) {
				pathPart += trailingSlashSentinel
				if verb != "" {
					rewritten = pathPart + ":" + verb
				} else {
					rewritten = pathPart
				}
			}
		}

		if pathPart, verb := splitEscapedVerb(rewritten); verb != "" {
			if canonVerb, ok := canonicalizeLiteralSegment(verb); ok && canonVerb != verb {
				if _, isKnown := meta.verbs[canonVerb]; isKnown {
					rewritten = pathPart + ":" + canonVerb
				}
			}
		}

		if rewritten != escaped {
			if unescaped, err := url.PathUnescape(rewritten); err == nil {
				r2 := r.Clone(r.Context())
				r2.URL.Path = unescaped
				r2.URL.RawPath = rewritten
				r = r2
			}
		}
		restHandler.ServeHTTP(w, r)
	})
}

func splitEscapedVerb(uriPath string) (pathPart, verb string) {
	last := strings.LastIndexByte(uriPath, '/') + 1
	if colon := strings.LastIndexByte(uriPath[last:], ':'); colon >= 0 {
		return uriPath[:last+colon], uriPath[last+colon+1:]
	}
	return uriPath, ""
}

const upperhex = "0123456789ABCDEF"

func canonicalizeLiteralSegment(seg string) (string, bool) {
	unescaped, err := unescapeSingleSegment(seg)
	if err != nil {
		return "", false
	}
	return escapeSingleSegment(unescaped), true
}

func unescapeSingleSegment(input string) (string, error) {
	if !strings.ContainsRune(input, '%') {
		return input, nil
	}
	var sb strings.Builder
	sb.Grow(len(input))
	for i := 0; i < len(input); i++ {
		if input[i] == '%' {
			if i+2 >= len(input) || !isHexByte(input[i+1]) || !isHexByte(input[i+2]) {
				return "", url.EscapeError(input[i:])
			}
			sb.WriteByte((unhexByte(input[i+1]) << 4) | unhexByte(input[i+2]))
			i += 2
			continue
		}
		sb.WriteByte(input[i])
	}
	return sb.String(), nil
}

func escapeSingleSegment(input string) string {
	var sb strings.Builder
	sb.Grow(len(input))
	for i := 0; i < len(input); i++ {
		c := input[i]
		if isUnreservedPathByte(c) {
			sb.WriteByte(c)
		} else {
			sb.WriteByte('%')
			sb.WriteByte(upperhex[c>>4])
			sb.WriteByte(upperhex[c&15])
		}
	}
	return sb.String()
}

func isUnreservedPathByte(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '_' || c == '.' || c == '-' || c == '~'
}

func isHexByte(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

func unhexByte(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	default:
		return 0
	}
}

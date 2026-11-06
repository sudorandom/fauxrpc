package fauxrpc

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type methodDescCtxKey struct{}

// WithMethodDescriptor attaches the active RPC MethodDescriptor to ctx so fake
// data generation can tailor field values to the method's options (such as
// google.api.http path templates).
func WithMethodDescriptor(ctx context.Context, md protoreflect.MethodDescriptor) context.Context {
	if md == nil {
		return ctx
	}
	return context.WithValue(ctx, methodDescCtxKey{}, md)
}

// MethodDescriptorFromContext returns the MethodDescriptor attached to ctx, or nil.
func MethodDescriptorFromContext(ctx context.Context) protoreflect.MethodDescriptor {
	if ctx == nil {
		return nil
	}
	md, _ := ctx.Value(methodDescCtxKey{}).(protoreflect.MethodDescriptor)
	return md
}

type httpSegKind int

const (
	httpSegLiteral httpSegKind = iota
	httpSegSingleWildcard
	httpSegMultiWildcard
)

type httpPathSegment struct {
	kind        httpSegKind
	rawLiteral  string
	singleUnesc string
	multiUnesc  string
}

type httpVarCapture struct {
	fieldPath []string
	startSeg  int
	endSeg    int
}

type httpPathTemplate struct {
	raw      string
	segments []httpPathSegment
	captures []httpVarCapture
	verb     string
}

var (
	httpFileTemplatesCache   sync.Map // map[protoreflect.FileDescriptor]map[protoreflect.FullName][]*httpPathTemplate
	httpMethodTemplatesCache sync.Map // map[protoreflect.MethodDescriptor][]*httpPathTemplate
)

// ExtractHTTPRule safely extracts the google.api.http HttpRule from a MethodDescriptor,
// handling both typed *annotations.HttpRule values and dynamic extension messages.
func ExtractHTTPRule(md protoreflect.MethodDescriptor) *annotations.HttpRule {
	if md == nil {
		return nil
	}
	opts := md.Options()
	if opts == nil {
		return nil
	}
	extDesc := annotations.E_Http.TypeDescriptor()
	refl := opts.ProtoReflect()
	var extVal protoreflect.Value
	var found bool
	refl.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if fd.IsExtension() && fd.Number() == extDesc.Number() {
			extVal = v
			found = true
			return false
		}
		return true
	})
	if !found || !extVal.IsValid() {
		return nil
	}
	msgInterface := extVal.Message().Interface()
	if rule, ok := msgInterface.(*annotations.HttpRule); ok {
		return rule
	}
	b, err := proto.Marshal(msgInterface)
	if err != nil {
		return nil
	}
	rule := new(annotations.HttpRule)
	if err := proto.Unmarshal(b, rule); err != nil {
		return nil
	}
	return rule
}

func httpRulePatternPath(rule *annotations.HttpRule) string {
	if rule == nil {
		return ""
	}
	switch p := rule.GetPattern().(type) {
	case *annotations.HttpRule_Get:
		return p.Get
	case *annotations.HttpRule_Put:
		return p.Put
	case *annotations.HttpRule_Post:
		return p.Post
	case *annotations.HttpRule_Delete:
		return p.Delete
	case *annotations.HttpRule_Patch:
		return p.Patch
	case *annotations.HttpRule_Custom:
		if p.Custom != nil {
			return p.Custom.GetPath()
		}
	}
	return ""
}

func templatesForMethod(md protoreflect.MethodDescriptor) []*httpPathTemplate {
	if md == nil {
		return nil
	}
	if v, ok := httpMethodTemplatesCache.Load(md); ok {
		return v.([]*httpPathTemplate)
	}
	rule := ExtractHTTPRule(md)
	if rule == nil {
		httpMethodTemplatesCache.Store(md, ([]*httpPathTemplate)(nil))
		return nil
	}
	var out []*httpPathTemplate
	var collect func(r *annotations.HttpRule)
	collect = func(r *annotations.HttpRule) {
		if r == nil {
			return
		}
		if p := httpRulePatternPath(r); p != "" {
			if tpl, ok := parseHTTPPathTemplate(p); ok && len(tpl.captures) > 0 {
				out = append(out, tpl)
			}
		}
		for _, sub := range r.GetAdditionalBindings() {
			collect(sub)
		}
	}
	collect(rule)
	httpMethodTemplatesCache.Store(md, out)
	return out
}

func templatesForMessage(md protoreflect.MessageDescriptor, opts GenOptions) []*httpPathTemplate {
	if md == nil {
		return nil
	}
	if methodDesc := MethodDescriptorFromContext(opts.GetContext()); methodDesc != nil {
		if tpls := templatesForMethodAndMessage(methodDesc, md); len(tpls) > 0 {
			return tpls
		}
	}
	file := md.ParentFile()
	if file == nil {
		return nil
	}
	var byMsg map[protoreflect.FullName][]*httpPathTemplate
	if v, ok := httpFileTemplatesCache.Load(file); ok {
		byMsg = v.(map[protoreflect.FullName][]*httpPathTemplate)
	} else {
		byMsg = buildFileHTTPTemplates(file)
		httpFileTemplatesCache.Store(file, byMsg)
	}
	return byMsg[md.FullName()]
}

func templatesForMethodAndMessage(method protoreflect.MethodDescriptor, md protoreflect.MessageDescriptor) []*httpPathTemplate {
	tpls := templatesForMethod(method)
	if len(tpls) == 0 {
		return nil
	}
	if method.Input().FullName() == md.FullName() {
		return tpls
	}
	byMsg := make(map[protoreflect.FullName][]*httpPathTemplate)
	addMethodDerivedTemplates(method, tpls, byMsg)
	return byMsg[md.FullName()]
}

func buildFileHTTPTemplates(file protoreflect.FileDescriptor) map[protoreflect.FullName][]*httpPathTemplate {
	byMsg := make(map[protoreflect.FullName][]*httpPathTemplate)
	svcs := file.Services()
	for i := 0; i < svcs.Len(); i++ {
		methods := svcs.Get(i).Methods()
		for j := 0; j < methods.Len(); j++ {
			method := methods.Get(j)
			tpls := templatesForMethod(method)
			if len(tpls) == 0 {
				continue
			}
			inName := method.Input().FullName()
			byMsg[inName] = append(byMsg[inName], tpls...)
			addMethodDerivedTemplates(method, tpls, byMsg)
		}
	}
	return byMsg
}

// addMethodDerivedTemplates propagates structured resource path templates
// (such as {book.name=shelves/*/books/*} or GetBookRequest.name -> Book.name)
// to nested resource messages and response messages.
func addMethodDerivedTemplates(
	method protoreflect.MethodDescriptor,
	tpls []*httpPathTemplate,
	byMsg map[protoreflect.FullName][]*httpPathTemplate,
) {
	inputDesc := method.Input()
	outputDesc := method.Output()
	for _, tpl := range tpls {
		for _, cap := range tpl.captures {
			subSegs := tpl.segments[cap.startSeg:cap.endSeg]
			if !hasStructuredSegments(subSegs) {
				continue
			}
			// 1. Nested message field path (e.g. {book.name=shelves/*/books/*} -> Book.name)
			if len(cap.fieldPath) > 1 {
				cur := inputDesc
				for k := 0; k < len(cap.fieldPath)-1 && cur != nil; k++ {
					fd := cur.Fields().ByName(protoreflect.Name(cap.fieldPath[k]))
					if fd == nil || fd.Kind() != protoreflect.MessageKind || fd.IsList() || fd.IsMap() {
						cur = nil
						break
					}
					cur = fd.Message()
				}
				if cur != nil {
					leafName := cap.fieldPath[len(cap.fieldPath)-1]
					if leafFD := cur.Fields().ByName(protoreflect.Name(leafName)); leafFD != nil && leafFD.Kind() == protoreflect.StringKind {
						derived := &httpPathTemplate{
							segments: subSegs,
							captures: []httpVarCapture{
								{
									fieldPath: []string{leafName},
									startSeg:  0,
									endSeg:    len(subSegs),
								},
							},
						}
						byMsg[cur.FullName()] = append(byMsg[cur.FullName()], derived)
					}
				}
			}
			// 2. Top-level resource identifier shared between Request and Response (e.g. GetBookRequest.name -> Book.name)
			if len(cap.fieldPath) == 1 && outputDesc != nil && outputDesc.FullName() != inputDesc.FullName() {
				fieldName := cap.fieldPath[0]
				if outFD := outputDesc.Fields().ByName(protoreflect.Name(fieldName)); outFD != nil && outFD.Kind() == protoreflect.StringKind && !outFD.IsList() {
					derived := &httpPathTemplate{
						segments: subSegs,
						captures: []httpVarCapture{
							{
								fieldPath: []string{fieldName},
								startSeg:  0,
								endSeg:    len(subSegs),
							},
						},
					}
					byMsg[outputDesc.FullName()] = append(byMsg[outputDesc.FullName()], derived)
				}
			}
		}
	}
}

func hasStructuredSegments(segs []httpPathSegment) bool {
	if len(segs) > 1 {
		return true
	}
	for _, s := range segs {
		if s.kind == httpSegLiteral || s.kind == httpSegMultiWildcard {
			return true
		}
	}
	return false
}

func parseHTTPPathTemplate(raw string) (*httpPathTemplate, bool) {
	s := strings.TrimSpace(raw)
	if s == "" || !strings.HasPrefix(s, "/") {
		return nil, false
	}
	s = strings.TrimPrefix(s, "/")

	tpl := &httpPathTemplate{raw: raw}
	lastColon := strings.LastIndexByte(s, ':')
	lastBrace := strings.LastIndexByte(s, '}')
	lastSlash := strings.LastIndexByte(s, '/')
	if lastColon >= 0 && lastColon > lastBrace && lastColon > lastSlash {
		tpl.verb = s[lastColon+1:]
		s = s[:lastColon]
	}

	if !parseHTTPPathSegments(s, tpl) {
		return nil, false
	}
	return tpl, true
}

func parseHTTPPathSegments(s string, tpl *httpPathTemplate) bool {
	if s == "" {
		return true
	}
	i := 0
	for i < len(s) {
		if s[i] == '/' {
			i++
			continue
		}
		if s[i] == '{' {
			// Find '=' or '}' for the variable name
			i++
			nameStart := i
			for i < len(s) && s[i] != '=' && s[i] != '}' && s[i] != '/' && s[i] != '{' {
				i++
			}
			if i >= len(s) || i == nameStart {
				return false
			}
			fieldPathStr := s[nameStart:i]
			fieldPath := strings.Split(fieldPathStr, ".")
			for _, part := range fieldPath {
				if part == "" {
					return false
				}
			}
			startSeg := len(tpl.segments)
			switch s[i] {
			case '}':
				i++
				tpl.segments = append(tpl.segments, httpPathSegment{kind: httpSegSingleWildcard})
			case '=':
				i++
				innerStart := i
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
				if depth != 0 {
					return false
				}
				inner := s[innerStart:i]
				i++ // consume '}'
				if inner == "" {
					return false
				}
				if !parseHTTPPathSegments(inner, tpl) {
					return false
				}
			default:
				return false
			}
			endSeg := len(tpl.segments)
			tpl.captures = append(tpl.captures, httpVarCapture{
				fieldPath: fieldPath,
				startSeg:  startSeg,
				endSeg:    endSeg,
			})
			if i < len(s) && s[i] != '/' {
				return false
			}
			continue
		}

		segStart := i
		for i < len(s) && s[i] != '/' {
			if s[i] == '{' || s[i] == '}' {
				return false
			}
			i++
		}
		tok := s[segStart:i]
		switch tok {
		case "":
			continue
		case "*":
			tpl.segments = append(tpl.segments, httpPathSegment{kind: httpSegSingleWildcard})
		case "**":
			tpl.segments = append(tpl.segments, httpPathSegment{kind: httpSegMultiWildcard})
		default:
			tpl.segments = append(tpl.segments, httpPathSegment{
				kind:        httpSegLiteral,
				rawLiteral:  tok,
				singleUnesc: unescapeHTTPPathLiteral(tok, false),
				multiUnesc:  unescapeHTTPPathLiteral(tok, true),
			})
		}
	}
	return true
}

func unescapeHTTPPathLiteral(s string, multi bool) string {
	if !strings.ContainsRune(s, '%') {
		return s
	}
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) {
			hi, ok1 := fromHexChar(s[i+1])
			lo, ok2 := fromHexChar(s[i+2])
			if ok1 && ok2 {
				b := (hi << 4) | lo
				if multi && b == '/' {
					out = append(out, '%', '2', 'F')
				} else {
					out = append(out, b)
				}
				i += 2
				continue
			}
		}
		out = append(out, s[i])
	}
	return string(out)
}

func fromHexChar(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	default:
		return 0, false
	}
}

func setHTTPPathDataOnMessage(msg protoreflect.Message, opts GenOptions) {
	if opts.ViolateRules > 0 {
		return
	}
	tpls := templatesForMessage(msg.Descriptor(), opts)
	if len(tpls) == 0 {
		return
	}
	// Apply the primary HTTP rule template (index 0).
	applyHTTPPathTemplate(msg, tpls[0], opts)
}

func applyHTTPPathTemplate(msg protoreflect.Message, tpl *httpPathTemplate, opts GenOptions) {
	if tpl == nil || len(tpl.captures) == 0 {
		return
	}
	singleVal := make([]string, len(tpl.segments))
	multiVal := make([]string, len(tpl.segments))

	// Map each segment index to any single-segment capture that binds directly to it.
	singleCapBySeg := make(map[int]httpVarCapture)
	for _, cap := range tpl.captures {
		if cap.endSeg-cap.startSeg == 1 && tpl.segments[cap.startSeg].kind == httpSegSingleWildcard {
			singleCapBySeg[cap.startSeg] = cap
		}
	}

	for i, seg := range tpl.segments {
		switch seg.kind {
		case httpSegLiteral:
			singleVal[i] = seg.singleUnesc
			multiVal[i] = seg.multiUnesc
		case httpSegMultiWildcard:
			s := generateHTTPPathSlug(opts) + "/" + generateHTTPPathSlug(opts)
			singleVal[i] = s
			multiVal[i] = s
		case httpSegSingleWildcard:
			if cap, ok := singleCapBySeg[i]; ok {
				if val, ok := generateOrReadSingleSegmentField(msg, cap.fieldPath, opts); ok {
					singleVal[i] = val
					multiVal[i] = val
					continue
				}
			}
			s := generateHTTPPathSlug(opts)
			singleVal[i] = s
			multiVal[i] = s
		}
	}

	for _, cap := range tpl.captures {
		if cap.endSeg <= cap.startSeg {
			continue
		}
		var captured string
		if cap.endSeg-cap.startSeg == 1 && tpl.segments[cap.startSeg].kind != httpSegMultiWildcard {
			captured = singleVal[cap.startSeg]
		} else {
			captured = strings.Join(multiVal[cap.startSeg:cap.endSeg], "/")
		}
		setStringAtFieldPath(msg, cap.fieldPath, captured)
	}
}

func generateOrReadSingleSegmentField(msg protoreflect.Message, fieldPath []string, opts GenOptions) (string, bool) {
	targetMsg, leafFD := resolveFieldPathForWrite(msg, fieldPath)
	if targetMsg == nil || leafFD == nil || leafFD.IsList() || leafFD.IsMap() {
		return "", false
	}
	if leafFD.Kind() == protoreflect.StringKind {
		if getFieldConstraints(leafFD, opts) != nil && targetMsg.Has(leafFD) {
			existing := targetMsg.Get(leafFD).String()
			if isValidSinglePathSegment(existing) {
				return existing, true
			}
		}
		val := generateSingleSegmentForField(leafFD, opts)
		targetMsg.Set(leafFD, protoreflect.ValueOfString(val))
		return val, true
	}

	// Non-string scalar path variable (e.g. int32, int64, uint32, bool, enum)
	if !targetMsg.Has(leafFD) {
		if v := FieldValue(leafFD, opts); v != nil {
			targetMsg.Set(leafFD, *v)
		}
	}
	v := targetMsg.Get(leafFD)
	switch leafFD.Kind() {
	case protoreflect.EnumKind:
		if ev := leafFD.Enum().Values().ByNumber(v.Enum()); ev != nil {
			return string(ev.Name()), true
		}
		return strconv.FormatInt(int64(v.Enum()), 10), true
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return strconv.FormatInt(v.Int(), 10), true
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return strconv.FormatUint(v.Uint(), 10), true
	case protoreflect.BoolKind:
		return strconv.FormatBool(v.Bool()), true
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return fmt.Sprint(v.Float()), true
	default:
		return "", false
	}
}

func resolveFieldPathForWrite(msg protoreflect.Message, fieldPath []string) (protoreflect.Message, protoreflect.FieldDescriptor) {
	cur := msg
	for i := 0; i < len(fieldPath)-1; i++ {
		fd := cur.Descriptor().Fields().ByName(protoreflect.Name(fieldPath[i]))
		if fd == nil || fd.Kind() != protoreflect.MessageKind || fd.IsList() || fd.IsMap() {
			return nil, nil
		}
		cur = cur.Mutable(fd).Message()
	}
	leafFD := cur.Descriptor().Fields().ByName(protoreflect.Name(fieldPath[len(fieldPath)-1]))
	if leafFD == nil {
		return nil, nil
	}
	return cur, leafFD
}

func setStringAtFieldPath(msg protoreflect.Message, fieldPath []string, val string) {
	targetMsg, leafFD := resolveFieldPathForWrite(msg, fieldPath)
	if targetMsg == nil || leafFD == nil || leafFD.IsList() || leafFD.IsMap() {
		return
	}
	if leafFD.Kind() == protoreflect.StringKind {
		targetMsg.Set(leafFD, protoreflect.ValueOfString(val))
	}
}

func stringByHTTPPath(fd protoreflect.FieldDescriptor, opts GenOptions) (val string, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			val, ok = "", false
		}
	}()
	if fd == nil || fd.Kind() != protoreflect.StringKind || opts.ViolateRules > 0 {
		return "", false
	}
	parent := fd.ContainingMessage()
	if parent == nil {
		return "", false
	}
	tpls := templatesForMessage(parent, opts)
	if len(tpls) == 0 {
		return "", false
	}
	fieldName := string(fd.Name())
	for _, tpl := range tpls {
		for _, cap := range tpl.captures {
			if len(cap.fieldPath) == 1 && cap.fieldPath[0] == fieldName {
				return renderCaptureValue(tpl, cap, fd, opts), true
			}
		}
	}
	return "", false
}

func renderCaptureValue(tpl *httpPathTemplate, cap httpVarCapture, fd protoreflect.FieldDescriptor, opts GenOptions) string {
	if cap.endSeg-cap.startSeg == 1 && tpl.segments[cap.startSeg].kind != httpSegMultiWildcard {
		seg := tpl.segments[cap.startSeg]
		if seg.kind == httpSegLiteral {
			return seg.singleUnesc
		}
		return generateSingleSegmentForField(fd, opts)
	}
	parts := make([]string, 0, cap.endSeg-cap.startSeg)
	for i := cap.startSeg; i < cap.endSeg; i++ {
		seg := tpl.segments[i]
		switch seg.kind {
		case httpSegLiteral:
			parts = append(parts, seg.multiUnesc)
		case httpSegMultiWildcard:
			parts = append(parts, generateHTTPPathSlug(opts)+"/"+generateHTTPPathSlug(opts))
		default:
			parts = append(parts, generateHTTPPathSlug(opts))
		}
	}
	return strings.Join(parts, "/")
}

func generateSingleSegmentForField(fd protoreflect.FieldDescriptor, opts GenOptions) string {
	if s, ok := stringByHeuristics(fd, opts); ok && isValidSinglePathSegment(s) {
		return s
	}
	return generateHTTPPathSlug(opts)
}

func isValidSinglePathSegment(s string) bool {
	if s == "" || s == "." || s == ".." {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		isUnreserved := (c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' || c == '~'
		if !isUnreserved {
			return false
		}
	}
	return true
}

func generateHTTPPathSlug(opts GenOptions) string {
	b := strings.Builder{}
	for _, r := range strings.ToLower(opts.fake().Word()) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		b.WriteString(strings.ToLower(opts.fake().LetterN(6)))
	}
	b.WriteByte('-')
	b.WriteString(strconv.Itoa(opts.fake().IntRange(1, 999)))
	return b.String()
}

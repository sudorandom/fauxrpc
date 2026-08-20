package fauxrpc

import (
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	// maxFieldMaskPathDepth is how many message levels a single generated path
	// may walk down. "author.address.city" is depth 3.
	maxFieldMaskPathDepth = 3

	// maxFieldMaskPaths is the largest number of paths a generated mask holds
	// before normalization.
	maxFieldMaskPaths = 5
)

// GoogleFieldMask generates a random google.protobuf.FieldMask value.
//
// A field mask is only meaningful relative to some other message: its paths
// have to name fields that actually exist, or the mask is useless to whatever
// reads it. So the message the mask most likely applies to is guessed from the
// mask's surroundings (see fieldMaskTarget) and paths are drawn from that
// message's fields.
func GoogleFieldMask(fd protoreflect.FieldDescriptor, opts GenOptions) *fieldmaskpb.FieldMask {
	target, skip := fieldMaskTarget(fd)
	return fieldMask(target, skip, opts)
}

// fieldMask draws paths from target, inventing them when target is nil.
func fieldMask(target protoreflect.MessageDescriptor, skip protoreflect.FieldDescriptor, opts GenOptions) *fieldmaskpb.FieldMask {
	mask := &fieldmaskpb.FieldMask{}
	seen := map[string]struct{}{}
	for range opts.fake().IntRange(1, maxFieldMaskPaths) {
		var path string
		if target != nil {
			path = fieldMaskPath(target, skip, maxFieldMaskPathDepth, opts)
		} else {
			path = inventedFieldMaskPath(opts)
		}
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		mask.Paths = append(mask.Paths, path)
	}
	// Sorts the paths and drops any that a broader path already covers, so
	// "author" and "author.name" never show up together.
	mask.Normalize()
	return mask
}

// setFieldMaskOnMessage fills in a FieldMask that is generated on its own,
// rather than as some message's field, which leaves nothing around it to say
// what it masks.
//
// Treating it as an ordinary message is not an option: that fills paths with
// random sentences, and protojson refuses to marshal a path that is not
// lower_snake_case, so the whole response would fail to serialize.
func setFieldMaskOnMessage(msg protoreflect.Message, opts GenOptions) {
	fd := msg.Descriptor().Fields().ByName("paths")
	if fd == nil {
		return
	}
	list := msg.Mutable(fd).List()
	for _, path := range fieldMask(nil, nil, opts).GetPaths() {
		list.Append(protoreflect.ValueOfString(path))
	}
}

// fieldMaskTarget guesses which message the mask at fd describes. It also
// returns a field to leave out of the generated paths, which is set when the
// mask ends up pointing at the message holding it: a mask naming itself is
// technically legal but never what anyone means.
func fieldMaskTarget(fd protoreflect.FieldDescriptor) (protoreflect.MessageDescriptor, protoreflect.FieldDescriptor) {
	parent := fd.ContainingMessage()
	if parent == nil {
		return nil, nil
	}
	if parent.IsMapEntry() {
		// A mask used as a map value hangs off the synthetic entry message,
		// whose only other field is the map key. Resolve against the message
		// that holds the map instead, so the paths describe something.
		mapField := enclosingMapField(parent)
		if mapField == nil {
			return nil, nil
		}
		return fieldMaskTarget(mapField)
	}
	if subject := fieldMaskSubject(parent, fd); subject != nil {
		return subject, nil
	}
	// Nothing in the message looks like the thing being masked, so fall back to
	// masking the fields of the message itself. Common for requests that name
	// their resource by ID rather than carrying it: GetBookRequest{name,
	// read_mask} has nothing better to offer.
	return parent, fd
}

// enclosingMapField returns the map field whose values are entry messages.
// Returns nil when entry is not nested in a message, which the compiler does
// not produce for map entries.
func enclosingMapField(entry protoreflect.MessageDescriptor) protoreflect.FieldDescriptor {
	outer, ok := entry.Parent().(protoreflect.MessageDescriptor)
	if !ok {
		return nil
	}
	fields := outer.Fields()
	for i := range fields.Len() {
		field := fields.Get(i)
		if field.IsMap() && field.Message().FullName() == entry.FullName() {
			return field
		}
	}
	return nil
}

// fieldMaskSubject looks for the sibling field that mask describes.
//
// Two conventions cover most real schemas. A mask named "<x>_mask" describes a
// sibling field named "<x>", and a request named "<Verb><Resource>Request"
// describes its "<Resource>" field, which is how AIP-134 update requests are
// laid out: UpdateBookRequest{Book book, FieldMask update_mask}.
func fieldMaskSubject(parent protoreflect.MessageDescriptor, mask protoreflect.FieldDescriptor) protoreflect.MessageDescriptor {
	base := strings.TrimSuffix(string(mask.Name()), "_mask")
	if base == string(mask.Name()) {
		base = ""
	}
	hint := strings.TrimSuffix(string(parent.Name()), "Request")

	var byName, byResource, first protoreflect.MessageDescriptor
	fields := parent.Fields()
	for i := range fields.Len() {
		field := fields.Get(i)
		if !isFieldMaskSubjectCandidate(field, mask) {
			continue
		}
		switch {
		case base != "" && string(field.Name()) == base:
			byName = field.Message()
		case strings.HasSuffix(hint, string(field.Message().Name())):
			byResource = field.Message()
		}
		if first == nil {
			first = field.Message()
		}
	}

	switch {
	case byName != nil:
		return byName
	case byResource != nil:
		return byResource
	default:
		// No convention matched. The lowest-numbered message field is still a
		// better guess than the request itself, and is nil when there is none.
		return first
	}
}

// isFieldMaskSubjectCandidate reports whether field could be the thing that
// mask describes. Lists and maps are excluded: a mask cannot address the
// contents of either, so it would have nothing to say about them.
func isFieldMaskSubjectCandidate(field, mask protoreflect.FieldDescriptor) bool {
	if field.FullName() == mask.FullName() {
		return false
	}
	if field.IsList() || field.IsMap() || field.Kind() != protoreflect.MessageKind {
		return false
	}
	return !isWellKnownMessage(field.Message())
}

// fieldMaskPath builds one random path into md, descending into nested messages
// some of the time. Returns "" when md has no field worth naming.
func fieldMaskPath(md protoreflect.MessageDescriptor, skip protoreflect.FieldDescriptor, depth int, opts GenOptions) string {
	fields := md.Fields()
	candidates := make([]protoreflect.FieldDescriptor, 0, fields.Len())
	for i := range fields.Len() {
		field := fields.Get(i)
		if skip != nil && field.FullName() == skip.FullName() {
			continue
		}
		candidates = append(candidates, field)
	}
	if len(candidates) == 0 {
		return ""
	}

	field := candidates[opts.fake().IntRange(0, len(candidates)-1)]
	path := string(field.Name())
	if depth <= 1 || !canDescendFieldMask(field) {
		return path
	}
	// Roughly a third of the paths reach into the nested message instead of
	// stopping at it, so masks carry a mix of both.
	if opts.fake().IntRange(0, 2) != 0 {
		return path
	}
	sub := fieldMaskPath(field.Message(), nil, depth-1, opts)
	if sub == "" {
		return path
	}
	return path + "." + sub
}

// inventedFieldMaskPath builds one path for a mask with no known target. The
// segments name nothing real, but a path still has to look like a path.
func inventedFieldMaskPath(opts GenOptions) string {
	segments := make([]string, 0, maxFieldMaskPathDepth)
	for len(segments) < maxFieldMaskPathDepth {
		segment := fieldMaskSegment(opts)
		if segment == "" {
			break
		}
		segments = append(segments, segment)
		// Same mix as the targeted paths: roughly a third of them go one level
		// deeper than the level they are already at.
		if opts.fake().IntRange(0, 2) != 0 {
			break
		}
	}
	return strings.Join(segments, ".")
}

// fieldMaskSegment invents a single path segment, one or two words long.
// Returns "" when the words drawn hold nothing a segment can use.
func fieldMaskSegment(opts GenOptions) string {
	words := make([]string, 0, 2)
	for range opts.fake().IntRange(1, 2) {
		if word := fieldMaskWord(opts); word != "" {
			words = append(words, word)
		}
	}
	return strings.Join(words, "_")
}

// fieldMaskWord lowercases a random word and drops every character a protobuf
// field name cannot hold, leading digits included. protojson only accepts a
// mask path that survives a round trip through camelCase, so a stray hyphen or
// capital in a path is enough to make the message unserializable.
func fieldMaskWord(opts GenOptions) string {
	word := strings.Builder{}
	for _, r := range strings.ToLower(opts.fake().Word()) {
		switch {
		case r >= 'a' && r <= 'z':
			word.WriteRune(r)
		case r >= '0' && r <= '9' && word.Len() > 0:
			word.WriteRune(r)
		}
	}
	return word.String()
}

// canDescendFieldMask reports whether a path may continue past field. Only
// singular message fields can be walked into: field mask paths cannot name a
// field inside a list or a map entry.
func canDescendFieldMask(field protoreflect.FieldDescriptor) bool {
	if field.IsList() || field.IsMap() || field.Kind() != protoreflect.MessageKind {
		return false
	}
	// Naming the innards of a well-known type ("create_time.seconds") is legal
	// but not how these types are used, so paths stop at them.
	return !isWellKnownMessage(field.Message())
}

func isWellKnownMessage(md protoreflect.MessageDescriptor) bool {
	return md.FullName().Parent() == "google.protobuf"
}

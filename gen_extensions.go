package fauxrpc

import (
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// extensionSetProbability is how often a single extension of a message is
// filled in. Extensions are add-ons, so a message that has some of them set and
// some not is more like the real thing than one carrying every extension its
// schema declares.
const extensionSetProbability = 0.5

// maxExtensionListItems is how many items a repeated extension gets.
const maxExtensionListItems = 3

// setExtensionsOnMessage fills in some of the extensions declared for msg,
// which is a no-op unless opts carries a resolver to find them with.
//
// Extension fields are invisible to the loop over msg's own fields: the fields
// of a message descriptor are the ones it declares, and an extension is
// declared elsewhere, against a number the message only reserved.
func setExtensionsOnMessage(msg protoreflect.Message, opts GenOptions) {
	if opts.Extensions == nil || opts.MaxDepth <= 0 {
		return
	}
	desc := msg.Descriptor()
	if desc.ExtensionRanges().Len() == 0 || isExtensibleOptions(desc) {
		return
	}
	opts.Extensions.RangeExtensionsByMessage(desc.FullName(), func(xt protoreflect.ExtensionType) bool {
		if opts.fake().Float64() >= extensionSetProbability {
			return true
		}
		if val, ok := extensionValue(xt, opts); ok {
			msg.Set(xt.TypeDescriptor(), val)
		}
		return true
	})
}

// isExtensibleOptions reports whether md is one of the descriptor.proto options
// messages, which every custom option in a schema extends.
//
// These are skipped. Their extensions describe a schema rather than carry data,
// so filling them in would hand back a message claiming, say, a field rule that
// nothing generated actually honours, and anything reading the schema back out
// would be reading fiction.
func isExtensibleOptions(md protoreflect.MessageDescriptor) bool {
	return md.FullName().Parent() == "google.protobuf" && strings.HasSuffix(string(md.Name()), "Options")
}

// extensionValue generates a value for one extension. The value is built from
// the extension's own type rather than through FieldValue, because an extension
// value has to match the type that holds it: the message it extends may be a
// generated Go type or a dynamic one, and so may the extension.
func extensionValue(xt protoreflect.ExtensionType, opts GenOptions) (protoreflect.Value, bool) {
	xd := xt.TypeDescriptor()
	switch {
	case xd.IsList():
		list := xt.New().List()
		for range opts.fake().IntRange(1, maxExtensionListItems) {
			if item, ok := extensionItem(list.NewElement(), xd, opts); ok {
				list.Append(item)
			}
		}
		if list.Len() == 0 {
			return protoreflect.Value{}, false
		}
		return protoreflect.ValueOfList(list), true
	case xd.IsMap():
		// Not reachable: a map field cannot be an extension.
		return protoreflect.Value{}, false
	default:
		return extensionItem(xt.New(), xd, opts)
	}
}

// extensionItem generates a single extension value, filling in the message
// empty already holds when the extension is a message, and generating a scalar
// the ordinary way otherwise.
func extensionItem(empty protoreflect.Value, xd protoreflect.ExtensionDescriptor, opts GenOptions) (protoreflect.Value, bool) {
	switch xd.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		nopts := opts.nested()
		if nopts.MaxDepth <= 0 {
			return protoreflect.Value{}, false
		}
		if err := setDataOnMessage(empty.Message().Interface(), nopts); err != nil {
			return protoreflect.Value{}, false
		}
		return empty, true
	default:
		val := FieldValue(xd, opts)
		if val == nil {
			return protoreflect.Value{}, false
		}
		return *val, true
	}
}

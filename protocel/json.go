package protocel

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func UnmarshalDynamicMessageJSON(md protoreflect.MessageDescriptor, data []byte) (*celMessage, error) {
	m := map[string]any{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	fields, err := jsonToMessage(md, m)
	if err != nil {
		return nil, err
	}

	return NewCELMessage(md, fields)
}

func jsonToMessage(md protoreflect.MessageDescriptor, m map[string]any) (nodeMessage, error) {
	slog.Debug("jsonToMessage", "path", md.FullName())
	pbfields := md.Fields()
	fields := make(map[string]Node, len(m))
	for k, ival := range m {
		field := getFieldFromName(pbfields, k)
		if field == nil {
			return nil, fmt.Errorf("%s field not found", k)
		}
		val, err := jsonToNode(field, ival)
		if err != nil {
			return nil, err
		}
		fields[k] = val
	}
	return Message(fields), nil
}

func jsonToRepeated(fd protoreflect.FieldDescriptor, m []any) (nodeRepeated, error) {
	slog.Debug("jsonToRepeated", "path", fd.FullName())
	repeated := make([]Node, len(m))
	for i, ival := range m {
		val, err := jsonToNode(fd, ival)
		if err != nil {
			return nil, err
		}
		repeated[i] = val
	}
	return Repeated(repeated), nil
}

func jsonToMap(fd protoreflect.FieldDescriptor, m map[any]any) (nodeMap, error) {
	slog.Debug("jsonToMap", "path", fd.FullName())
	fields := make(map[Node]Node, len(m))
	for k, v := range m {
		key, err := jsonToNode(fd.MapKey(), k)
		if err != nil {
			return nil, err
		}
		val, err := jsonToNode(fd.MapValue(), v)
		if err != nil {
			return nil, err
		}
		fields[key] = val
	}
	return Map(fields), nil
}

func jsonToStringMap(fd protoreflect.FieldDescriptor, m map[string]any) (nodeMap, error) {
	slog.Debug("jsonToStringMap", "path", fd.FullName())
	fields := make(map[Node]Node, len(m))
	for k, v := range m {
		key, err := jsonToNode(fd.MapKey(), k)
		if err != nil {
			return nil, err
		}
		val, err := jsonToNode(fd.MapValue(), v)
		if err != nil {
			return nil, err
		}
		fields[key] = val
	}
	return Map(fields), nil
}

func jsonToNode(fd protoreflect.FieldDescriptor, ival any) (Node, error) {
	slog.Debug("jsonToNode", "path", fd.FullName())
	switch tt := ival.(type) {
	case string:
		return CEL(tt), nil
	case map[string]any:
		if fd.IsMap() {
			return jsonToStringMap(fd, tt)
		}
		switch fd.Kind() {
		case protoreflect.MessageKind:
			msg := fd.Message()
			if msg == nil {
				return nil, fmt.Errorf("%s: wrong type, expected a message but was: %s", fd.FullName(), fd.Kind())
			}
			return jsonToMessage(msg, tt)
		default:
			return nil, fmt.Errorf("%s: wrong type, expected a map or message but was: %s", fd.FullName(), fd.Kind())
		}
	case map[any]any:
		if !fd.IsMap() {
			return nil, fmt.Errorf("%s: wrong type, expected a map but was: %s", fd.FullName(), fd.Kind())
		}
		return jsonToMap(fd, tt)
	case []any:
		if !fd.IsList() {
			return nil, fmt.Errorf("%s: wrong type, expected a list but was: %s", fd.FullName(), fd.Kind())
		}
		return jsonToRepeated(fd, tt)
	default:
		return nil, fmt.Errorf("unhandled type: %T", ival)
	}
}

package idobfuscation

import (
	"context"
	"fmt"
	"time"

	"cbt-test-mini-project/util/idcodec"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func ResponseRewriter() func(ctx context.Context, response proto.Message) (any, error) {
	return func(ctx context.Context, response proto.Message) (any, error) {
		if response == nil {
			return response, nil
		}
		result, err := rewriteMessage(response.ProtoReflect())
		if err != nil {
			return nil, fmt.Errorf("idobfuscation: response rewrite failed: %w", err)
		}
		return result, nil
	}
}

func rewriteMessage(msg protoreflect.Message) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	fields := msg.Descriptor().Fields()

	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		jsonName := string(fd.Name())

		if !msg.Has(fd) {
			if fd.IsList() {
				result[jsonName] = []interface{}{}
			}
			continue
		}

		val := msg.Get(fd)
		encoded, err := rewriteValue(fd, val)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", jsonName, err)
		}
		result[jsonName] = encoded
	}

	return result, nil
}

func rewriteValue(fd protoreflect.FieldDescriptor, val protoreflect.Value) (interface{}, error) {
	fieldName := string(fd.Name())

	if fd.IsList() {
		return rewriteList(fd, val.List())
	}

	if fd.IsMap() {
		return rewriteMap(fd, val.Map())
	}

	if fd.Kind() == protoreflect.MessageKind || fd.Kind() == protoreflect.GroupKind {
		if isTimestampField(fd) {
			return formatProtoTimestamp(val.Message()), nil
		}
		return rewriteMessage(val.Message())
	}

	if isInt64Kind(fd.Kind()) && idcodec.IsIDField(fieldName) {
		id := val.Int()
		if id == 0 {
			return "", nil
		}
		encoded, err := idcodec.Encode(id)
		if err != nil {
			return nil, fmt.Errorf("encode id %d: %w", id, err)
		}
		return encoded, nil
	}

	if fd.Kind() == protoreflect.EnumKind {
		enumVal := fd.Enum().Values().ByNumber(val.Enum())
		if enumVal != nil {
			return string(enumVal.Name()), nil
		}
		return int32(val.Enum()), nil
	}

	return scalarToInterface(fd.Kind(), val), nil
}

func rewriteList(fd protoreflect.FieldDescriptor, list protoreflect.List) ([]interface{}, error) {
	result := make([]interface{}, 0, list.Len())
	for i := 0; i < list.Len(); i++ {
		val := list.Get(i)

		if fd.Kind() == protoreflect.MessageKind || fd.Kind() == protoreflect.GroupKind {
			if isTimestampField(fd) {
				result = append(result, formatProtoTimestamp(val.Message()))
				continue
			}
			encoded, err := rewriteMessage(val.Message())
			if err != nil {
				return nil, err
			}
			result = append(result, encoded)
			continue
		}

		fieldName := string(fd.Name())
		if isInt64Kind(fd.Kind()) && idcodec.IsIDField(fieldName) {
			id := val.Int()
			if id == 0 {
				result = append(result, "")
				continue
			}
			encoded, err := idcodec.Encode(id)
			if err != nil {
				return nil, err
			}
			result = append(result, encoded)
			continue
		}

		result = append(result, scalarToInterface(fd.Kind(), val))
	}
	return result, nil
}

func rewriteMap(fd protoreflect.FieldDescriptor, m protoreflect.Map) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	var rangeErr error
	m.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		key := fmt.Sprintf("%v", k.Value().Interface())
		valueFd := fd.MapValue()

		if valueFd.Kind() == protoreflect.MessageKind {
			encoded, err := rewriteMessage(v.Message())
			if err != nil {
				rangeErr = err
				return false
			}
			result[key] = encoded
		} else {
			result[key] = scalarToInterface(valueFd.Kind(), v)
		}
		return true
	})
	return result, rangeErr
}

func scalarToInterface(kind protoreflect.Kind, val protoreflect.Value) interface{} {
	switch kind {
	case protoreflect.BoolKind:
		return val.Bool()
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return int32(val.Int())
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return val.Int()
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return uint32(val.Uint())
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return val.Uint()
	case protoreflect.FloatKind:
		return float32(val.Float())
	case protoreflect.DoubleKind:
		return val.Float()
	case protoreflect.StringKind:
		return val.String()
	case protoreflect.BytesKind:
		return val.Bytes()
	default:
		return val.Interface()
	}
}

func isInt64Kind(kind protoreflect.Kind) bool {
	switch kind {
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return true
	default:
		return false
	}
}

func isTimestampField(fd protoreflect.FieldDescriptor) bool {
	if fd == nil || fd.Message() == nil {
		return false
	}
	return string(fd.Message().FullName()) == "google.protobuf.Timestamp"
}

func formatProtoTimestamp(msg protoreflect.Message) string {
	if !msg.IsValid() {
		return ""
	}

	secondsField := msg.Descriptor().Fields().ByName("seconds")
	nanosField := msg.Descriptor().Fields().ByName("nanos")
	if secondsField == nil || nanosField == nil {
		return ""
	}

	seconds := msg.Get(secondsField).Int()
	nanos := msg.Get(nanosField).Int()

	return time.Unix(seconds, nanos).UTC().Format(time.RFC3339Nano)
}

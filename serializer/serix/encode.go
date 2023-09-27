package serix

import (
	"bytes"
	"context"
	"math/big"
	"reflect"
	"time"

	"github.com/pkg/errors"

	"github.com/izuc/zipp.foundation/serializer"
	"github.com/izuc/zipp.foundation/serializer/byteutils"
)

func (api *API) encode(ctx context.Context, value reflect.Value, ts TypeSettings, opts *options) (b []byte, err error) {
	valueI := value.Interface()
	valueType := value.Type()
	if opts.validation {
		if err = api.callSyntacticValidator(ctx, value, valueType); err != nil {
			return nil, errors.Wrap(err, "pre-serialization validation failed")
		}
	}

	if serializable, ok := valueI.(Serializable); ok {
		typeSettingValue := value
		if valueType.Kind() == reflect.Interface {
			typeSettingValue = value.Elem()
		}
		globalTS, _ := api.getTypeSettings(typeSettingValue.Type())
		ts = ts.merge(globalTS)

		var bPrefix, bEncoded []byte
		if objectType := ts.ObjectType(); objectType != nil {
			s := serializer.NewSerializer()
			s.WriteNum(objectType, func(err error) error {
				return errors.Wrap(err, "failed to write object type code into serializer")
			})
			bPrefix, err = s.Serialize()
			if err != nil {
				return nil, errors.WithStack(err)
			}
		}

		bEncoded, err = serializable.Encode()
		if err != nil {
			return nil, errors.Wrap(err, "object failed to serialize itself")
		}
		b = byteutils.ConcatBytes(bPrefix, bEncoded)
	} else {
		b, err = api.encodeBasedOnType(ctx, value, valueI, valueType, ts, opts)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	if opts.validation {
		if err = api.callBytesValidator(ctx, valueType, b); err != nil {
			return nil, errors.Wrap(err, "post-serialization validation failed")
		}
	}

	return b, nil
}

func (api *API) encodeBasedOnType(
	ctx context.Context, value reflect.Value, valueI interface{}, valueType reflect.Type, ts TypeSettings, opts *options,
) ([]byte, error) {
	globalTS, _ := api.getTypeSettings(valueType)
	ts = ts.merge(globalTS)

	if opts.validation {
		if err := api.checkMinMaxBounds(value, ts); err != nil {
			return nil, err
		}
	}

	switch value.Kind() {
	case reflect.Ptr:
		if valueBigInt, ok := valueI.(*big.Int); ok {
			seri := serializer.NewSerializer()

			return seri.WriteUint256(valueBigInt, func(err error) error {
				return errors.Wrap(err, "failed to write math big int to serializer")
			}).Serialize()
		}
		elemValue := reflect.Indirect(value)
		if !elemValue.IsValid() {
			return nil, errors.Errorf("unexpected nil pointer for type %T", valueI)
		}
		if elemValue.Kind() == reflect.Struct {
			return api.encodeStruct(ctx, elemValue, elemValue.Interface(), elemValue.Type(), ts, opts)
		}

	case reflect.Struct:
		return api.encodeStruct(ctx, value, valueI, valueType, ts, opts)
	case reflect.Slice:
		return api.encodeSlice(ctx, value, valueType, ts, opts)
	case reflect.Map:
		return api.encodeMap(ctx, value, valueType, ts, opts)
	case reflect.Array:
		sliceValue := sliceFromArray(value)
		sliceValueType := sliceValue.Type()
		if sliceValueType.AssignableTo(bytesType) {
			seri := serializer.NewSerializer()

			return seri.WriteBytes(sliceValue.Bytes(), func(err error) error {
				return errors.Wrap(err, "failed to write array of bytes to serializer")
			}).Serialize()
		}

		return api.encodeSlice(ctx, sliceValue, sliceValueType, ts, opts)
	case reflect.Interface:
		return api.encodeInterface(ctx, value, valueType, ts, opts)
	case reflect.String:
		lengthPrefixType, set := ts.LengthPrefixType()
		if !set {
			return nil, errors.Errorf("can't serialize 'string' type: no LengthPrefixType was provided")
		}
		minLen, maxLen := ts.MinMaxLen()
		seri := serializer.NewSerializer()

		return seri.WriteString(
			value.String(),
			serializer.SeriLengthPrefixType(lengthPrefixType),
			func(err error) error {
				return errors.Wrap(err, "failed to write string value to serializer")
			}, minLen, maxLen).Serialize()

	case reflect.Bool:
		seri := serializer.NewSerializer()

		return seri.WriteBool(value.Bool(), func(err error) error {
			return errors.Wrap(err, "failed to write bool value to serializer")
		}).Serialize()

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		_, typeToConvert, _ := getNumberTypeToConvert(valueType.Kind())
		value = value.Convert(typeToConvert)
		valueI = value.Interface()
		seri := serializer.NewSerializer()

		return seri.WriteNum(valueI, func(err error) error {
			return errors.Wrap(err, "failed to write number value to serializer")
		}).Serialize()
	default:
	}

	return nil, errors.Errorf("can't encode: unsupported type %T", valueI)
}

// checks whether the given value is within its defined bounds in case it has a length.
func (api *API) checkMinMaxBounds(v reflect.Value, ts TypeSettings) error {
	if has := hasLength(v); !has {
		return nil
	}

	l := uint(v.Len())
	if minLen, ok := ts.MinLen(); ok {
		if l < minLen {
			return errors.Wrapf(serializer.ErrArrayValidationMinElementsNotReached, "can't serialize '%s' type: min length %d not reached (len %d)", v.Kind(), minLen, l)
		}
	}
	if maxLen, ok := ts.MaxLen(); ok {
		if l > maxLen {
			return errors.Wrapf(serializer.ErrArrayValidationMaxElementsExceeded, "can't serialize '%s' type: max length %d exceeded (len %d)", v.Kind(), maxLen, l)
		}
	}

	return nil
}

// checks whether the given value has the concept of a length.
func hasLength(v reflect.Value) bool {
	k := v.Kind()
	switch k {
	case reflect.Array:
	case reflect.Map:
	case reflect.Slice:
	case reflect.String:
	default:
		return false
	}

	return true
}

func (api *API) encodeInterface(
	ctx context.Context, value reflect.Value, valueType reflect.Type, ts TypeSettings, opts *options,
) ([]byte, error) {
	elemValue := value.Elem()
	if !elemValue.IsValid() {
		return nil, errors.Errorf("can't serialize interface %s it must have underlying value", valueType)
	}
	registry := api.getInterfaceObjects(valueType)
	if registry == nil {
		return nil, errors.Errorf("interface %s isn't registered", valueType)
	}
	elemType := elemValue.Type()
	if _, exists := registry.fromTypeToCode[elemType]; !exists {
		return nil, errors.Errorf("underlying type %s hasn't been registered for interface type %s",
			elemType, valueType)
	}
	encodedBytes, err := api.encode(ctx, elemValue, ts, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to encode interface element %s", elemType)
	}

	return encodedBytes, nil
}

func (api *API) encodeStruct(
	ctx context.Context, value reflect.Value, valueI interface{}, valueType reflect.Type, ts TypeSettings, opts *options,
) ([]byte, error) {
	if valueTime, ok := valueI.(time.Time); ok {
		seri := serializer.NewSerializer()

		return seri.WriteTime(valueTime, func(err error) error {
			return errors.Wrap(err, "failed to write time to serializer")
		}).Serialize()
	}
	s := serializer.NewSerializer()
	if objectType := ts.ObjectType(); objectType != nil {
		s.WriteNum(objectType, func(err error) error {
			return errors.Wrap(err, "failed to write object type code into serializer")
		})
	}
	if err := api.encodeStructFields(ctx, s, value, valueType, opts); err != nil {
		return nil, errors.WithStack(err)
	}

	return s.Serialize()
}

func (api *API) encodeStructFields(
	ctx context.Context, s *serializer.Serializer, value reflect.Value, valueType reflect.Type, opts *options,
) error {
	structFields, err := api.parseStructType(valueType)
	if err != nil {
		return errors.Wrapf(err, "can't parse struct type %s", valueType)
	}
	if len(structFields) == 0 {
		return nil
	}

	for _, sField := range structFields {
		fieldValue := value.Field(sField.index)
		if sField.isEmbeddedStruct && !sField.settings.nest {
			fieldType := sField.fType
			if fieldValue.Kind() == reflect.Ptr {
				if fieldValue.IsNil() {
					continue
				}
				fieldValue = fieldValue.Elem()
				fieldType = fieldType.Elem()
			}
			if err := api.encodeStructFields(ctx, s, fieldValue, fieldType, opts); err != nil {
				return errors.Wrapf(err, "can't serialize embedded struct %s", sField.name)
			}

			continue
		}
		var fieldBytes []byte
		if sField.settings.isOptional {
			if fieldValue.IsNil() {
				s.WritePayloadLength(0, func(err error) error {
					return errors.Wrapf(err,
						"failed to write zero length for an optional struct field %s to serializer",
						//nolint:scopelint // false positive
						sField.name,
					)
				})

				continue
			}
			fieldBytes, err = api.encode(ctx, fieldValue, sField.settings.ts, opts)
			if err != nil {
				return errors.Wrapf(err, "failed to serialize optional struct field %s", sField.name)
			}
			s.WritePayloadLength(len(fieldBytes), func(err error) error {
				return errors.Wrapf(err,
					"failed to write length for an optional struct field %s to serializer",
					//nolint:scopelint // false positive
					sField.name,
				)
			})
		} else {
			b, err := api.encode(ctx, fieldValue, sField.settings.ts, opts)
			if err != nil {
				return errors.Wrapf(err, "failed to serialize struct field %s", sField.name)
			}
			fieldBytes = b
		}
		s.WriteBytes(fieldBytes, func(err error) error {
			return errors.Wrapf(err,
				"failed to write serialized struct field bytes to serializer, field=%s",
				//nolint:scopelint // false positive
				sField.name,
			)
		})
	}

	return nil
}

func (api *API) encodeSlice(ctx context.Context, value reflect.Value, valueType reflect.Type,
	ts TypeSettings, opts *options) ([]byte, error) {

	if valueType.AssignableTo(bytesType) {
		lengthPrefixType, set := ts.LengthPrefixType()
		if !set {
			return nil, errors.Errorf("no LengthPrefixType was provided for slice type %s", valueType)
		}
		minLen, maxLen := ts.MinMaxLen()

		seri := serializer.NewSerializer()
		seri.WriteVariableByteSlice(value.Bytes(),
			serializer.SeriLengthPrefixType(lengthPrefixType),
			func(err error) error {
				return errors.Wrap(err, "failed to write bytes to serializer")
			}, minLen, maxLen)

		return seri.Serialize()
	}
	sliceLen := value.Len()
	data := make([][]byte, sliceLen)
	for i := 0; i < sliceLen; i++ {
		elemValue := value.Index(i)
		elemBytes, err := api.encode(ctx, elemValue, TypeSettings{}, opts)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to encode element with index %d of slice %s", i, valueType)
		}
		data[i] = elemBytes
	}

	return encodeSliceOfBytes(data, valueType, ts, opts)
}

func (api *API) encodeMap(ctx context.Context, value reflect.Value, valueType reflect.Type,
	ts TypeSettings, opts *options) ([]byte, error) {
	size := value.Len()
	data := make([][]byte, size)
	iter := value.MapRange()
	for i := 0; iter.Next(); i++ {
		key := iter.Key()
		elem := iter.Value()
		b, err := api.encodeMapKVPair(ctx, key, elem, opts)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		data[i] = b
	}
	ts = ts.ensureOrdering()

	return encodeSliceOfBytes(data, valueType, ts, opts)
}

func (api *API) encodeMapKVPair(ctx context.Context, key, val reflect.Value, opts *options) ([]byte, error) {
	keyBytes, err := api.encode(ctx, key, TypeSettings{}, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to encode map key of type %s", key.Type())
	}
	elemBytes, err := api.encode(ctx, val, TypeSettings{}, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to encode map element of type %s", val.Type())
	}
	buf := bytes.NewBuffer(keyBytes)
	buf.Write(elemBytes)

	return buf.Bytes(), nil
}

func encodeSliceOfBytes(data [][]byte, valueType reflect.Type, ts TypeSettings, opts *options) ([]byte, error) {
	lengthPrefixType, set := ts.LengthPrefixType()
	if !set {
		return nil, errors.Errorf("no LengthPrefixType was provided for type %s", valueType)
	}
	arrayRules := ts.ArrayRules()
	if arrayRules == nil {
		arrayRules = new(ArrayRules)
	}
	serializationMode := ts.toMode(opts)
	serializerArrayRules := serializer.ArrayRules(*arrayRules)
	serializerArrayRulesPtr := &serializerArrayRules
	seri := serializer.NewSerializer()
	seri.WriteSliceOfByteSlices(data,
		serializationMode,
		serializer.SeriLengthPrefixType(lengthPrefixType),
		serializerArrayRulesPtr,
		func(err error) error {
			return errors.Wrapf(err,
				"serializer failed to write %s as slice of bytes", valueType,
			)
		})

	return seri.Serialize()
}

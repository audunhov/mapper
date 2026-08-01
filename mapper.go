package mapper

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ImportMap defines a runtime mapping from input source keys (e.g. CSV header names)
// to exact struct field names.
type ImportMap map[string]string

// ConvertOption allows customizing the behavior of Convert.
type ConvertOption func(*convertConfig)

type convertConfig struct {
	errorHandler func(err error, rowIndex int, row map[string]string) error
}

// WithErrorHandler provides a custom callback to handle conversion errors on a per-row basis.
// If the callback returns nil, the row is skipped and conversion continues with the next row.
// If the callback returns a non-nil error, conversion is aborted immediately and that error is returned.
func WithErrorHandler(handler func(err error, rowIndex int, row map[string]string) error) ConvertOption {
	return func(cfg *convertConfig) {
		cfg.errorHandler = handler
	}
}

// WithSkipOnError is a convenience option that tells Convert to silently skip any rows that fail conversion.
func WithSkipOnError() ConvertOption {
	return func(cfg *convertConfig) {
		cfg.errorHandler = func(err error, rowIndex int, row map[string]string) error {
			return nil
		}
	}
}

// Convert maps rows in Imported into a slice of type T (struct or pointer to struct)
// using the provided runtime ImportMap. If mapping is nil or empty, fields are mapped
// automatically using struct tags (`map`) and matching field names.
// It accepts optional ConvertOption parameters to customize error handling.
func Convert[T any](imported *Imported, mapping ImportMap, opts ...ConvertOption) ([]T, error) {
	var result []T

	var dummy T
	t := reflect.TypeOf(dummy)

	// Determine underlying struct type
	isPtr := t.Kind() == reflect.Ptr
	structType := t
	if isPtr {
		structType = t.Elem()
	}
	if structType.Kind() != reflect.Struct {
		return nil, errors.New("type parameter T must be a struct or a pointer to a struct")
	}

	// Gather source keys for auto-mapping
	var sourceKeys []string
	if len(imported.Headers) > 0 {
		sourceKeys = imported.Headers
	} else if len(imported.Rows) > 0 {
		for k := range imported.Rows[0] {
			sourceKeys = append(sourceKeys, k)
		}
	}

	resolvedMapping := resolveMapping(sourceKeys, structType, mapping)
	fieldConfigs := getFieldConfigs(structType)

	cfg := &convertConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	for i, row := range imported.Rows {
		newElem := reflect.New(structType).Elem()
		if err := mapRowToStruct(row, newElem, resolvedMapping, fieldConfigs); err != nil {
			// CSV line numbers are generally: header line (1) + data row index (1-based) + 1 = i + 2
			wrappedErr := fmt.Errorf("row %d (line %d): %w", i+1, i+2, err)
			if cfg.errorHandler != nil {
				if cbErr := cfg.errorHandler(wrappedErr, i, row); cbErr != nil {
					return nil, cbErr
				}
				continue
			}
			return nil, wrappedErr
		}

		var val any
		if isPtr {
			val = newElem.Addr().Interface()
		} else {
			val = newElem.Interface()
		}

		typedVal, ok := val.(T)
		if !ok {
			return nil, fmt.Errorf("failed to assert value of type %T to target type %T", val, dummy)
		}
		result = append(result, typedVal)
	}

	return result, nil
}

type fieldConfig struct {
	layout string
}

func getFieldConfigs(structType reflect.Type) map[string]fieldConfig {
	configs := make(map[string]fieldConfig)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.PkgPath != "" {
			continue
		}
		tag := field.Tag.Get("map")
		if tag == "-" || tag == "" {
			continue
		}
		parts := strings.Split(tag, ",")
		var layout string
		for _, opt := range parts[1:] {
			opt = strings.TrimSpace(opt)
			if strings.HasPrefix(opt, "layout:") {
				layout = strings.TrimPrefix(opt, "layout:")
			} else if strings.HasPrefix(opt, "layout=") {
				layout = strings.TrimPrefix(opt, "layout=")
			}
		}
		if layout != "" {
			configs[field.Name] = fieldConfig{layout: layout}
		}
	}
	return configs
}

func normalizeKey(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			b.WriteRune(r)
		} else if r >= 'A' && r <= 'Z' {
			b.WriteRune(r + ('a' - 'A'))
		} else if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func resolveMapping(sourceKeys []string, structType reflect.Type, mapping ImportMap) ImportMap {
	resolved := make(ImportMap)

	// Pre-process struct fields
	taggedFields := make(map[string]string)
	exactFields := make(map[string]string)
	normalizedFields := make(map[string]string)

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.PkgPath != "" { // unexported
			continue
		}

		tag := field.Tag.Get("map")
		if tag == "-" {
			continue
		}

		var headerName string
		if tag != "" {
			parts := strings.Split(tag, ",")
			headerName = strings.TrimSpace(parts[0])
		}

		if headerName != "" {
			taggedFields[headerName] = field.Name
		}
		exactFields[field.Name] = field.Name
		normalizedFields[normalizeKey(field.Name)] = field.Name
	}

	// If a mapping is explicitly provided, we start with it
	if len(mapping) > 0 {
		for k, v := range mapping {
			resolved[k] = v
		}
		return resolved
	}

	// Otherwise, we perform auto-inference based on sourceKeys
	for _, key := range sourceKeys {
		// 1. Try explicit tag match
		if fieldName, ok := taggedFields[key]; ok {
			resolved[key] = fieldName
			continue
		}

		// 2. Try exact field name match
		if fieldName, ok := exactFields[key]; ok {
			resolved[key] = fieldName
			continue
		}

		// 3. Try normalized match
		normKey := normalizeKey(key)
		if fieldName, ok := normalizedFields[normKey]; ok {
			resolved[key] = fieldName
			continue
		}
	}

	return resolved
}

func mapRowToStruct(row map[string]string, val reflect.Value, mapping ImportMap, configs map[string]fieldConfig) error {
	for srcKey, destField := range mapping {
		field := val.FieldByName(destField)
		if !field.IsValid() {
			return fmt.Errorf("field %s does not exist on target struct", destField)
		}
		if !field.CanSet() {
			return fmt.Errorf("field %s cannot be set", destField)
		}

		rowVal, ok := row[srcKey]
		if !ok {
			continue
		}

		// Handle pointer fields
		targetField := field
		if field.Kind() == reflect.Ptr {
			if rowVal == "" {
				continue
			}
			// Resolve pointer chain
			for targetField.Kind() == reflect.Ptr {
				if targetField.IsNil() {
					targetField.Set(reflect.New(targetField.Type().Elem()))
				}
				targetField = targetField.Elem()
			}
		}

		// Check for time.Time
		if targetField.Type() == reflect.TypeOf(time.Time{}) {
			layout := time.RFC3339
			if cfg, exists := configs[destField]; exists && cfg.layout != "" {
				layout = cfg.layout
			}
			tVal, err := time.Parse(layout, rowVal)
			if err != nil {
				// If parsing with RFC3339 failed and layout was default, try "2006-01-02" fallback
				if layout == time.RFC3339 {
					if tVal, errFallback := time.Parse("2006-01-02", rowVal); errFallback == nil {
						targetField.Set(reflect.ValueOf(tVal))
						continue
					}
				}
				return fmt.Errorf("cannot parse %q as time for field %s using layout %q: %w", rowVal, destField, layout, err)
			}
			targetField.Set(reflect.ValueOf(tVal))
			continue
		}

		// Support TextUnmarshaler, so other custom types can override default behavior
		if targetField.CanAddr() {
			if unmarshaler, ok := targetField.Addr().Interface().(encoding.TextUnmarshaler); ok {
				if err := unmarshaler.UnmarshalText([]byte(rowVal)); err != nil {
					return fmt.Errorf("cannot unmarshal %q into field %s: %w", rowVal, destField, err)
				}
				continue
			}
		}
		if unmarshaler, ok := targetField.Interface().(encoding.TextUnmarshaler); ok {
			if err := unmarshaler.UnmarshalText([]byte(rowVal)); err != nil {
				return fmt.Errorf("cannot unmarshal %q into field %s: %w", rowVal, destField, err)
			}
			continue
		}

		switch targetField.Kind() {
		case reflect.String:
			targetField.SetString(rowVal)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intVal, err := strconv.ParseInt(rowVal, 10, 64)
			if err != nil {
				return fmt.Errorf("cannot parse %q as int for field %s: %w", rowVal, destField, err)
			}
			targetField.SetInt(intVal)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintVal, err := strconv.ParseUint(rowVal, 10, 64)
			if err != nil {
				return fmt.Errorf("cannot parse %q as uint for field %s: %w", rowVal, destField, err)
			}
			targetField.SetUint(uintVal)
		case reflect.Float32, reflect.Float64:
			floatVal, err := strconv.ParseFloat(rowVal, 64)
			if err != nil {
				return fmt.Errorf("cannot parse %q as float for field %s: %w", rowVal, destField, err)
			}
			targetField.SetFloat(floatVal)
		case reflect.Bool:
			boolVal, err := strconv.ParseBool(rowVal)
			if err != nil {
				return fmt.Errorf("cannot parse %q as bool for field %s: %w", rowVal, destField, err)
			}
			targetField.SetBool(boolVal)
		default:
			return fmt.Errorf("unsupported field type %s for field %s", targetField.Kind(), destField)
		}
	}
	return nil
}


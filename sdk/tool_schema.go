package sdk

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
)

// resolveSchema converts a Tool's Parameters value into a *jsonschema.Schema.
//
// Accepted types:
//   - *jsonschema.Schema: returned as-is
//   - Go struct (or pointer to struct): schema is inferred via jsonschema.ForType
//   - nil: returns nil
func resolveSchema(v any) (*jsonschema.Schema, error) {
	if v == nil {
		return nil, nil
	}
	if s, ok := v.(*jsonschema.Schema); ok {
		return s, nil
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("twilightai: Parameters must be *jsonschema.Schema or a struct, got %T", v)
	}
	return jsonschema.ForType(t, nil)
}

// NewTool creates a Tool with the JSON Schema inferred from the type parameter T
// and a type-safe Execute handler. T must be a struct type with exported fields.
//
// The inferred schema uses json struct tags for property names, the jsonschema
// struct tag for descriptions, and omitempty/omitzero to determine optional fields.
//
// NewTool panics if schema inference fails (a programming error in T's definition).
func NewTool[T any](name, description string, execute func(ctx *ToolExecContext, input T) (any, error)) Tool {
	schema, err := jsonschema.For[T](nil)
	if err != nil {
		panic(fmt.Sprintf("twilightai: cannot infer schema for tool %q: %v", name, err))
	}
	return Tool{
		Name:        name,
		Description: description,
		Parameters:  schema,
		Execute: func(ctx *ToolExecContext, input any) (any, error) {
			data, err := json.Marshal(input)
			if err != nil {
				return nil, fmt.Errorf("twilightai: marshal tool input: %w", err)
			}
			var typed T
			if err := json.Unmarshal(data, &typed); err != nil {
				return nil, fmt.Errorf("twilightai: unmarshal tool input to %T: %w", typed, err)
			}
			return execute(ctx, typed)
		},
	}
}

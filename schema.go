package huma

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	timeType      = reflect.TypeOf(time.Time{})
	ipType        = reflect.TypeOf(net.IP{})
	uriType       = reflect.TypeOf(url.URL{})
	byteSliceType = reflect.TypeOf([]byte(nil))
)

// getTagValue returns a value of the schema's type for the given tag string.
// Uses JSON parsing if the schema is not a string.
func getTagValue(s *Schema, value string) (interface{}, error) {
	if s.Type == "string" {
		return value, nil
	}

	var v interface{}
	if err := json.Unmarshal([]byte(value), &v); err != nil {
		return nil, err
	}
	return v, nil
}

// Schema represents a JSON Schema which can be generated from Go structs
type Schema struct {
	Type             string             `json:"type,omitempty"`
	Description      string             `json:"description,omitempty"`
	Items            *Schema            `json:"items,omitempty"`
	Properties       map[string]*Schema `json:"properties,omitempty"`
	Required         []string           `json:"required,omitempty"`
	Format           string             `json:"format,omitempty"`
	Enum             []interface{}      `json:"enum,omitempty"`
	Default          interface{}        `json:"default,omitempty"`
	Example          interface{}        `json:"example,omitempty"`
	Minimum          *int               `json:"minimum,omitempty"`
	ExclusiveMinimum *int               `json:"exclusiveMinimum,omitempty"`
	Maximum          *int               `json:"maximum,omitempty"`
	ExclusiveMaximum *int               `json:"exclusiveMaximum,omitempty"`
	MultipleOf       int                `json:"multipleOf,omitempty"`
	MinLength        *int               `json:"minLength,omitempty"`
	MaxLength        *int               `json:"maxLength,omitempty"`
	Pattern          string             `json:"pattern,omitempty"`
	MinItems         *int               `json:"minItems,omitempty"`
	MaxItems         *int               `json:"maxItems,omitempty"`
	UniqueItems      bool               `json:"uniqueItems,omitempty"`
	MinProperties    *int               `json:"minProperties,omitempty"`
	MaxProperties    *int               `json:"maxProperties,omitempty"`
}

// GenerateSchema creates a JSON schema for a Go type. Struct field tags
// can be used to provide additional metadata such as descriptions and
// validation.
func GenerateSchema(t reflect.Type) (*Schema, error) {
	schema := &Schema{}

	if t == ipType {
		// Special case: IP address.
		return &Schema{Type: "string", Format: "ipv4"}, nil
	}

	switch t.Kind() {
	case reflect.Struct:
		// Handle special cases.
		switch t {
		case timeType:
			return &Schema{Type: "string", Format: "date-time"}, nil
		case uriType:
			return &Schema{Type: "string", Format: "uri"}, nil
		}

		properties := make(map[string]*Schema)
		required := make([]string, 0)
		schema.Type = "object"

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			jsonTags := strings.Split(f.Tag.Get("json"), ",")

			name := f.Name
			if len(jsonTags) > 0 {
				name = jsonTags[0]
			}

			s, err := GenerateSchema(f.Type)
			if err != nil {
				return nil, err
			}
			properties[name] = s

			if t, ok := f.Tag.Lookup("description"); ok {
				s.Description = t
			}

			if t, ok := f.Tag.Lookup("format"); ok {
				s.Format = t
			}

			if t, ok := f.Tag.Lookup("enum"); ok {
				s.Enum = []interface{}{}
				for _, v := range strings.Split(t, ",") {
					parsed, err := getTagValue(s, v)
					if err != nil {
						return nil, err
					}
					s.Enum = append(s.Enum, parsed)
				}
			}

			if t, ok := f.Tag.Lookup("default"); ok {
				v, err := getTagValue(s, t)
				if err != nil {
					return nil, err
				}

				s.Default = v
			}

			if t, ok := f.Tag.Lookup("example"); ok {
				v, err := getTagValue(s, t)
				if err != nil {
					return nil, err
				}

				s.Example = v
			}

			if t, ok := f.Tag.Lookup("minimum"); ok {
				min, err := strconv.Atoi(t)
				if err != nil {
					return nil, err
				}
				s.Minimum = &min
			}

			if t, ok := f.Tag.Lookup("exclusiveMinimum"); ok {
				min, err := strconv.Atoi(t)
				if err != nil {
					return nil, err
				}
				s.ExclusiveMinimum = &min
			}

			if t, ok := f.Tag.Lookup("maximum"); ok {
				max, err := strconv.Atoi(t)
				if err != nil {
					return nil, err
				}
				s.Maximum = &max
			}

			if t, ok := f.Tag.Lookup("exclusiveMaximum"); ok {
				max, err := strconv.Atoi(t)
				if err != nil {
					return nil, err
				}
				s.ExclusiveMaximum = &max
			}

			if t, ok := f.Tag.Lookup("multipleOf"); ok {
				mof, err := strconv.Atoi(t)
				if err != nil {
					return nil, err
				}
				s.MultipleOf = mof
			}

			if t, ok := f.Tag.Lookup("minLength"); ok {
				min, err := strconv.Atoi(t)
				if err != nil {
					return nil, err
				}
				s.MinLength = &min
			}

			if t, ok := f.Tag.Lookup("maxLength"); ok {
				max, err := strconv.Atoi(t)
				if err != nil {
					return nil, err
				}
				s.MaxLength = &max
			}

			if t, ok := f.Tag.Lookup("pattern"); ok {
				s.Pattern = t
			}

			if t, ok := f.Tag.Lookup("minItems"); ok {
				min, err := strconv.Atoi(t)
				if err != nil {
					return nil, err
				}
				s.MinItems = &min
			}

			if t, ok := f.Tag.Lookup("maxItems"); ok {
				max, err := strconv.Atoi(t)
				if err != nil {
					return nil, err
				}
				s.MaxItems = &max
			}

			if t, ok := f.Tag.Lookup("uniqueItems"); ok {
				s.UniqueItems = t == "true"
			}

			if t, ok := f.Tag.Lookup("minProperties"); ok {
				min, err := strconv.Atoi(t)
				if err != nil {
					return nil, err
				}
				s.MinProperties = &min
			}

			if t, ok := f.Tag.Lookup("maxProperties"); ok {
				max, err := strconv.Atoi(t)
				if err != nil {
					return nil, err
				}
				s.MaxProperties = &max
			}

			optional := false
			for _, tag := range jsonTags[1:] {
				if tag == "omitempty" {
					optional = true
				}
			}
			if !optional {
				required = append(required, name)
			}
		}

		if len(properties) > 0 {
			schema.Properties = properties
		}

		if len(required) > 0 {
			schema.Required = required
		}

	case reflect.Map:
		// pass
	case reflect.Slice, reflect.Array:
		schema.Type = "array"
		s, err := GenerateSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		schema.Items = s
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &Schema{
			Type: "integer",
		}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// Unsigned integers can't be negative.
		min := 0
		return &Schema{
			Type:    "integer",
			Minimum: &min,
		}, nil
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}, nil
	case reflect.Bool:
		return &Schema{Type: "boolean"}, nil
	case reflect.String:
		return &Schema{Type: "string"}, nil
	case reflect.Ptr:
		return GenerateSchema(t.Elem())
	default:
		return nil, fmt.Errorf("unsupported type %s from %s", t.Kind(), t)
	}

	return schema, nil
}

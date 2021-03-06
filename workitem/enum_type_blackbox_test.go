package workitem_test

import (
	"testing"

	"github.com/fabric8-services/fabric8-wit/convert"
	"github.com/fabric8-services/fabric8-wit/resource"
	w "github.com/fabric8-services/fabric8-wit/workitem"
	"github.com/stretchr/testify/require"
)

func TestEnumType_EqualAndEqualValue(t *testing.T) {
	t.Parallel()
	resource.Require(t, resource.UnitTest)

	a := w.EnumType{
		SimpleType:       w.SimpleType{Kind: w.KindEnum},
		BaseType:         w.SimpleType{Kind: w.KindString},
		Values:           []interface{}{"foo", "bar"},
		RewritableValues: false,
		DefaultValue:     "fooooooobar",
	}
	t.Run("type inequality", func(t *testing.T) {
		t.Parallel()
		require.False(t, a.Equal(convert.DummyEqualer{}))
		require.False(t, a.EqualValue(convert.DummyEqualer{}))
	})

	t.Run("simple type difference", func(t *testing.T) {
		t.Parallel()
		b := a
		b.SimpleType = w.SimpleType{Kind: w.KindArea}
		require.False(t, a.Equal(b))
		require.False(t, a.EqualValue(b))
	})

	t.Run("base type difference", func(t *testing.T) {
		t.Parallel()
		b := a
		b.BaseType = w.SimpleType{Kind: w.KindInteger}
		require.False(t, a.Equal(b))
		require.False(t, a.EqualValue(b))
	})

	t.Run("default value difference", func(t *testing.T) {
		t.Parallel()
		b := a
		b.DefaultValue = "foo"
		require.False(t, a.Equal(b))
		require.False(t, a.EqualValue(b))
	})

	t.Run("value difference", func(t *testing.T) {
		t.Parallel()
		t.Run("not equal", func(t *testing.T) {
			t.Parallel()
			b := a
			b.Values = []interface{}{"foo1", "bar2"}
			require.False(t, a.Equal(b))
			require.False(t, a.EqualValue(b))
		})

		t.Run("new type has overwritable values but old not", func(t *testing.T) {
			t.Parallel()
			b := a
			b.Values = []interface{}{"foo1", "bar2"}
			b.RewritableValues = true
			require.False(t, a.Equal(b))
			require.False(t, a.EqualValue(b))
		})

		t.Run("old type has overwritable values but new not", func(t *testing.T) {
			t.Parallel()
			b := a
			b.Values = []interface{}{"foo1", "bar2"}
			b.RewritableValues = true
			require.True(t, b.Equal(a))
			require.True(t, b.EqualValue(a))
		})
		t.Run("old and new type have overwritable values", func(t *testing.T) {
			t.Parallel()
			b := a
			b.RewritableValues = true
			b.Values = []interface{}{"foo1", "bar2"}
			c := a
			c.RewritableValues = true
			require.True(t, b.Equal(c))
			require.True(t, c.Equal(b))
			require.True(t, b.EqualValue(c))
			require.True(t, c.EqualValue(b))
		})
	})
}

func TestEnumType_GetDefaultValue(t *testing.T) {
	t.Parallel()
	resource.Require(t, resource.UnitTest)

	tests := []struct {
		name           string
		enum           w.EnumType
		expectedOutput interface{}
	}{
		{"return first value of enum when default is nil", w.EnumType{
			SimpleType: w.SimpleType{Kind: w.KindEnum},
			BaseType:   w.SimpleType{Kind: w.KindString},
			Values:     []interface{}{"first", "second", "third"},
		}, "first"},
		{"return custom default when a default is set", w.EnumType{
			SimpleType:   w.SimpleType{Kind: w.KindEnum},
			BaseType:     w.SimpleType{Kind: w.KindInteger},
			Values:       []interface{}{111, 222, 333},
			DefaultValue: 222,
		}, 222},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.expectedOutput, tt.enum.GetDefaultValue())
		})
	}
}

func TestEnumType_SetDefaultValue(t *testing.T) {
	t.Parallel()
	resource.Require(t, resource.UnitTest)

	tests := []struct {
		name           string
		enum           w.EnumType
		defVal         interface{}
		expectedOutput w.FieldType
		wantErr        bool
	}{
		{"set default to allowed value",
			w.EnumType{
				SimpleType: w.SimpleType{Kind: w.KindEnum},
				BaseType:   w.SimpleType{Kind: w.KindString},
				Values:     []interface{}{"first", "second", "third"},
			},
			"second",
			&w.EnumType{
				SimpleType:   w.SimpleType{Kind: w.KindEnum},
				BaseType:     w.SimpleType{Kind: w.KindString},
				Values:       []interface{}{"first", "second", "third"},
				DefaultValue: "second",
			},
			false},
		{"set default to nil value",
			w.EnumType{
				SimpleType: w.SimpleType{Kind: w.KindEnum},
				BaseType:   w.SimpleType{Kind: w.KindString},
				Values:     []interface{}{"first", "second", "third"},
			},
			nil,
			&w.EnumType{
				SimpleType:   w.SimpleType{Kind: w.KindEnum},
				BaseType:     w.SimpleType{Kind: w.KindString},
				Values:       []interface{}{"first", "second", "third"},
				DefaultValue: nil,
			},
			false},
		{"set default to not-allowed value (wrong base type)",
			w.EnumType{
				SimpleType: w.SimpleType{Kind: w.KindEnum},
				BaseType:   w.SimpleType{Kind: w.KindString},
				Values:     []interface{}{"first", "second", "third"},
			},
			123,
			nil,
			true},
		{"set default to not-allowed value (not in list)",
			w.EnumType{
				SimpleType: w.SimpleType{Kind: w.KindEnum},
				BaseType:   w.SimpleType{Kind: w.KindString},
				Values:     []interface{}{"first", "second", "third"},
			},
			"foobar",
			nil,
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output, err := tt.enum.SetDefaultValue(tt.defVal)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func TestEnumType_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		obj     w.EnumType
		wantErr bool
	}{
		{"ok", w.EnumType{
			SimpleType:       w.SimpleType{Kind: w.KindEnum},
			BaseType:         w.SimpleType{Kind: w.KindString},
			Values:           []interface{}{"who", "let", "the", "dogs", "out"},
			RewritableValues: false,
			DefaultValue:     "the",
		}, false},
		{"error - empty values", w.EnumType{
			SimpleType:       w.SimpleType{Kind: w.KindEnum},
			BaseType:         w.SimpleType{Kind: w.KindString},
			Values:           []interface{}{},
			RewritableValues: false,
			DefaultValue:     "the",
		}, true},
		{"error - nil values", w.EnumType{
			SimpleType:       w.SimpleType{Kind: w.KindEnum},
			BaseType:         w.SimpleType{Kind: w.KindString},
			Values:           nil,
			RewritableValues: false,
			DefaultValue:     "the",
		}, true},
		{"invalid type", w.EnumType{
			SimpleType:       w.SimpleType{Kind: w.KindString},
			BaseType:         w.SimpleType{Kind: w.KindString},
			Values:           []interface{}{"who", "let", "the", "dogs", "out"},
			RewritableValues: false,
			DefaultValue:     "the",
		}, true},
		{"invalid base type (list)", w.EnumType{
			SimpleType:       w.SimpleType{Kind: w.KindEnum},
			BaseType:         w.SimpleType{Kind: w.KindList},
			Values:           []interface{}{"who", "let", "the", "dogs", "out"},
			RewritableValues: false,
			DefaultValue:     "the",
		}, true},
		{"invalid base type (enum)", w.EnumType{
			SimpleType:       w.SimpleType{Kind: w.KindEnum},
			BaseType:         w.SimpleType{Kind: w.KindEnum},
			Values:           []interface{}{"who", "let", "the", "dogs", "out"},
			RewritableValues: false,
			DefaultValue:     "the",
		}, true},
		{"invalid string values", w.EnumType{
			SimpleType:       w.SimpleType{Kind: w.KindEnum},
			BaseType:         w.SimpleType{Kind: w.KindString},
			Values:           []interface{}{"who", 1, "the", "dogs", "out"},
			RewritableValues: false,
			DefaultValue:     "the",
		}, true},
		{"invalid integer values", w.EnumType{
			SimpleType:       w.SimpleType{Kind: w.KindEnum},
			BaseType:         w.SimpleType{Kind: w.KindInteger},
			Values:           []interface{}{1, 2, "the", 4, 5},
			RewritableValues: false,
			DefaultValue:     "the",
		}, true},
		{"invalid default value (wrong type)", w.EnumType{
			SimpleType:       w.SimpleType{Kind: w.KindEnum},
			BaseType:         w.SimpleType{Kind: w.KindInteger},
			Values:           []interface{}{1, 2, 3, 4, 5},
			RewritableValues: false,
			DefaultValue:     "the",
		}, true},
		{"invalid default value (not in allowed values)", w.EnumType{
			SimpleType:       w.SimpleType{Kind: w.KindEnum},
			BaseType:         w.SimpleType{Kind: w.KindInteger},
			Values:           []interface{}{1, 2, 3, 4, 5},
			RewritableValues: false,
			DefaultValue:     42,
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.obj.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEnumType_EqualEnclosing(t *testing.T) {
	t.Parallel()
	resource.Require(t, resource.UnitTest)

	a := w.EnumType{
		SimpleType:       w.SimpleType{Kind: w.KindEnum},
		BaseType:         w.SimpleType{Kind: w.KindString},
		Values:           []interface{}{"foo", "bar", "baz"},
		RewritableValues: false,
	}

	t.Run("simple type difference", func(t *testing.T) {
		b := a
		b.SimpleType = w.SimpleType{Kind: w.KindArea}
		require.False(t, a.EqualEnclosing(b))
	})

	t.Run("base type difference", func(t *testing.T) {
		b := a
		b.BaseType = w.SimpleType{Kind: w.KindInteger}
		require.False(t, a.EqualEnclosing(b))
	})

	t.Run("value difference", func(t *testing.T) {
		t.Run("not equal", func(t *testing.T) {
			b := a
			b.Values = []interface{}{"foo1", "bar2"}
			require.False(t, a.EqualEnclosing(b))
		})

		t.Run("new type has subset values", func(t *testing.T) {
			b := a
			b.Values = []interface{}{"foo", "bar"}
			require.False(t, b.EqualEnclosing(a))
		})

		t.Run("new type has more than subset values but not all of old set", func(t *testing.T) {
			b := a
			b.Values = []interface{}{"foo", "bar", "hello"}
			require.False(t, b.EqualEnclosing(a))
		})

		t.Run("new type has more than subset values", func(t *testing.T) {
			b := a
			b.Values = []interface{}{"foo", "bar", "baz", "hello"}
			require.True(t, b.EqualEnclosing(a))
		})

		t.Run("new type has empty values", func(t *testing.T) {
			b := a
			b.Values = []interface{}{}
			require.False(t, b.EqualEnclosing(a))
		})
	})
}

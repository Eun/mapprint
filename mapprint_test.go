package mapprint

import (
	"io"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTypes(t *testing.T) {
	require.Equal(t, "Hello World!", Sprintf("Hello %Planet!", map[string]string{"Planet": "World"}))

	require.Equal(t, "1 + 2 = 3", Sprintf("%Number1 + %Number2 = %Number3", map[string]interface{}{
		"Number1": 1,
		"Number2": 2,
		"Number3": 3,
	}))

	require.Equal(t, "[Earth, Kepler-107, Starkiller Base]", Sprintf("%Planets", map[string]interface{}{
		"Planets": []string{"Earth", "Kepler-107", "Starkiller Base"},
	}))

	require.Equal(t, "[Earth, Kepler-107, Starkiller Base]", Sprintf("%Planets", map[string]interface{}{
		"Planets": []string{"Earth", "Kepler-107", "Starkiller Base"},
	}))

	// Only number keys are not allowed
	require.Equal(t, "%1", Sprintf("%1", map[string]interface{}{
		"1": "Hello",
	}))

	// functions as value
	require.Equal(t, "3", Sprintf("%value", map[string]interface{}{
		"value": func() interface{} {
			return 3
		},
	}))
	require.Equal(t, "Hello World", Sprintf("%value", map[string]interface{}{
		"value": func() string {
			return "Hello World"
		},
	}))
	require.Equal(t, "3", Sprintf("%value", map[string]interface{}{
		"value": func() int {
			return 3
		},
	}))
	require.Equal(t, "[Hello, 1]", Sprintf("%value", map[string]interface{}{
		"value": func() interface{} {
			return []interface{}{"Hello", 1}
		},
	}))
	require.Equal(t, "[Hello, 1]", Sprintf("%value", map[string]interface{}{
		"value": func() (string, int) {
			return "Hello", 1
		},
	}))
	require.Equal(t, "Hello", Sprintf("%.0value", map[string]interface{}{
		"value": func() (string, int) {
			return "Hello", 1
		},
	}))
	require.Equal(t, "", Sprintf("%value", map[string]interface{}{
		"value": func() {},
	}))

	// integers
	require.Equal(t, "6", Sprintf("%value", map[string]interface{}{"value": int(6)}))
	require.Equal(t, "6", Sprintf("%value", map[string]interface{}{"value": uint(6)}))

	// floats
	require.Equal(t, "6.200000", Sprintf("%value", map[string]interface{}{"value": float32(6.2)}))
	require.Equal(t, "6.200000", Sprintf("%value", map[string]interface{}{"value": float64(6.2)}))
	require.Equal(t, "6.230000", Sprintf("%value", map[string]interface{}{"value": float32(6.23)}))
	require.Equal(t, "6.230000", Sprintf("%value", map[string]interface{}{"value": float64(6.23)}))

	// bool
	require.Equal(t, "true", Sprintf("%value", map[string]interface{}{"value": true}))
	require.Equal(t, "false", Sprintf("%value", map[string]interface{}{"value": false}))

	// rune
	require.Equal(t, "65", Sprintf("%value", map[string]interface{}{"value": 'A'}))

	// ptr
	s := "Hello"
	require.Equal(t, "Hello", Sprintf("%value", map[string]interface{}{"value": &s}))

	// test unsupported
	require.Equal(t, "", Sprintf("%value", map[string]interface{}{
		"value": map[string]interface{}{},
	}))

	// test nil
	require.Equal(t, "", Sprintf("%value", map[string]interface{}{
		"value": nil,
	}))

	// test %%
	t.Run("%%", func(t *testing.T) {
		p := Printer{
			DefaultBindings: map[string]interface{}{
				"Percent": 100,
			},
			KeyNotFound: func(w io.Writer, printer *Printer, prefix, key []rune, defaultPrinter PrintValueFunc) (int, error) {
				panic("Should not have been called")
			},
		}
		require.Equal(t, "Foo % Bar", p.Sprintf("Foo % Bar"))
		require.Equal(t, "Foo % Bar", p.Sprintf("Foo %% Bar"))
		
		require.Equal(t, "Foo 100% Bar", p.Sprintf("Foo %Percent% Bar"))
		require.Equal(t, "Foo 100% Bar", p.Sprintf("Foo %Percent%% Bar"))
	
		require.Equal(t, "Foo %100% Bar", p.Sprintf("Foo %%%Percent% Bar"))
		require.Equal(t, "Foo %100% Bar", p.Sprintf("Foo %%%Percent%% Bar"))
	
		require.Equal(t, "Foo%100%Bar", p.Sprintf("Foo%%%Percent%%Bar"))
		require.Equal(t, "Foo %%%%%%%100 Bar", p.Sprintf("Foo %%10Percent Bar"))
	})
}

func TestInitializedPrinter(t *testing.T) {
	t.Run("No Settings", func(t *testing.T) {
		var p Printer
		require.Equal(t, "Hello Earth!", p.Sprintf("Hello %Planet!", map[string]interface{}{"Planet": "Earth"}))
	})
	t.Run("CustomPrint", func(t *testing.T) {
		customPrintCalled := false
		p := Printer{
			PrintValue: func(w io.Writer, printer *Printer, prefix []rune, key []rune, value reflect.Value) (int, error) {
				require.Equal(t, "8", string(prefix))
				require.Equal(t, "Planet", string(key))
				require.Equal(t, "Earth", defaultReflectPrinter.Sprint(value))
				customPrintCalled = true
				return 0, nil
			},
		}
		require.Equal(t, "Hello !", p.Sprintf("Hello %8Planet!", map[string]interface{}{"Planet": "Earth"}))
		require.True(t, customPrintCalled)
	})
}

func TestKeyNotFound(t *testing.T) {
	t.Run("DefaultValue", func(t *testing.T) {
		p := Printer{
			KeyNotFound: DefaultValue("Mars"),
		}
		require.Equal(t, "Mars", p.Sprintf("%Planet"))
		require.Equal(t, "Mars!", p.Sprintf("%Planet!"))
		require.Equal(t, "Foo Mars!", p.Sprintf("Foo %Planet!"))
		require.Equal(t, "Foo Mars", p.Sprintf("Foo %Planet"))
		require.Equal(t, "Mars! Bar", p.Sprintf("%Planet! Bar"))
		require.Equal(t, "Mars Bar", p.Sprintf("%Planet Bar"))
		require.Equal(t, "Foo Mars! Bar", p.Sprintf("Foo %Planet! Bar"))
		require.Equal(t, "Foo Mars Bar", p.Sprintf("Foo %Planet Bar"))
		require.Equal(t, "Foo       Mars Bar", p.Sprintf("Foo %10Planet Bar"))
	})
	t.Run("KeepKey", func(t *testing.T) {
		p := Printer{
			KeyNotFound: KeepKey(),
		}
		require.Equal(t, "%Planet", p.Sprintf("%Planet"))
		require.Equal(t, "%Planet!", p.Sprintf("%Planet!"))
		require.Equal(t, "Foo %Planet!", p.Sprintf("Foo %Planet!"))
		require.Equal(t, "Foo %Planet", p.Sprintf("Foo %Planet"))
		require.Equal(t, "%Planet! Bar", p.Sprintf("%Planet! Bar"))
		require.Equal(t, "%Planet Bar", p.Sprintf("%Planet Bar"))
		require.Equal(t, "Foo %Planet! Bar", p.Sprintf("Foo %Planet! Bar"))
		require.Equal(t, "Foo %Planet Bar", p.Sprintf("Foo %Planet Bar"))
		require.Equal(t, "Foo %10Planet Bar", p.Sprintf("Foo %10Planet Bar"))
	})
	t.Run("ClearKey", func(t *testing.T) {
		p := Printer{
			KeyNotFound: ClearKey(),
		}
		require.Equal(t, "", p.Sprintf("%Planet"))
		require.Equal(t, "!", p.Sprintf("%Planet!"))
		require.Equal(t, "Foo !", p.Sprintf("Foo %Planet!"))
		require.Equal(t, "Foo ", p.Sprintf("Foo %Planet"))
		require.Equal(t, "! Bar", p.Sprintf("%Planet! Bar"))
		require.Equal(t, " Bar", p.Sprintf("%Planet Bar"))
		require.Equal(t, "Foo ! Bar", p.Sprintf("Foo %Planet! Bar"))
		require.Equal(t, "Foo  Bar", p.Sprintf("Foo %Planet Bar"))
		require.Equal(t, "Foo  Bar", p.Sprintf("Foo %10Planet Bar"))
	})
	t.Run("Custom", func(t *testing.T) {
		calledCustomFunc := false
		p := Printer{
			KeyToken: '%',
			KeyNotFound: func(w io.Writer, printer *Printer, prefix, key []rune, defaultPrinter PrintValueFunc) (int, error) {
				s := string(key)
				require.Equal(t, s, "Planet")
				calledCustomFunc = true
				return defaultPrinter(w, printer, prefix, key, reflect.ValueOf("Mars"))
			},
		}
		require.Equal(t, "Mars", p.Sprintf("%Planet"))
		require.Equal(t, "Mars!", p.Sprintf("%Planet!"))
		require.Equal(t, "Foo Mars!", p.Sprintf("Foo %Planet!"))
		require.Equal(t, "Foo Mars", p.Sprintf("Foo %Planet"))
		require.Equal(t, "Mars! Bar", p.Sprintf("%Planet! Bar"))
		require.Equal(t, "Mars Bar", p.Sprintf("%Planet Bar"))
		require.Equal(t, "Foo Mars! Bar", p.Sprintf("Foo %Planet! Bar"))
		require.Equal(t, "Foo Mars Bar", p.Sprintf("Foo %Planet Bar"))
		require.Equal(t, "Foo       Mars Bar", p.Sprintf("Foo %10Planet Bar"))
		require.True(t, calledCustomFunc)
	})
	t.Run("CustomWithError", func(t *testing.T) {
		t.Run("Suppressed", func(t *testing.T) {
			calledCustomFunc := false
			p := Printer{
				KeyToken:       '%',
				SuppressErrors: true,
				KeyNotFound: func(w io.Writer, printer *Printer, prefix, key []rune, defaultPrinter PrintValueFunc) (int, error) {
					calledCustomFunc = true
					return 0, internalError{}
				},
			}
			require.Equal(t, "", p.Sprintf("%Planet"))
			require.True(t, calledCustomFunc)
		})
		t.Run("Not Suppressed", func(t *testing.T) {
			calledCustomFunc := false
			p := Printer{
				KeyToken:       '%',
				SuppressErrors: false,
				KeyNotFound: func(w io.Writer, printer *Printer, prefix, key []rune, defaultPrinter PrintValueFunc) (int, error) {
					calledCustomFunc = true
					return 0, internalError{}
				},
			}
			require.Panics(t, func() {
				p.Sprintf("%Planet")
			})
			require.True(t, calledCustomFunc)
		})
	})
}

func TestBindings(t *testing.T) {
	p := Printer{
		SuppressErrors: true,
		DefaultBindings: map[string]interface{}{
			"Key1": "Value1",
			"Key2": "Value2",
		},
	}
	require.Equal(t, "Value1 Value2 %Key3 %Key4", p.Sprintf("%Key1 %Key2 %Key3 %Key4"))
	require.Equal(t, "Value1 Value2 Value3 %Key4", p.Sprintf("%Key1 %Key2 %Key3 %Key4", map[string]interface{}{"Key3": "Value3"}))
	require.Equal(t, "Value1 Value2 Value3 Value4", p.Sprintf("%Key1 %Key2 %Key3 %Key4", map[string]interface{}{"Key3": "Value3"}, map[string]interface{}{"Key4": "Value4"}))

	// override default binding
	require.Equal(t, "Value2", p.Sprintf("%Key1", map[string]interface{}{"Key1": "Value2"}))

	// override previous binding
	require.Equal(t, "Value3", p.Sprintf("%Key1", map[string]interface{}{"Key1": "Value2"}, map[string]interface{}{"Key1": "Value3"}))

	// empty binding
	require.Equal(t, "Value1", p.Sprintf("%Key1", nil))

	// invalid binding value
	require.Equal(t, "", p.Sprintf("%Key3", map[string]interface{}{
		"Key3": nil,
	}))
	// invalid binding key
	require.Equal(t, "%Key3", p.Sprintf("%Key3", map[complex64]interface{}{
		complex64(1): "test",
	}))

	// invalid binding
	require.Equal(t, "Value1", p.Sprintf("%Key1", 1))

	// ptr binding
	type st struct {
		Key1 string
	}
	require.Equal(t, "Value3", p.Sprintf("%Key1", &st{
		Key1: "Value3",
	}))

	var ptr *st
	require.Equal(t, "Value1", p.Sprintf("%Key1", ptr))
}

func TestNoSuppressErrors(t *testing.T) {
	p := Printer{
		SuppressErrors: false,
	}

	// invalid binding
	require.Panics(t, func() {
		p.Sprintf("%Key1", 1)
	})

	//ptr binding
	require.Panics(t, func() {
		type st struct {
			Key1 string
		}
		var ptr *st
		p.Sprintf("%Key1", ptr)
	})

	// invalid binding value
	require.Panics(t, func() {
		p.Sprintf("%Key1", map[string]interface{}{
			"Key1": nil,
		})
	})

	// invalid binding key
	require.Panics(t, func() {
		p.Sprintf("%1", map[complex64]interface{}{
			complex64(1): "test",
		})
	})
}

func TestMultipleKeys(t *testing.T) {
	require.Equal(t, "Hello World!", Sprintf("Hello %Foo!", map[string]interface{}{
		"Foo":    "World",
		"FooBar": "Jupiter",
		"Bar":    "Mars",
	}))

	require.Equal(t, "Goodbye", Sprintf("%textbye", map[string]interface{}{
		"text":    "Hello",
		"textbye": "Goodbye",
	}))
}

func TestPrefix(t *testing.T) {
	// length
	require.Equal(t, "Hello      Earth!", Sprintf("Hello %10Planet!", map[string]interface{}{"Planet": "Earth"}))
	require.Equal(t, "Hello Kepler-107!", Sprintf("Hello %10Planet!", map[string]interface{}{"Planet": "Kepler-107"}))
	require.Equal(t, "Hello Starkiller Base!", Sprintf("Hello %10Planet!", map[string]interface{}{"Planet": "Starkiller Base"}))

	// explicit left alignment
	require.Equal(t, "Hello      Earth!", Sprintf("Hello %+10Planet!", map[string]interface{}{"Planet": "Earth"}))
	require.Equal(t, "Hello Kepler-107!", Sprintf("Hello %+10Planet!", map[string]interface{}{"Planet": "Kepler-107"}))
	require.Equal(t, "Hello Starkiller Base!", Sprintf("Hello %+10Planet!", map[string]interface{}{"Planet": "Starkiller Base"}))

	// explicit right alignment
	require.Equal(t, "Hello Earth     !", Sprintf("Hello %-10Planet!", map[string]interface{}{"Planet": "Earth"}))
	require.Equal(t, "Hello Kepler-107!", Sprintf("Hello %-10Planet!", map[string]interface{}{"Planet": "Kepler-107"}))
	require.Equal(t, "Hello Starkiller Base!", Sprintf("Hello %-10Planet!", map[string]interface{}{"Planet": "Starkiller Base"}))

	// explicit center alignment
	require.Equal(t, "Hello   Earth   !", Sprintf("Hello %|10Planet!", map[string]interface{}{"Planet": "Earth"}))
	require.Equal(t, "Hello Kepler-107!", Sprintf("Hello %|10Planet!", map[string]interface{}{"Planet": "Kepler-107"}))
	require.Equal(t, "Hello Starkiller Base!", Sprintf("Hello %|10Planet!", map[string]interface{}{"Planet": "Starkiller Base"}))

	require.Equal(t, "Hello ABEarthABA!", Sprintf("Hello %|AB10Planet!", map[string]interface{}{"Planet": "Earth"}))
	require.Equal(t, "Hello Kepler-107!", Sprintf("Hello %|AB10Planet!", map[string]interface{}{"Planet": "Kepler-107"}))
	require.Equal(t, "Hello Starkiller Base!", Sprintf("Hello %|AB10Planet!", map[string]interface{}{"Planet": "Starkiller Base"}))

	// floats
	require.Equal(t, "1.23", Sprintf("%.2f", map[string]interface{}{"f": 1.2345}))
	require.Equal(t, " 1", Sprintf("%2.f", map[string]interface{}{"f": 1.2345}))
	require.Equal(t, "1.234", Sprintf("%2.3f", map[string]interface{}{"f": 1.2345}))
	require.Equal(t, "1.234", Sprintf("%02.3f", map[string]interface{}{"f": 1.2345}))
	require.Equal(t, "01.234", Sprintf("%06.3f", map[string]interface{}{"f": 1.2345}))
	require.Equal(t, " 1.234", Sprintf("%6.3f", map[string]interface{}{"f": 1.2345}))
	require.Equal(t, "ABABA1.234", Sprintf("%AB10.3f", map[string]interface{}{"f": 1.2345}))

	// treat an int as a float
	require.Equal(t, "1.000", Sprintf("%.3f", map[string]interface{}{"f": 1}))
	// treat an uint as a float
	require.Equal(t, "1.000", Sprintf("%.3f", map[string]interface{}{"f": uint(1)}))

	// slices
	require.Equal(t, "Hello Earth!", Sprintf("Hello %.0Planets!", map[string]interface{}{"Planets": []string{"Earth", "Kepler-107", "Starkiller Base"}}))
	require.Equal(t, "Hello Earth!", Sprintf("Hello %.Planets!", map[string]interface{}{"Planets": []string{"Earth", "Kepler-107", "Starkiller Base"}}))
	require.Equal(t, "Hello      Earth!", Sprintf("Hello %+10.0Planets!", map[string]interface{}{"Planets": []string{"Earth", "Kepler-107", "Starkiller Base"}}))
	require.Equal(t, "Hello !", Sprintf("Hello %.4Planets!", map[string]interface{}{"Planets": []string{"Earth", "Kepler-107", "Starkiller Base"}}))

	require.Equal(t, "Hello Earth-----!", Sprintf("Hello %--10Planet!", map[string]interface{}{"Planet": "Earth"}))
	require.Equal(t, "Hello -----Earth!", Sprintf("Hello %+-10Planet!", map[string]interface{}{"Planet": "Earth"}))
	require.Equal(t, "Hello Earth+++++!", Sprintf("Hello %-+10Planet!", map[string]interface{}{"Planet": "Earth"}))
	require.Equal(t, "Hello +++++Earth!", Sprintf("Hello %++10Planet!", map[string]interface{}{"Planet": "Earth"}))
}

func TestCustomVariableToken(t *testing.T) {
	p := &Printer{
		KeyToken: '$',
	}

	require.Equal(t, "Hello Mercury!", p.Sprintf("Hello $Planet!", map[string]interface{}{
		"Planet": "Mercury",
	}))
}

func TestMakeBindings(t *testing.T) {
	t.Run("Map", func(t *testing.T) {
		binds, err := makeBindings(reflect.ValueOf(map[string]interface{}{
			"Key1": "Value1",
			"Key2": "Value2",
		}))
		require.NoError(t, err)

		require.Equal(t, "Value1", defaultReflectPrinter.Sprint(binds.Get([]rune("Key1")).Value))
		require.Equal(t, "Value2", defaultReflectPrinter.Sprint(binds.Get([]rune("Key2")).Value))
	})

	t.Run("Struct", func(t *testing.T) {
		type st struct {
			Key1 string
			Key2 string
		}
		binds, err := makeBindings(reflect.ValueOf(st{
			Key1: "Value1",
			Key2: "Value2",
		}))
		require.NoError(t, err)

		require.Equal(t, "Value1", defaultReflectPrinter.Sprint(binds.Get([]rune("Key1")).Value))
		require.Equal(t, "Value2", defaultReflectPrinter.Sprint(binds.Get([]rune("Key2")).Value))
	})

	t.Run("Struct embedded", func(t *testing.T) {
		type embedded struct {
			Key3 string
		}
		type st struct {
			Key1 string
			Key2 string
			embedded
		}
		binds, err := makeBindings(reflect.ValueOf(st{
			Key1: "Value1",
			Key2: "Value2",
			embedded: embedded{
				Key3: "Value3",
			},
		}))
		require.NoError(t, err)

		require.Equal(t, "Value1", defaultReflectPrinter.Sprint(binds.Get([]rune("Key1")).Value))
		require.Equal(t, "Value2", defaultReflectPrinter.Sprint(binds.Get([]rune("Key2")).Value))
		// require.Equal(t, "Value3", defaultReflectPrinter.Sprint(binds.Get([]rune("Key3")).Value))
	})

	t.Run("Struct Child", func(t *testing.T) {
		type st2 struct {
			Key3 string
		}
		type st struct {
			Key1 string
			Key2 string
			Key3 st2
		}
		binds, err := makeBindings(reflect.ValueOf(st{
			Key1: "Value1",
			Key2: "Value2",
			Key3: st2{
				Key3: "Value3",
			},
		}))
		require.NoError(t, err)

		require.Equal(t, "Value1", defaultReflectPrinter.Sprint(binds.Get([]rune("Key1")).Value))
		require.Equal(t, "Value2", defaultReflectPrinter.Sprint(binds.Get([]rune("Key2")).Value))
		// require.Equal(t, "", defaultReflectPrinter.Sprint(binds.Get([]rune("Key3")).Value))
	})
}

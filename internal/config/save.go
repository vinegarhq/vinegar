package config

import (
	"io"
	"os"
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/vinegarhq/vinegar/internal/dirs"
)

// Encode writes the TOML representation of the [Config] to the given Writer.
func (c *Config) Encode(w io.Writer) error {
	enc := toml.NewEncoder(w)
	enc.Indent = ""
	return enc.Encode(c.Diff())
}

// Diff returns the raw representation value of the configuration with
// the defaults stripped. Used in [Encode].
func (c *Config) Diff() map[string]any {
	return diff(reflect.ValueOf(Default()).Elem(), reflect.ValueOf(c).Elem())
}

// Save calls [Encode] to the default configuration path.
func (c *Config) Save() error {
	if err := os.MkdirAll(dirs.Config, 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(dirs.ConfigPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	return c.Encode(f)
}

func diff(dv, v reflect.Value) map[string]any {
	out := map[string]any{}
	for i := 0; i < dv.NumField(); i++ {
		df, t, f := dv.Field(i), dv.Type().Field(i), v.Field(i)
		name := t.Tag.Get("toml")
		if name == "" {
			panic("config: " + t.Name + " unnamed")
		}

		switch df.Kind() {
		case reflect.Struct:
			if sub := diff(df, f); len(sub) > 0 {
				out[name] = sub
			}
		case reflect.Map:
			m := map[string]any{}
			for _, k := range df.MapKeys() {
				v1, v2 := df.MapIndex(k), f.MapIndex(k)
				if v2.IsValid() && !reflect.DeepEqual(v1.Interface(), v2.Interface()) {
					m[k.Interface().(string)] = v2.Interface()
				}
			}
			for _, k := range f.MapKeys() {
				if !df.MapIndex(k).IsValid() {
					m[k.Interface().(string)] = f.MapIndex(k).Interface()
				}
			}
			if len(m) > 0 {
				out[name] = m
			}
		default:
			if !reflect.DeepEqual(df.Interface(), f.Interface()) {
				out[name] = f.Interface()
			}
		}
	}
	return out
}

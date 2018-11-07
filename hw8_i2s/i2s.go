package main

import (
	"errors"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	return walk(out, data)
}

func walk(u interface{}, data interface {}) (er error) {
	defer func() {
		if r := recover(); r != nil {
			er = errors.New("reflect error")
		}
	}()

	val := reflect.Indirect(reflect.ValueOf(u))
	t := val.Type()

	switch val.Kind() {
	case reflect.Struct:
		data := data.(map[string]interface{})
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldVal := reflect.Indirect(val.Field(i))

			switch fieldVal.Kind() {
			case reflect.Int:
				d := int64(data[field.Name].(float64))
				//fmt.Printf("Field %q = %d\n", field.Name, d)
				fieldVal.SetInt(d)
			case reflect.Bool:
				d := data[field.Name].(bool)
				//fmt.Printf("Field %q = %t\n", field.Name, d)
				fieldVal.SetBool(d)
			case reflect.String:
				d := data[field.Name].(string)
				//fmt.Printf("Field %q = %s\n", field.Name, d)
				fieldVal.SetString(d)
			case reflect.Slice:
				typ := fieldVal.Type().Elem()

				d := data[field.Name].([]interface{})
				for _, u := range d {
					elem := reflect.New(typ)
					walk(elem.Interface(), u)
					fieldVal.Set(reflect.Append(fieldVal, elem.Elem()))
				}
			case reflect.Struct:
				d := data[field.Name]
				elem := reflect.New(field.Type)
				walk(elem.Interface(), d)
				fieldVal.Set(elem.Elem())
			}
		}
	case reflect.Slice:
		data := data.([]interface{})
		typ := val.Type().Elem()

		for _, u := range data {
			elem := reflect.New(typ)
			walk(elem.Interface(), u)
			val.Set(reflect.Append(val, elem.Elem()))
		}
	}

	return nil
}

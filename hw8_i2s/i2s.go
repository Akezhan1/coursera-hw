package main

import (
	"errors"
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	inputV := reflect.ValueOf(data)
	outV := reflect.ValueOf(out)
	if outV.Kind() != reflect.Ptr {
		return errors.New("out type must be ptr")
	}
	if err := set(inputV, outV); err != nil {
		return err
	}
	return nil
}

func set(inp, out reflect.Value) error {
	if !out.IsValid() {
		return errors.New("no such field")
	}
	switch out.Kind() {
	case reflect.Int:
		switch inp.Kind() {
		case reflect.Int:
			out.SetInt(inp.Int())
		case reflect.Float64:
			out.SetInt(int64(inp.Float()))
		default:
			return errors.New(fmt.Sprintf("type %s can`t set to %s", inp.Kind(), out.Kind()))
		}
	case reflect.String:
		if inp.Kind() != reflect.String {
			return errors.New(fmt.Sprintf("type %s can`t set to %s", inp.Kind(), out.Kind()))
		}
		out.SetString(inp.String())
	case reflect.Bool:
		if inp.Kind() != reflect.Bool {
			return errors.New(fmt.Sprintf("type %s can`t set to %s", inp.Kind(), out.Kind()))
		}
		out.SetBool(inp.Bool())
	case reflect.Struct:
		if inp.Kind() != reflect.Map {
			return errors.New(fmt.Sprintf("type %s can`t set to %s", inp.Kind(), out.Kind()))
		}
		for _, key := range inp.MapKeys() {
			sout := out.FieldByName(key.String())
			minp := inp.MapIndex(key).Elem()
			if err := set(minp, sout); err != nil {
				return err
			}
		}
	case reflect.Slice:
		if inp.Kind() != reflect.Slice {
			return errors.New(fmt.Sprintf("type %s can`t set to %s", inp.Kind(), out.Kind()))
		}
		slice := reflect.MakeSlice(out.Type(), inp.Len(), inp.Len())
		for i := 0; i < inp.Len(); i++ {
			if err := set(inp.Index(i).Elem(), slice.Index(i)); err != nil {
				return err
			}
		}
		out.Set(slice)
	case reflect.Ptr, reflect.Interface:
		if err := set(inp, out.Elem()); err != nil {
			return err
		}
	default:
		return errors.New("unknown error")
	}
	return nil
}

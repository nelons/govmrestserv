package internal

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/Jeffail/gabs/v2"
)

func serialise_object(obj any, current *gabs.Container) (*gabs.Container, error) {
	var objData *gabs.Container

	v := reflect.ValueOf(obj)
	t := reflect.TypeOf(obj)

	if t.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
		t = v.Type()
	}

	if t.Kind() != reflect.Struct {
		fmt.Printf("object of type %s is not a struct.\n", t.Name())
		return nil, errors.New("unsupported object type")
	}

	field_count := t.NumField()

	if current == nil {
		objData = gabs.New()
		objData.Set(t.Name(), "_typename")

	} else {
		objData = current
	}

	for i := 0; i < field_count; i++ {
		ft := t.Field(i)
		fv := v.Field(i)

		if fv.Kind() == reflect.Struct && ft.Anonymous {
			serialise_object(fv.Interface(), objData)

		} else {
			serialise_value(objData, fv, ft.Type, ft.Name)

		}
	}

	return objData, nil
}

func serialise_value(objData *gabs.Container, fv reflect.Value, ft reflect.Type, field_name string) error {

	switch fv.Kind() {
	default:
		fmt.Printf("UNHANDLED VALUE name %s, type %T\n", field_name, ft.Kind())
		objData.Set(nil, field_name)

	case reflect.String:
		objData.Set(fv.String(), field_name)

	case reflect.Bool:
		objData.Set(fv.Bool(), field_name)

	case reflect.Int8:
		objData.Set(int8(fv.Int()), field_name)

	case reflect.Int16:
		objData.Set(int16(fv.Int()), field_name)

	case reflect.Int32:
		objData.Set(int32(fv.Int()), field_name)

	case reflect.Int64:
		objData.Set(fv.Int(), field_name)

	case reflect.Uint8:
		objData.Set(uint8(fv.Uint()), field_name)

	case reflect.Uint16:
		objData.Set(uint16(fv.Uint()), field_name)

	case reflect.Uint32:
		objData.Set(uint32(fv.Uint()), field_name)

	case reflect.Uint64:
		objData.Set(uint64(fv.Uint()), field_name)

	case reflect.Struct:
		child, err := serialise_object(fv.Interface(), nil)
		if err == nil {
			objData.Set(child, field_name)

		} else {

			objData.Set(nil, field_name)
		}

	case reflect.Ptr:
		is_valid := fv.IsValid()
		is_nil := fv.IsNil()
		if is_valid && !is_nil {
			child_ptr := fv.Elem()
			serialise_value(objData, child_ptr, reflect.TypeOf(child_ptr), field_name)

		} else {
			objData.Set(nil, field_name)

		}

	case reflect.Slice:
		if fv.Len() > 0 {
			objData.Array(field_name)

			for k := 0; k < fv.Len(); k++ {
				switch ft.Elem().Kind() {
				case reflect.Int:
					iv := fv.Index(k).Int()
					objData.ArrayAppend(iv, field_name)

				case reflect.String:
					str := fv.Index(k).String()
					objData.ArrayAppend(str, field_name)

				case reflect.Int8:
					iv := fv.Index(k).Int()
					objData.ArrayAppend(int8(iv), field_name)

				case reflect.Int16:
					iv := fv.Index(k).Int()
					objData.ArrayAppend(int16(iv), field_name)

				case reflect.Int32:
					iv := fv.Index(k).Int()
					objData.ArrayAppend(int32(iv), field_name)

				case reflect.Int64:
					iv := fv.Index(k).Int()
					objData.ArrayAppend(iv, field_name)

				case reflect.Uint8:
					iv := fv.Index(k).Uint()
					objData.ArrayAppend(uint8(iv), field_name)

				case reflect.Uint16:
					iv := fv.Index(k).Uint()
					objData.ArrayAppend(uint16(iv), field_name)

				case reflect.Uint32:
					iv := fv.Index(k).Uint()
					objData.ArrayAppend(uint32(iv), field_name)

				case reflect.Uint64:
					iv := fv.Index(k).Uint()
					objData.ArrayAppend(iv, field_name)

				case reflect.Struct:
					child := fv.Index(k).Interface()
					child_json, err := serialise_object(child, nil)
					if err == nil {
						objData.ArrayAppend(child_json, field_name)
					}

				case reflect.Interface:
					iv := fv.Index(k).Interface()
					if !fv.IsValid() {
						continue
					}

					it := reflect.TypeOf(iv)
					//fmt.Printf("Interface with type %v\n", it.Kind().String())

					// This should be a pointer
					if it.Kind() != reflect.Ptr {
						fmt.Printf("Item %v in Array for field %v is a Interface but not a pointer.\n", k, field_name)
						continue
					}

					// Get the value pointed to, but continue to next iteration
					// If it points to nothing.
					pv := reflect.Indirect(fv)
					if pv == reflect.Zero(it) {
						fmt.Printf("Item %v in Array for field %v is a Pointer with a zero value.\n", k, field_name)
						continue
					}

					if reflect.TypeOf(pv).Kind() != reflect.Struct {
						fmt.Printf("Item %v in Array for field %v is not a struct but a %v\n", k, field_name, reflect.TypeOf(pv).Kind().String())
						continue
					}

					// We only write structs here !
					//child, err := serialise_interface_json(iv)
					child, err := serialise_object(iv, nil)
					if err == nil {
						objData.ArrayAppend(child, field_name)
					}

				default:
					fmt.Printf("SLICE - Field name %s of type %v and with length %v\n", field_name, ft.Elem().Kind().String(), fv.Len())
				}
			}

		} else {
			objData.Set(nil, field_name)

		}

	case reflect.Interface:
		if fv.CanInterface() {
			iv := fv.Interface()
			if iv != nil {
				it := reflect.TypeOf(iv)
				//pv := reflect.Indirect(fv)

				if it.Kind() == reflect.Ptr {
					child, _ := serialise_object(iv, nil)
					if child != nil {
						objData.Set(child, field_name)
					}

				} else {
					tv := reflect.ValueOf(iv)
					err := serialise_value(objData, tv, it, field_name)
					if err != nil {
						fmt.Printf("Field %v is an Interface but not a PTR. It's a %v.\n", field_name, it.Kind().String())
					}
				}

				/*if it != nil {
					fmt.Printf("Field %v has interface with type %v\n", field_name, it.Kind().String())
				} else {
					fmt.Printf("Field %v has interface but typeOf(value) is nil.\n", field_name)
				}*/

			} else {
				//fmt.Printf("Field %v interface value is nil.\n", field_name)
				objData.Set(nil, field_name)

			}
		}
	}

	return nil
}

package configByTag

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	envTagName string = "env"
	required   string = "required"
	defValName string = "defVal"
	defValSep  string = ":"

	errMsgPointer          string = "Structure must be passed as a pointer"
	errMsgCanSet           string = "Unmutable structure"
	errMsgRequired         string = "Required env parameter %s not filled"
	infoMsgUnsupportedType string = "Unsupported type: %s, field: %s, env: %s"
)

func Load(c interface{}) []error {
	p := reflect.ValueOf(c)
	if p.Type().Kind() != reflect.Ptr {
		log.Fatal(errMsgPointer)
	}
	v := p.Elem()
	if !v.CanSet() {
		return []error{errors.New(errMsgCanSet)}
	}
	return parseField(v)
}

func parseField(v reflect.Value) (errs []error) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fieldValue := v.Field(i)
		if f.Type.Kind() == reflect.Struct {
			if err := parseField(fieldValue); err != nil {
				errs = append(errs, err...)
			}
		} else {
			if err := runTagField(f, &fieldValue); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errs
}

func runTagField(field reflect.StructField, value *reflect.Value) error {
	tf := &tagField{
		field: field,
		value: value,
	}
	return tf.Run()
}

type tagField struct {
	field     reflect.StructField
	value     *reflect.Value
	name      string
	isReqired bool
	defValue  string
}

func (tf *tagField) Run() error {
	tf.parse()
	return tf.set()
}
func (tf *tagField) parse() {
	tagValue := tf.field.Tag.Get(envTagName)
	if tagValue == "" {
		return
	}
	params := strings.Split(tagValue, ",")
	if len(params) > 0 {
		tf.name = params[0]
	}
	for i := 1; i < len(params); i++ {
		v := params[i]
		// Required
		if v == required {
			tf.isReqired = true
		}
		// DefVal
		if strings.Index(v, defValName+defValSep) == 0 {
			tf.defValue = v[len(defValName+defValSep):]
		}
	}
}
func (tf *tagField) set() error {
	switch tf.field.Type {
	// time.Duration
	case reflect.TypeOf(time.Duration(1)):
		return tf.durationType()
	// []string
	case reflect.TypeOf([]string{}):
		return tf.sliceType(tf.stringSlice)

	// []int
	case reflect.TypeOf([]int{}):
		return tf.sliceType(tf.intSlice(0))
	case reflect.TypeOf([]int8{}):
		return tf.sliceType(tf.intSlice(8))
	case reflect.TypeOf([]int16{}):
		return tf.sliceType(tf.intSlice(16))
	case reflect.TypeOf([]int32{}):
		return tf.sliceType(tf.intSlice(32))
	case reflect.TypeOf([]int32{}):
		return tf.sliceType(tf.intSlice(64))

	// []uint
	case reflect.TypeOf([]uint{}):
		return tf.sliceType(tf.uintSlice(0))
	case reflect.TypeOf([]uint8{}):
		return tf.sliceType(tf.uintSlice(8))
	case reflect.TypeOf([]uint16{}):
		return tf.sliceType(tf.uintSlice(16))
	case reflect.TypeOf([]uint32{}):
		return tf.sliceType(tf.uintSlice(32))
	case reflect.TypeOf([]uint32{}):
		return tf.sliceType(tf.uintSlice(64))

	// []bool
	case reflect.TypeOf([]bool{}):
		return tf.sliceType(tf.boolSlice)
	}

	kind := tf.field.Type.Kind()
	switch kind {
	case reflect.String:
		return tf.stringType()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return tf.intType()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return tf.uintType()
	case reflect.Bool:
		return tf.boolType()
	case reflect.Float32, reflect.Float64:
		return tf.floatType()
	default:
		return errors.New(fmt.Sprintf(infoMsgUnsupportedType, kind, tf.field.Name, tf.name))
	}
}

func (tf *tagField) getEnv() (string, error) {
	s, exists := os.LookupEnv(tf.name)
	if !exists {
		if tf.defValue != "" {
			s = tf.defValue
		} else if tf.isReqired {
			return "", errors.New(fmt.Sprintf(errMsgRequired, tf.name))
		}
	}
	return s, nil
}
func (tf *tagField) stringType() error {
	s, err := tf.getEnv()
	if err != nil {
		return err
	}
	tf.value.SetString(s)
	return nil
}
func (tf *tagField) intType() error {
	s, err := tf.getEnv()
	if err != nil {
		return err
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	tf.value.SetInt(n)
	return nil
}
func (tf *tagField) uintType() error {
	s, err := tf.getEnv()
	if err != nil {
		return err
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	tf.value.SetUint(n)
	return nil
}
func (tf *tagField) floatType() error {
	s, err := tf.getEnv()
	if err != nil {
		return err
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	tf.value.SetFloat(n)
	return nil
}
func (tf *tagField) boolType() error {
	s, err := tf.getEnv()
	if err != nil {
		return err
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	tf.value.SetBool(b)
	return nil
}
func (tf *tagField) durationType() error {
	s, err := tf.getEnv()
	if err != nil {
		return err
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	tf.value.SetInt(int64(d))
	return nil
}

type fType func(string) (v reflect.Value, err error)

func (tf *tagField) sliceType(f fType) error {
	sEnv, err := tf.getEnv()
	if err != nil {
		return err
	}
	value := reflect.New(tf.value.Type()).Elem()
	for _, s := range strings.Split(sEnv, ",") {
		val, err := f(s)
		if err != nil {
			return err
		}
		value = reflect.Append(value, val)
	}
	tf.value.Set(value)
	return nil

}
func (tf *tagField) stringSlice(s string) (v reflect.Value, err error) {
	return reflect.ValueOf(s), nil
}
func (tf *tagField) intSlice(bitSize int) fType {
	return func(s string) (v reflect.Value, err error) {
		var n int64
		if n, err = strconv.ParseInt(s, 10, bitSize); err != nil {
			return v, err
		}
		switch bitSize {
		case 0:
			v = reflect.ValueOf(int(n))
		case 8:
			v = reflect.ValueOf(int8(n))
		case 16:
			v = reflect.ValueOf(int16(n))
		case 32:
			v = reflect.ValueOf(int32(n))
		case 64:
			v = reflect.ValueOf(int64(n))
		}
		return v, nil
	}
}
func (tf *tagField) uintSlice(bitSize int) fType {
	return func(s string) (v reflect.Value, err error) {
		var n uint64
		if n, err = strconv.ParseUint(s, 10, bitSize); err != nil {
			return v, err
		}
		switch bitSize {
		case 0:
			v = reflect.ValueOf(uint(n))
		case 8:
			v = reflect.ValueOf(uint8(n))
		case 16:
			v = reflect.ValueOf(uint16(n))
		case 32:
			v = reflect.ValueOf(uint32(n))
		case 64:
			v = reflect.ValueOf(uint64(n))
		}
		return v, nil
	}
}
func (tf *tagField) boolSlice(s string) (v reflect.Value, err error) {
	var b bool
	if b, err = strconv.ParseBool(s); err != nil {
		return v, err
	}
	return reflect.ValueOf(b), nil
}

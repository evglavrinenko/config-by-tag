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
	separator  string = ":"
	minName    string = "min"
	maxName    string = "max"

	errMsgPointer          string = "Structure must be passed as a pointer"
	errMsgCanSet           string = "Unmutable structure"
	errMsgRequired         string = "Required env parameter %s not filled"
	errMsgMin              string = "The value: %s is less than the min: %v < %v"
	errMsgMax              string = "The value: %s is greater than the max: %v > %v"
	infoMsgUnsupportedType string = "Unsupported type: %s, field: %s, env: %s"
)

func Load(c interface{}) error {
	p := reflect.ValueOf(c)
	if p.Type().Kind() != reflect.Ptr {
		log.Fatal(errMsgPointer)
	}
	v := p.Elem()
	if !v.CanSet() {
		return errors.New(errMsgCanSet)
	}

	return parseField(v)
}

func parseField(v reflect.Value) (err error) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fieldValue := v.Field(i)
		if f.Type.Kind() == reflect.Struct {
			if err = parseField(fieldValue); err != nil {
				return err
			}
		} else {
			if fieldValue.CanSet() {
				if err = runTagField(f, &fieldValue); err != nil {
					return err
				}
			}
		}
	}
	return nil
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
	min       string
	max       string
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
		name := defValName + separator
		if strings.Index(v, name) == 0 {
			tf.defValue = v[len(name):]
		}
		//Min
		name = minName + separator
		if strings.Index(v, name) == 0 {
			tf.min = v[len(name):]
		}
		//Max
		name = maxName + separator
		if strings.Index(v, name) == 0 {
			tf.max = v[len(name):]
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
	// Validation
	if err = tf.valid(tf.compareString(s)); err != nil {
		return err
	}

	tf.value.SetString(s)
	return nil
}
func (tf *tagField) intType() (err error) {
	var s string
	if s, err = tf.getEnv(); err != nil {
		return err
	} else if strings.TrimSpace(s) == "" {
		return nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	// Validation
	if err = tf.valid(tf.compareInt64(n)); err != nil {
		return err
	}
	tf.value.SetInt(n)
	return nil
}
func (tf *tagField) uintType() (err error) {
	var s string
	if s, err = tf.getEnv(); err != nil {
		return err
	} else if strings.TrimSpace(s) == "" {
		return nil
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	// Validation
	if err = tf.valid(tf.compareUint64(n)); err != nil {
		return err
	}
	tf.value.SetUint(n)
	return nil
}
func (tf *tagField) floatType() (err error) {
	var s string
	if s, err = tf.getEnv(); err != nil {
		return err
	} else if strings.TrimSpace(s) == "" {
		return nil
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	// Validation
	if err = tf.valid(tf.compareFloat64(n)); err != nil {
		return err
	}
	tf.value.SetFloat(n)
	return nil
}
func (tf *tagField) boolType() (err error) {
	var s string
	if s, err = tf.getEnv(); err != nil {
		return err
	} else if strings.TrimSpace(s) == "" {
		return nil
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	tf.value.SetBool(b)
	return nil
}
func (tf *tagField) durationType() (err error) {
	var (
		s string
		d time.Duration
	)
	if s, err = tf.getEnv(); err != nil {
		return err
	} else if strings.TrimSpace(s) == "" {
		return nil
	}
	if d, err = time.ParseDuration(s); err != nil {
		return err
	}
	// Validation
	if err = tf.valid(tf.compareDuration(d)); err != nil {
		return err
	}
	tf.value.SetInt(int64(d))
	return nil
}

type TCompareFunc func(isMin bool, field, msg string) error

func (tf *tagField) valid(cf TCompareFunc) error {
	if tf.min != "" {
		if err := cf(true, tf.min, errMsgMin); err != nil {
			return err
		}
	}
	if tf.max != "" {
		if err := cf(false, tf.max, errMsgMax); err != nil {
			return err
		}
	}
	return nil
}

func (tf *tagField) compareString(s string) TCompareFunc {
	return func(isMin bool, field, msg string) error {
		compValue, err := strconv.Atoi(field)
		if err != nil {
			return err
		}
		if (len(s) < compValue && isMin) || (len(s) > compValue && !isMin) {
			return errors.New(fmt.Sprintf(msg, tf.name, field))
		}
		return nil
	}
}
func (tf *tagField) compareInt64(n int64) TCompareFunc {
	return func(isMin bool, field, msg string) error {
		compValue, err := strconv.ParseInt(field, 10, 64)
		if err != nil {
			return err
		}
		if (n < compValue && isMin) || (n > compValue && !isMin) {
			return errors.New(fmt.Sprintf(msg, tf.name, field))
		}
		return nil
	}
}
func (tf *tagField) compareUint64(n uint64) TCompareFunc {
	return func(isMin bool, field, msg string) error {
		compValue, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return err
		}
		if (n < compValue && isMin) || (n > compValue && !isMin) {
			return errors.New(fmt.Sprintf(msg, tf.name, field))
		}
		return nil
	}
}
func (tf *tagField) compareFloat64(n float64) TCompareFunc {
	return func(isMin bool, field, msg string) error {
		compValue, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return err
		}
		if (n < compValue && isMin) || (n > compValue && !isMin) {
			return errors.New(fmt.Sprintf(msg, tf.name, n, field))
		}
		return nil
	}
}
func (tf *tagField) compareDuration(d time.Duration) TCompareFunc {
	return func(isMin bool, field, msg string) error {
		compValue, err := time.ParseDuration(field)
		if err != nil {
			return err
		}
		if (d < compValue && isMin) || (d > compValue && !isMin) {
			return errors.New(fmt.Sprintf(msg, tf.name, d, field))
		}
		return nil
	}
}

type fType func(string) (v reflect.Value, err error)

func (tf *tagField) sliceType(f fType) error {
	sEnv, err := tf.getEnv()
	if err != nil {
		return err
	}
	value := reflect.New(tf.value.Type()).Elem()
	slice := strings.Split(sEnv, ",")
	for _, s := range slice {
		val, err := f(s)
		if err != nil {
			return err
		}
		value = reflect.Append(value, val)
	}
	// Validation
	if err = tf.valid(tf.compareInt64(int64(len(slice)))); err != nil {
		return err
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

package configTag

import "C"
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
	kind := tf.field.Type.Kind()
	if tf.field.Type.String() == "time.Duration" {
		return tf.durationType()
	}
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

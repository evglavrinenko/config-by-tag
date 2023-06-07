package configByTag

import (
	"reflect"
	"testing"
	"time"
)

type Config struct {
	testUnEnv string
	//template  struct {
	//	back       string
	//	backStatic string
	//	storage    string
	//}
	TestString string `env:"ENV_STRING,defVal:1"`
	//TestStringRequired string        `env:"ENV_STRING_R,required"`
	TestDuration  time.Duration `env:"ENV_DURATION"`
	TestDuration2 time.Duration `env:"ENV_DURATION2,defVal:4s,min:1s,max:1m"`
	TestBool      bool          `env:"ENV_BOOLEAN1,defVal:true"`
	TestInt       int           `env:"ENV_INT,defVal:1"`
	TestInt8      int8          `env:"ENV_INT8,defVal:8"`

	TestFloat32 float32 `env:"ENV_FLOAT32,defVal:143.123,min:140"`
	TestFloat64 float64 `env:"ENV_FLOAT64,defVal:18.9,max:123.12"`

	Block struct {
		TestUint   uint  `env:"ENV_UINT,defVal:2"`
		TestUint16 uint8 `env:"ENV_UINT16,defVal:16"`
		SubBlock   struct {
			TestSubUint uint `env:"ENT_UINT_SUB,defVal:9,min:1"`
		}
	}
	TestSliceString []string `env:"ENV_SLICE_STRING"`
	TestSliceInt    []uint8  `env:"ENV_SLICE_INT"`
	TestSliceBool   []bool   `env:"ENV_SLICE_BOOL"`
}

func TestLoad(t *testing.T) {
	var Conf Config

	t.Setenv("ENV_STRING", "testString")
	//t.Setenv("ENV_STRING_R", "testStringRequired")
	t.Setenv("ENV_DURATION", (3 * time.Second).String())
	t.Setenv("ENV_BOOLEAN", "true")
	t.Setenv("ENV_INT8", "7")
	t.Setenv("ENV_UINT", "2")
	t.Setenv("ENV_SLICE_STRING", "stroka1,stroka2")
	t.Setenv("ENV_SLICE_INT", "5,6,7")
	t.Setenv("ENV_SLICE_BOOL", "true,false,true")

	//t.Setenv("ENV_FLOAT32", "7.12")
	t.Setenv("ENV_FLOAT64", "43.12")

	Conf.TestSliceString = []string{"test1"}
	if err := Load(&Conf); err != nil {
		t.Error(err)
	}

	if Conf.TestString != "testString" {
		t.Error("Test string")
	}
	if Conf.TestDuration != time.Second*3 {
		t.Error("Test duration")
	}
	if Conf.TestDuration2 != time.Second*4 {
		t.Error("Test duration2")
	}

	if !Conf.TestBool {
		t.Error("Test bool")
	}

	s := []string{"stroka1", "stroka2"}
	if !reflect.DeepEqual(Conf.TestSliceString, s) {
		t.Error("Test []string")
	}

	n := []uint8{5, 6, 7}
	if !reflect.DeepEqual(Conf.TestSliceInt, n) {
		t.Error("Test []uint8")
	}

	b := []bool{true, false, true}
	if !reflect.DeepEqual(Conf.TestSliceBool, b) {
		t.Error("Test []bool")
	}
}

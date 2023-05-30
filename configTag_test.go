package configTag

import (
	"strings"
	"testing"
	"time"
)

type Config struct {
	TestString         string        `env:"ENV_STRING,defVal:1"`
	TestStringRequired string        `env:"ENV_STRING_R,required"`
	TestDuration       time.Duration `env:"ENV_DURATION"`
	TestDuration2      time.Duration `env:"ENV_DURATION2,defVal:4s"`
	TestBool           bool          `env:"ENV_BOOLEAN1,defVal:true"`
	TestInt            int           `env:"ENV_INT,defVal:1"`
	TestInt8           int8          `env:"ENV_INT8,defVal:8"`

	Block struct {
		TestUint   uint  `env:"ENV_UINT,defVal:2"`
		TestUint16 uint8 `env:"ENV_UINT16,defVal:16"`
		SubBlock   struct {
			TestSubUint uint `env:"ENT_UINT_SUB,defVal:9"`
		}
	}
	TestUnsupport []string `env:"ENV_UNSUPPORT"`
}

func TestLoad(t *testing.T) {
	var Conf Config

	t.Setenv("ENV_STRING", "testString")
	//t.Setenv("ENV_STRING_R", "testStringRequired")
	t.Setenv("ENV_DURATION", (3 * time.Second).String())
	t.Setenv("ENV_BOOLEAN", "true")
	t.Setenv("ENV_INT8", "7")
	t.Setenv("ENV_INT8", "7")
	t.Setenv("ENV_UINT", "2")

	errs := Load(&Conf)
	//t.Log(Conf)

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

	ok := 0
	for _, err := range errs {
		if strings.TrimSpace(err.Error()) == "Required env parameter ENV_STRING_R not filled" {
			ok++
		}
		if strings.TrimSpace(err.Error()) == "Unsupported type: slice, field: TestUnsupport, env: ENV_UNSUPPORT" {
			ok++
		}
	}

	if ok != 2 {
		t.Error(errs)
	}
}

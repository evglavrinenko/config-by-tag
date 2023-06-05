# configByTag

Поддерживаемые типы:
 - string
 - int, int8-64
 - uint, uint8-uint64
 - bool
 - float32, float64
 - time.Duration
 - []string
 - []int, []int8-[]int64
 - []uint, []uint8-[]uint64
 - []bool

type SomeConfig struct {
    Test string `env:"ENV_TEST,required,defVal:test-string"`
}
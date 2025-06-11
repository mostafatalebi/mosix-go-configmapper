package inputs

import (
	"errors"
	"os"
	"strconv"
)

const InputEnvName = "env"

func NewOsEnv() *InputOsEnv {
	return &InputOsEnv{}
}

type InputOsEnv struct {
}

func (e *InputOsEnv) GetBoolean(key string) (bool, error) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return false, errors.New("key is not found")
	}
	return strconv.ParseBool(v)
}
func (e *InputOsEnv) GetNumber(key string) (float64, error) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return 0, errors.New("key is not found")
	}
	return strconv.ParseFloat(v, 64)
}
func (e *InputOsEnv) GetString(key string) (string, error) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return "", errors.New("key is not found")
	}
	return v, nil
}

func (e *InputOsEnv) CanRefresh() bool {
	return false
}

func (e *InputOsEnv) Has(key string) bool {
	_, ok := os.LookupEnv(key)
	return ok
}

func (e *InputOsEnv) GetInputName() string {
	return InputEnvName
}

func (e *InputOsEnv) Reload() error {
	return errors.New("is not implemented")
}

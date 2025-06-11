package inputs

import (
	"errors"
)

type InputMock struct {
	KeysBool   map[string]bool
	KeysNumber map[string]float64
	KeysStr    map[string]string

	ShouldErr map[string]error
}

const InputMockName = "mock"

func NewInputMock() *InputMock {
	return &InputMock{
		KeysBool:   make(map[string]bool),
		KeysNumber: make(map[string]float64),
		KeysStr:    make(map[string]string),
		ShouldErr:  make(map[string]error),
	}
}

func (f *InputMock) Has(key string) bool {
	if _, ok := f.ShouldErr[key]; ok {
		return false
	}
	if _, ok := f.KeysStr[key]; ok {
		return true
	}
	return false
}

func (f *InputMock) GetBoolean(key string) (bool, error) {
	if v, ok := f.ShouldErr[key]; ok {
		return false, v
	}
	if v, ok := f.KeysBool[key]; ok {
		return v, nil
	}
	return false, errors.New("key not found")
}
func (f *InputMock) GetNumber(key string) (float64, error) {
	if v, ok := f.ShouldErr[key]; ok {
		return 0, v
	}
	if v, ok := f.KeysNumber[key]; ok {
		return v, nil
	}
	return 0, errors.New("key not found")
}
func (f *InputMock) GetString(key string) (string, error) {
	if v, ok := f.ShouldErr[key]; ok {
		return "", v
	}
	if v, ok := f.KeysStr[key]; ok {
		return v, nil
	}
	return "", errors.New("key not found")
}

func (f *InputMock) ShouldError(keyName string, err error) *InputMock {
	f.ShouldErr[keyName] = err
	return f
}

func (f *InputMock) CanRefresh() bool {
	return false
}

// ShouldReturn if previously you have used ShouldErr(), use this function to remove that condition
func (f *InputMock) ShouldReturn(keyName string) *InputMock {
	if _, ok := f.ShouldErr[keyName]; ok {
		delete(f.ShouldErr, keyName)
	}
	return f
}

func (f *InputMock) GetInputName() string {
	return InputMockName
}

func (f *InputMock) Reload() error {
	return errors.New("is not implemented")
}

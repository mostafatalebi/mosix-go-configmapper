package inputs

import (
	"errors"
)

type InputMock struct {
	keysBool   map[string]bool
	keysNumber map[string]float64
	keysStr    map[string]string

	shouldErr map[string]error
}

const InputMockName = "mock"

func NewInputMock() *InputMock {
	return &InputMock{
		keysBool:   make(map[string]bool),
		keysNumber: make(map[string]float64),
		keysStr:    make(map[string]string),
		shouldErr:  make(map[string]error),
	}
}

func (f *InputMock) Has(key string) bool {
	if _, ok := f.shouldErr[key]; ok {
		return false
	}
	if _, ok := f.keysStr[key]; ok {
		return true
	}
	return false
}

func (f *InputMock) GetBoolean(key string) (bool, error) {
	if v, ok := f.shouldErr[key]; ok {
		return false, v
	}
	if v, ok := f.keysBool[key]; ok {
		return v, nil
	}
	return false, errors.New("key not found")
}
func (f *InputMock) GetNumber(key string) (float64, error) {
	if v, ok := f.shouldErr[key]; ok {
		return 0, v
	}
	if v, ok := f.keysNumber[key]; ok {
		return v, nil
	}
	return 0, errors.New("key not found")
}
func (f *InputMock) GetString(key string) (string, error) {
	if v, ok := f.shouldErr[key]; ok {
		return "", v
	}
	if v, ok := f.keysStr[key]; ok {
		return v, nil
	}
	return "", errors.New("key not found")
}

func (f *InputMock) ShouldError(keyName string, err error) *InputMock {
	f.shouldErr[keyName] = err
	return f
}

func (f *InputMock) CanRefresh() bool {
	return false
}

// ShouldReturn if previously you have used ShouldErr(), use this function to remove that condition
func (f *InputMock) ShouldReturn(keyName string) *InputMock {
	if _, ok := f.shouldErr[keyName]; ok {
		delete(f.shouldErr, keyName)
	}
	return f
}

func (f *InputMock) GetInputName() string {
	return InputMockName
}

func (f *InputMock) Reload() error {
	return errors.New("is not implemented")
}

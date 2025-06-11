package configmapper

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"mosix-go-configmapper/inputs"
	"mosix-go-configmapper/types"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	datasize "github.com/c2h5oh/datasize"
)

const (
	ValidationTagName = "validation"
	ReasonRequired    = "required"
	ReasonNotFound    = "notFound"
	ReasonValidation  = "validation"

	ErrorCritical = "critical"
)

// NewInputController
// creates an input controller and tries to load the values of fields of the given struct
// by trying each input. If an input returns a value, it will keep that value for the field
// and ignores other inputs and continues to the next field.
// Hence, it is important to pass the most important input as the first input the list of inputs
// if there are several inputs passed (not important if you pass one input only).
// tagName string the name of the tag to look for inside struct's field tag list, default: tag
// defaultTagName string the name of the tag to look for the default value, in case any. default: default
func NewInputController(tagName, defaultTagName string, input ...inputs.ValueInputInterface) *InputController {
	if len(input) == 0 {
		panic("at least on ValueInput is required")
	}
	if tagName == "" {
		tagName = "tag"
	}
	if defaultTagName == "" {
		defaultTagName = "default"
	}
	return &InputController{
		lock:                 &sync.RWMutex{},
		input:                input,
		tagName:              tagName,
		defaultTagName:       defaultTagName,
		validationErrors:     make(map[string]map[string]string),
		internalCacheInt:     map[string]int{},
		internalCacheString:  map[string]string{},
		internalCacheBoolean: map[string]bool{},
		internalCacheFloat:   map[string]float64{},
		internalCacheUnInt:   map[string]uint64{},
	}
}

type InputController struct {
	lock  *sync.RWMutex
	input []inputs.ValueInputInterface
	// this is the name of the tag
	// looked for in passed struct's field
	tagName string

	// this is the name of the tag which is used
	// to load default value if ValueInputInterface doesn't return
	// anything
	defaultTagName string

	autoRefreshList     map[string]bool
	autoRefreshInterval time.Duration

	// in format of: map[validationName][]fieldNames
	validationErrors map[string]map[string]string

	internalCacheInt     map[string]int
	internalCacheString  map[string]string
	internalCacheBoolean map[string]bool
	internalCacheFloat   map[string]float64
	internalCacheUnInt   map[string]uint64

	enablePreprocessors bool
}

func (f *InputController) TogglePreprocessors(v bool) *InputController {
	f.enablePreprocessors = v
	return f
}

func (f *InputController) GetValidationError(field string, reason string) string {
	if _, ok := f.validationErrors[reason]; ok {
		if vvv, ok := f.validationErrors[reason][field]; ok {
			return vvv
		}
	}
	return ""
}

func (f *InputController) Base64Decode(v string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return v, err
	}
	return string(b), nil
}

func (f *InputController) Base64Encode(v string) string {
	return base64.StdEncoding.EncodeToString([]byte(v))
}

func (f *InputController) URLEncode(v string) string {
	return url.QueryEscape(v)
}

func (f *InputController) URLDecode(v string) (string, error) {
	return url.QueryUnescape(v)
}

// returns err and
func (f *InputController) UrlParse(v string, rules map[string]string) error {
	var urlObj, err = url.Parse(v)
	if err != nil {
		return fmt.Errorf("%s is not a correct URL, gor parsing error: $%s", v, err.Error())
	}

	if pRule, ok := rules[types.VdProtocols]; !ok || v == "" {
		return nil
	} else {
		protocols := strings.Split(pRule, ",")
		for _, p := range protocols {
			if p == urlObj.Scheme {
				return nil
			}
		}
		return fmt.Errorf(types.VdProtocols+": url %s is not among allowed protocols %s", v, pRule)
	}
}

func (f *InputController) JsonDecode(v string, obj interface{}) error {
	if err := json.Unmarshal([]byte(v), obj); err != nil {
		return err
	} else {
		return nil
	}
}

func (f *InputController) TimeDurationParse(v string) time.Duration {
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0
	}
	return d
}

// CheckStringPreProcessors checks to see if any string syntax are used or not, if used, then tries to parse the
// value accordingly and in case of success, returns the used syntax as well
func (f *InputController) CheckStringPreProcessors(v string, rules map[string]string) (string, Syntax, error) {
	if f.enablePreprocessors && strings.Index(v, SyntaxBase64Decoding) == 0 {
		v, err := f.Base64Decode(strings.Replace(v, SyntaxBase64Decoding, "", 1))
		if err != nil {
			return v, "", nil
		}
		return v, SyntaxBase64Decoding, nil
	}
	if f.enablePreprocessors && strings.Index(v, SyntaxBase64Encoding) == 0 {
		v = f.Base64Encode(strings.Replace(v, SyntaxBase64Encoding, "", 1))
		return v, SyntaxBase64Encoding, nil
	}
	if f.enablePreprocessors && strings.Index(v, SyntaxURLEncode) == 0 {
		v = f.URLEncode(strings.Replace(v, SyntaxURLEncode, "", 1))
		return v, SyntaxURLEncode, nil
	}
	if f.enablePreprocessors && strings.Index(v, SyntaxURLDecode) == 0 {
		v2, err := f.URLDecode(strings.Replace(v, SyntaxURLDecode, "", 1))
		if err != nil {
			return v, "", nil
		}
		return v2, SyntaxURLDecode, nil
	}
	if f.enablePreprocessors && strings.Index(v, SyntaxURLParse) == 0 {
		rawUrl := strings.Replace(v, SyntaxURLParse, "", 1)
		err := f.UrlParse(rawUrl, rules)
		if err != nil {
			if strings.Index(err.Error(), types.VdProtocols) == 0 {
				return v, "", err
			}
			return v, "", nil
		}
		return rawUrl, SyntaxURLParse, nil
	}
	return v, "", nil
}

func (f *InputController) CheckTimeDurationPreprocessor(v string) time.Duration {
	if f.enablePreprocessors && strings.Index(v, SyntaxTimeDuration) == 0 {
		return f.TimeDurationParse(strings.Replace(v, SyntaxTimeDuration, "", 1))
	}
	return 0
}
func (f *InputController) IsDataSize(v string) bool {
	return strings.Contains(v, SyntaxDataSize)
}
func (f *InputController) CheckDataSize(v string) (int64, error) {
	if f.enablePreprocessors && strings.Index(v, SyntaxDataSize) == 0 {
		var data = strings.Replace(v, SyntaxDataSize, "", 1)
		b, err := datasize.Parse([]byte(data))
		if err == nil {
			return int64(b.Bytes()), nil
		}
	}
	return 0, errors.New("no data-size string found")
}

func (f *InputController) CheckObjectPreprocessor(v string, obj interface{}) error {
	if f.enablePreprocessors && strings.Index(v, SyntaxJsonObject) == 0 {
		var err = f.JsonDecode(strings.Replace(v, SyntaxJsonObject, "", 1), obj)
		return err
	}
	return nil
}

// CheckIntArray
// it will create an integer array from a comma separated list of string numbers
// such as: 20,1, 0, 48, -27
// if a number comes with a space before or after it, the function automatically
// trims them
func (f *InputController) CheckIntArray(v string) ([]int, error) {
	if f.enablePreprocessors && strings.Index(v, SyntaxArrayInt) == 0 {
		nums := strings.Split(strings.Replace(v, SyntaxArrayInt, "", 1), ",")
		if len(nums) > 0 {
			var numbsCast = make([]int, 0)
			for _, v := range nums {
				n, err := strconv.ParseInt(strings.Trim(v, " "), 10, 64)
				if err != nil {
					fmt.Printf("config error: failed to cast str value %s to int\n", v)
					continue
				}
				numbsCast = append(numbsCast, int(n))
			}
			return numbsCast, nil
		}
	}
	return nil, nil
}

func (f *InputController) CheckFloatArray(v string) ([]float64, error) {
	if f.enablePreprocessors && strings.Index(v, SyntaxArrayFloat) == 0 {
		nums := strings.Split(strings.Replace(v, SyntaxArrayFloat, "", 1), ",")
		if len(nums) > 0 {
			var numbsCast = make([]float64, 0)
			for _, v := range nums {
				n, err := strconv.ParseFloat(strings.Trim(v, " "), 64)
				if err != nil {
					fmt.Printf("config error: failed to cast str value %s to float\n", v)
					continue
				}
				numbsCast = append(numbsCast, n)
			}
			return numbsCast, nil
		}
	}
	return nil, nil
}

func (f *InputController) CheckStrArray(v string) ([]string, error) {
	if f.enablePreprocessors && strings.Index(v, SyntaxArrayStr) == 0 {
		vals := strings.Split(strings.Replace(v, SyntaxArrayStr, "", 1), ",")
		return vals, nil
	}
	return nil, nil
}

// FetchKeysAndMapThem
// this function iterates over config object and picks up their ENV tag
// and finds that ENV key name in the input source, if not found, it uses a
// predefined default value. If feature hub connection is not found, it returns
// an error describing the condition
//
// configObj must be a variable of type struct (by value or pointer)
// if not, it panics
// It uses this struct as a base map to search for the keys&values
//
// For getting the errors of parsing/validation, you need to call f.HasCriticalErrors()
func (f *InputController) FetchKeysAndMapThem(configObj any) (err error) {
	if configObj == nil {
		err = errors.New("config object is null and cannot be mapped")
		return
	}
	var configTypes = reflect.TypeOf(configObj).Elem()
	var configValue = reflect.ValueOf(configObj)
	var fieldsCount = configTypes.NumField()
	for i := 0; i < fieldsCount; i++ {
		var currentField = configTypes.Field(i)
		tagValue := currentField.Tag
		fieldKeyName := tagValue.Get(f.tagName)
		if fieldKeyName == "" {
			continue
		}
		var currentFieldType = currentField.Type.String()
		var isStruct bool

		if currentField.Type.Kind() == reflect.Pointer {
			if currentField.Type.Elem().Kind() == reflect.Struct {
				isStruct = true
			}
		}

		if currentField.Type.Kind() == reflect.Map {
			isStruct = true
		}

		f.iterateOverTypes(i, currentFieldType, fieldKeyName, &tagValue, &configValue, isStruct)
	}

	return nil
}

// getValidationTags searches the validation sets of tags and returns a map of found tags
// with their values. To see list of tags, validation.tags file.
func (f *InputController) getValidationTags(tagValue *reflect.StructTag) map[string]string {
	var rules = map[string]string{}
	if vv, ok := tagValue.Lookup(types.VdSet); ok {
		rules[types.VdSet] = vv
	}
	if vv, ok := tagValue.Lookup(types.VdRange); ok {
		rules[types.VdRange] = vv
	}
	if vv, ok := tagValue.Lookup(types.VdGt); ok {
		rules[types.VdGt] = vv
	}
	if vv, ok := tagValue.Lookup(types.VdLt); ok {
		rules[types.VdLt] = vv
	}
	if vv, ok := tagValue.Lookup(types.VdLt); ok {
		rules[types.VdLt] = vv
	}
	if vv, ok := tagValue.Lookup(types.VdRequired); ok {
		rules[types.VdRequired] = vv
	}
	if vv, ok := tagValue.Lookup(types.VdProtocols); ok {
		rules[types.VdProtocols] = vv
	}
	return rules
}

func (f *InputController) iterateOverTypes(fieldIndex int, currentFieldType, fieldKeyName string,
	tagValue *reflect.StructTag, configValue *reflect.Value, isStruct bool) {
	var i = fieldIndex
	var validationsRules = f.getValidationTags(tagValue)
	var mainErr error
	var mainReason = ReasonNotFound
	if _, ok := validationsRules[types.VdRequired]; ok {
		if !f.exists(fieldKeyName, tagValue) {
			mainErr = fmt.Errorf("field %s is required and must exist", fieldKeyName)
		}
	}

	if mainErr == nil {
		switch currentFieldType {
		case "bool":
			v, err := f.resolveBoolean(fieldKeyName, tagValue)
			if err != nil {
				mainErr = err
				break
			}
			configValue.Elem().Field(i).SetBool(v)
		case "string", "url.URL", "*url.URL":
			v, skipped, err := f.resolveString(fieldKeyName, tagValue)
			if err != nil {
				mainErr = err
				break
			} else if skipped {
				mainErr = nil
				break
			}
			v, usedSyntax, vdErr := f.CheckStringPreProcessors(v, validationsRules)
			if vdErr != nil {
				mainReason = ReasonValidation
				mainErr = vdErr
				break
			}
			if strings.Contains(currentFieldType, "url.URL") && usedSyntax == SyntaxURLParse {
				up, err := url.Parse(v)
				if err != nil {
					mainErr = err
					mainReason = ReasonValidation
				}
				refV := reflect.ValueOf(up)
				configValue.Elem().Field(i).Set(refV)
			} else {
				err = types.ValidateStrings(v, validationsRules)
				if err != nil {
					mainErr = err
					break
				}
				configValue.Elem().Field(i).SetString(v)
			}

		case "time.Duration":
			v, skipped, err := f.resolveString(fieldKeyName, tagValue)
			if err != nil {
				mainErr = err
				break
			} else if skipped {
				mainErr = nil
				break
			}
			dur := f.CheckTimeDurationPreprocessor(v)
			mainErr = types.ValidateTimeDurations(dur, validationsRules)
			if mainErr != nil {
				mainReason = ReasonValidation
				break
			}
			refV := reflect.ValueOf(dur)
			configValue.Elem().Field(i).Set(refV)
		case "int", "int8", "int16", "int32", "int64":
			s, skipped, err := f.resolveString(fieldKeyName, tagValue)
			if err == nil && f.IsDataSize(s) {
				data, err := f.CheckDataSize(s)
				if err != nil {
					mainReason = ReasonNotFound
					break
				}
				configValue.Elem().Field(i).SetInt(int64(data))
				break
			} else if skipped {
				mainErr = nil
				break
			}
			v, err := f.resolveNumber(fieldKeyName, tagValue)
			if err != nil {
				mainErr = err
				mainReason = ReasonNotFound
				break
			}
			err = types.ValidateNumbers[int64](v, validationsRules)
			if err != nil {
				mainReason = ReasonValidation
				break
			}
			configValue.Elem().Field(i).SetInt(int64(v))
		case "float16", "float32", "float64":
			v, err := f.resolveNumber(fieldKeyName, tagValue)
			if err != nil {
				mainReason = ReasonNotFound
				break
			}
			err = types.ValidateNumbers[float64](v, validationsRules)
			if err != nil {
				mainReason = ReasonNotFound
				break
			}
			configValue.Elem().Field(i).SetFloat(v)
		case "uint", "uint8", "uint16", "uint32", "uint64":
			s, skipped, err := f.resolveString(fieldKeyName, tagValue)
			if err == nil && f.IsDataSize(s) {
				data, err := f.CheckDataSize(s)
				if err != nil {
					mainReason = ReasonNotFound
					break
				}
				configValue.Elem().Field(i).SetUint(uint64(data))
				break
			} else if skipped {
				mainErr = nil
				break
			}
			v, err := f.resolveNumber(fieldKeyName, tagValue)
			if err != nil {
				mainReason = ReasonNotFound
				break
			}
			err = types.ValidateNumbers[uint64](v, validationsRules)
			if err != nil {
				mainReason = ReasonNotFound
				break
			}
			configValue.Elem().Field(i).SetUint(uint64(v))
		case "[]int", "[]string", "[]float64":
			if originalName, matched, index := CheckNameIsArrayAndGetIndex(fieldKeyName); matched && index > 0 &&
				originalName != "" {
				var objVal reflect.Value
				if configValue.Elem().Field(i).IsNil() {
					objVal = reflect.New(configValue.Elem().Field(i).Type())
				} else {
					objVal = configValue.Elem().Field(i)
				}
				if currentFieldType == "[]int" {
					v, err := f.resolveNumber(fieldKeyName, tagValue)
					if err != nil {
						if err != nil {
							mainReason = ReasonNotFound
							break
						}
					}
					objVal.Index(index).SetInt(int64(v))
				} else if currentFieldType == "[]string" {
					v, skipped, err := f.resolveString(fieldKeyName, tagValue)
					if err != nil {
						if err != nil {
							mainReason = ReasonNotFound
							break
						}
					} else if skipped {
						mainErr = nil
						break
					}
					objVal.Index(index).SetString(v)
				} else if currentFieldType == "[]float64" {
					v, err := f.resolveNumber(fieldKeyName, tagValue)
					if err != nil {
						mainReason = ReasonNotFound
						break
					}
					objVal.Index(index).SetFloat(v)
				}

				configValue.Elem().Field(i).Set(objVal)
			}
			v, skipped, err := f.resolveString(fieldKeyName, tagValue)
			if err != nil {
				mainReason = ReasonNotFound
				break
			} else if skipped {
				mainErr = nil
				break
			}
			var values any
			if currentFieldType == "[]int" {
				values, err = f.CheckIntArray(v)
			} else if currentFieldType == "[]string" {
				values, err = f.CheckStrArray(v)
			} else if currentFieldType == "[]float64" {
				values, err = f.CheckFloatArray(v)
			}
			if values != nil {
				obj := reflect.ValueOf(values)
				configValue.Elem().Field(i).Set(obj)
			}
		default:
			if isStruct {
				v, skipped, err := f.resolveString(fieldKeyName, tagValue)
				if err != nil {
					mainReason = ReasonNotFound
					break
				} else if skipped {
					mainErr = nil
					break
				}
				obj := reflect.New(configValue.Elem().Field(i).Type())
				err = f.CheckObjectPreprocessor(v, obj.Interface())
				if err != nil {
					mainErr = err
					mainReason = ReasonValidation
				}
				configValue.Elem().Field(i).Set(obj.Elem())
			}
		}
	}

	if mainErr != nil {
		f.HandleError(mainErr, fieldKeyName, mainReason, validationsRules)
	}
}

// HandleError
// checks to see if a field has a RequiredError tag, and if so, insert it into critical errors section
// it also inserts the error into fields error, as well.
func (f *InputController) HandleError(err error, key string, reason string, rules map[string]string) {
	if err != nil {
		f.insertError(key, err, reason)
	}
}

func (f *InputController) RequiredError(key string, rules map[string]string) (err error, validationType string) {
	if _, ok := rules[types.VdRequired]; ok {
		return fmt.Errorf("key %s is required but it is not found in any of the configs or the value is not properly formatted", key), ReasonRequired
	}
	return nil, ""
}

func (f *InputController) insertError(key string, err error, validationType string) {
	if _, ok := f.validationErrors[validationType]; !ok {
		f.validationErrors[validationType] = make(map[string]string, 0)
	}
	f.validationErrors[validationType][key] = err.Error()
}

func (f *InputController) GetAllErrors() []string {
	var validationErrs []string
	if len(f.validationErrors) > 0 {
		validationErrs = make([]string, 0)
		for _, v := range f.validationErrors {
			for _, vv := range v {
				validationErrs = append(validationErrs, vv)
			}
		}
		if len(validationErrs) == 0 {
			validationErrs = nil
		}
	}
	return validationErrs
}

// MustSkip checks to see if current input can process current struct's field or not
// this is due to the feature that allows you to use 'skips' tag for a field
// and use a comma separated list of inputs which you DO NOT want to check/process
// your field
func (f *InputController) MustSkip(currInputName string, tag *reflect.StructTag) bool {
	t, ok := tag.Lookup("skips")
	if ok {
		names := strings.Split(t, ",")
		if names != nil && len(names) > 0 {
			for _, v := range names {
				if v == currInputName {
					return true
				}
			}
		}
	}
	return false
}

func (f *InputController) exists(key string, field *reflect.StructTag) bool {
	for _, v := range f.input {
		if f.MustSkip(v.GetInputName(), field) {
			continue
		}
		if v.Has(key) {
			return true
		}
	}
	return false
}

// Resolves a string through looking up in registered sources
// returns
// string value [if found],
// bool skipped [true if corresponding struct's tag has a skips="inputName" entry]
// error err [if any error has been encountered during lookup]
func (f *InputController) resolveString(key string, field *reflect.StructTag) (string, bool, error) {
	var allSkipped = true
	for _, v := range f.input {
		if f.MustSkip(v.GetInputName(), field) {
			continue
		} else if vv, err := v.GetString(key); err == nil {
			return vv, false, nil
		}
		allSkipped = false
	}
	if v := f.resolveDefault(field); v != "" {
		// if a tag has default value, we return skipped as false
		return v, false, nil
	}
	if !allSkipped {
		return "", false, fmt.Errorf("key '%s' not found in any of the config-input", key)
	}
	return "", allSkipped, nil
}

func (f *InputController) resolveNumber(key string, field *reflect.StructTag) (float64, error) {
	for _, v := range f.input {
		if f.MustSkip(v.GetInputName(), field) {
			continue
		}
		if vv, err := v.GetNumber(key); err == nil {
			return vv, nil
		}
	}
	if v := f.resolveDefault(field); v != "" {
		if vv, err := strconv.ParseFloat(v, 64); err != nil {
			return 0, fmt.Errorf("field %s has default but the value cannot be validated as number, got error: %s", key, err.Error())
		} else {
			return vv, nil
		}
	}
	return 0, fmt.Errorf("key '%s' not found in any of the config-input", key)
}

func (f *InputController) resolveBoolean(key string, field *reflect.StructTag) (bool, error) {
	for _, v := range f.input {
		if f.MustSkip(v.GetInputName(), field) {
			continue
		}
		if vv, err := v.GetBoolean(key); err == nil {
			return vv, nil
		}
	}
	if v := f.resolveDefault(field); v != "" {
		if vv, err := strconv.ParseBool(v); err != nil {
			return false, fmt.Errorf("field %s has default but the value cannot be validated as boolean, got error: %s", key, err.Error())
		} else {
			return vv, nil
		}
	}
	return false, fmt.Errorf("key '%s' not found in any of the config-input", key)
}

func (f *InputController) resolveDefault(v *reflect.StructTag) string {
	if v == nil {
		return ""
	}
	res, ok := v.Lookup(f.defaultTagName)
	if !ok {
		return ""
	}
	return res
}

func (f *InputController) Int(name string) int {
	f.lock.RLock()
	defer f.lock.RUnlock()
	if v, ok := f.internalCacheInt[name]; ok {
		return v
	}
	return 0
}

func (f *InputController) Bool(name string) bool {
	f.lock.RLock()
	defer f.lock.RUnlock()
	if v, ok := f.internalCacheBoolean[name]; ok {
		return v
	}
	return false
}

func (f *InputController) String(name string) string {
	f.lock.RLock()
	defer f.lock.RUnlock()
	if v, ok := f.internalCacheString[name]; ok {
		return v
	}
	return ""
}

func (f *InputController) Float(name string) float64 {
	f.lock.RLock()
	defer f.lock.RUnlock()
	if v, ok := f.internalCacheFloat[name]; ok {
		return v
	}
	return 0
}
func (f *InputController) UnInt(name string) uint64 {
	f.lock.RLock()
	defer f.lock.RUnlock()
	if v, ok := f.internalCacheUnInt[name]; ok {
		return v
	}
	return 0
}

func (f *InputController) Count() int {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return len(f.internalCacheInt) + len(f.internalCacheString) + len(f.internalCacheBoolean) + len(f.internalCacheFloat) + len(f.internalCacheUnInt)
}
func (f *InputController) Reload() {
	for _, v := range f.input {
		if v.CanRefresh() {
			if err := v.Reload(); err != nil {
				fmt.Printf("failed to reload the config-source: %s", err.Error())
			}
		}
	}
}

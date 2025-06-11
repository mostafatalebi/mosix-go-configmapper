package configmapper

import (
	"encoding/base64"
	"mosix-go-configmapper/inputs"
	"mosix-go-configmapper/types"
	"net/url"
	"testing"
	"time"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
)

func TestFeatureHub_ShouldReturnError(t *testing.T) {
	type SampleConfig struct {
		Host string
	}
	var cnf = &SampleConfig{}

	inputMock := inputs.NewInputMock()

	inp := NewInputController("env", "default", inputMock)

	var err = inp.FetchKeysAndMapThem(cnf)
	assert.NoError(t, err)
}

func TestFeatureHub_ShouldSetValue_AllTypes(t *testing.T) {
	type SampleConfig struct {
		Host      string  `name:"APP_HOST"`
		Port      int     `name:"APP_PORT"`
		Precision float32 `name:"APP_PRECISION"`
		Debug     bool    `name:"APP_DEBUG"`
		Threshold uint8   `name:"APP_THRESHOLD"`
		NotFound  int
	}
	var cnf = &SampleConfig{}
	inputMock := inputs.NewInputMock()
	var randVal = xid.New().String()
	inputMock.KeysStr["APP_HOST"] = randVal
	inputMock.KeysNumber["APP_PORT"] = 8000
	inputMock.KeysNumber["APP_PRECISION"] = 0.0_000_008
	inputMock.KeysNumber["APP_THRESHOLD"] = 4
	inputMock.KeysBool["APP_DEBUG"] = true

	inp := NewInputController("name", "default", inputMock)
	var err = inp.FetchKeysAndMapThem(cnf)
	assert.NoError(t, err)
	assert.Equal(t, randVal, cnf.Host)
	assert.Equal(t, 8000, cnf.Port)
	assert.Equal(t, float32(0.0_000_008), cnf.Precision)
	assert.Equal(t, uint8(4), cnf.Threshold)
	assert.True(t, cnf.Debug)
}

func TestFeatureHub_ShouldSetValue_AllTypes_SeveralInputs(t *testing.T) {
	type SampleConfig struct {
		Host      string  `name:"APP_HOST"`
		Port      int     `name:"APP_PORT"`
		Precision float32 `name:"APP_PRECISION"`
		Debug     bool    `name:"APP_DEBUG"`
		Threshold uint8   `name:"APP_THRESHOLD"`
		NotFound  int
	}
	var cnf = &SampleConfig{}
	input1Mock := inputs.NewInputMock()
	input2Mock := inputs.NewInputMock()

	input1Mock.KeysStr["APP_HOST"] = "FirstInput_Host"

	input2Mock.KeysStr["APP_HOST"] = "SecondInput_Host"
	input2Mock.KeysNumber["APP_PORT"] = 8000
	input2Mock.KeysNumber["APP_PRECISION"] = 0.0_000_008
	input2Mock.KeysNumber["APP_THRESHOLD"] = 4
	input2Mock.KeysBool["APP_DEBUG"] = true

	inp := NewInputController("name", "default", input1Mock, input2Mock)

	var err = inp.FetchKeysAndMapThem(cnf)
	assert.NoError(t, err)
	assert.Equal(t, "FirstInput_Host", cnf.Host)
	assert.Equal(t, 8000, cnf.Port)
	assert.Equal(t, float32(0.0_000_008), cnf.Precision)
	assert.Equal(t, uint8(4), cnf.Threshold)
	assert.True(t, cnf.Debug)
}

func TestFeatureHub_ValidationError(t *testing.T) {
	type SampleConfig struct {
		Host      string  `name:"APP_HOST" required:""`
		Port      int     `name:"APP_PORT"`
		Precision float32 `name:"APP_PRECISION"`
		Debug     bool    `name:"APP_DEBUG"`
		Threshold uint8   `name:"APP_THRESHOLD"`
		NotFound  int
	}
	var cnf = &SampleConfig{}
	input1Mock := inputs.NewInputMock()
	input2Mock := inputs.NewInputMock()
	input2Mock.KeysNumber["APP_PORT"] = 8000
	input2Mock.KeysNumber["APP_PRECISION"] = 0.0_000_008
	input2Mock.KeysNumber["APP_THRESHOLD"] = 4
	input2Mock.KeysBool["APP_DEBUG"] = true

	inp := NewInputController("name", "default", input1Mock, input2Mock)

	var err = inp.FetchKeysAndMapThem(cnf)
	assert.Nil(t, err)
	if len(inp.GetAllErrors()) > 0 {
		assert.Equal(t, 1, len(inp.GetAllErrors()))
		assert.Equal(t, "field APP_HOST is required and must exist", inp.GetAllErrors()[0])
	}
}

type TestJson struct {
	Name     string `json:"name"`
	LastName string `json:"lastName"`
}

func TestFeatureHub_CoveringAllTypes(t *testing.T) {
	type SampleConfig struct {
		Host        string  `name:"APP_HOST" required:""`
		Name        string  `name:"APP_NAME" default:"SampleName"`
		Port        int     `name:"APP_PORT" default:"8080"`
		Precision   float32 `name:"APP_PRECISION"`
		Debug       bool    `name:"APP_DEBUG"`
		Threshold   uint8   `name:"APP_THRESHOLD"`
		CannotBeSet string  `name:"APP_CANNOT_BE_SET" skips:"mock"`

		SampleBase64Decode string        `name:"SAMPLE_BASE64_DECODE"`
		SampleBase64Encode string        `name:"SAMPLE_BASE64_ENCODE"`
		SampleDuration     time.Duration `name:"SAMPLE_DURATION"`

		SampleJSON                     *TestJson   `name:"SAMPLE_JSON"`
		SampleArrayInt                 []int       `name:"SAMPLE_ARRAY_INT"`
		SampleArrayString              []string    `name:"SAMPLE_ARRAY_STRING"`
		SampleArrayFloat               []float64   `name:"SAMPLE_ARRAY_FLOAT"`
		SampleDataSize                 int64       `name:"SAMPLE_DATA_SIZE"`
		SampleDataSizeUint             uint32      `name:"SAMPLE_DATA_SIZE_UINT"`
		SampleDataSizeInt              int         `name:"SAMPLE_DATA_SIZE_INT"`
		SampleURLEncoded               string      `name:"SAMPLE_URL_ENCODED"`
		SampleURLDecoded               string      `name:"SAMPLE_URL_DECODED"`
		SampleNumberUnsignedRange      uint64      `name:"SAMPLE_NUMBER_UNSIGNED_RANGE" range:"-100..100"`
		SampleNumberUnsignedRangeError uint64      `name:"SAMPLE_NUMBER_UNSIGNED_RANGE_ERROR" range:"-100..100"`
		SampleNumberSet                int64       `name:"SAMPLE_NUMBER_SET" set:"1,3,11"`
		SampleNumberGT                 int64       `name:"SAMPLE_NUMBER_GT" greaterThan:"100"`
		SampleNumberLT                 int64       `name:"SAMPLE_NUMBER_LT" lessThan:"0"`
		SampleMap                      map[int]int `name:"SAMPLE_MAP"`
	}
	var rn = xid.New().String()
	var cnf = &SampleConfig{}
	cnf.CannotBeSet = "Cannot be overridden by Input, due to 'skips' tag"
	input1Mock := inputs.NewInputMock()
	input2Mock := inputs.NewInputMock()
	var randomText = xid.New().String()
	var randomEncoded = base64.StdEncoding.EncodeToString([]byte(randomText))
	input1Mock.KeysStr["SAMPLE_BASE64_DECODE"] = "base64.decode::" + randomEncoded
	input1Mock.KeysStr["SAMPLE_BASE64_ENCODE"] = "base64.encode::" + randomText
	input1Mock.KeysStr["SAMPLE_DURATION"] = "time.duration::300s"
	input1Mock.KeysStr["SAMPLE_ARRAY_INT"] = "array.int::50,20, 40, 33,18"
	input1Mock.KeysStr["SAMPLE_MAP"] = "json.object::{\"0\": 5, \"30\": 10, \"60\": 30, \"300\": 60}"
	input1Mock.KeysStr["SAMPLE_ARRAY_STRING"] = "array.string::foo,bar,john,sample"
	input1Mock.KeysStr["SAMPLE_ARRAY_FLOAT"] = "array.float::0.002,10.0, 100"
	input1Mock.KeysStr["SAMPLE_JSON"] = "json.object::{\"name\":\"foo\",\"lastName\":\"bar\"}"
	input1Mock.KeysStr["APP_HOST"] = "example.com" + rn
	input1Mock.KeysStr["SAMPLE_INT_ARRAY"] = "array.int::3,7,1,898984"
	input1Mock.KeysStr["SAMPLE_DATA_SIZE"] = "data.size::20kb"
	input1Mock.KeysStr["SAMPLE_DATA_SIZE_UINT"] = "data.size::10kb"
	input1Mock.KeysStr["SAMPLE_DATA_SIZE_INT"] = "data.size::1kb"
	input1Mock.KeysStr["SAMPLE_URL_ENCODED"] = "url.encode::value="
	input1Mock.KeysStr["SAMPLE_URL_DECODED"] = "url.decode::value%3D"
	input1Mock.KeysNumber["SAMPLE_NUMBER_UNSIGNED_RANGE"] = 2
	input1Mock.KeysNumber["SAMPLE_NUMBER_UNSIGNED_RANGE_ERROR"] = -2
	input1Mock.KeysNumber["SAMPLE_NUMBER_SET"] = 3
	input1Mock.KeysNumber["SAMPLE_NUMBER_GT"] = 64
	input1Mock.KeysNumber["SAMPLE_NUMBER_LT"] = -100
	input2Mock.KeysNumber["APP_PRECISION"] = 0.0_000_008
	input2Mock.KeysNumber["APP_THRESHOLD"] = 4
	input2Mock.KeysStr["APP_CANNOT_BE_SET"] = "A never mapped value"
	input2Mock.KeysBool["APP_DEBUG"] = true

	inp := NewInputController("name", "default", input1Mock, input2Mock)
	inp.TogglePreprocessors(true)
	err := inp.FetchKeysAndMapThem(cnf)
	assert.Equal(t, "example.com"+rn, cnf.Host)
	assert.Equal(t, "SampleName", cnf.Name)
	assert.Equal(t, 8080, cnf.Port)
	assert.Equal(t, float32(0.0_000_008), cnf.Precision)
	assert.Equal(t, uint8(4), cnf.Threshold)
	assert.Equal(t, randomText, cnf.SampleBase64Decode)
	assert.Equal(t, randomEncoded, cnf.SampleBase64Encode)
	assert.Equal(t, 300*time.Second, cnf.SampleDuration)
	assert.NotNil(t, cnf.SampleArrayInt)
	assert.NotEmpty(t, cnf.SampleArrayInt)
	assert.NotNil(t, cnf.SampleArrayString)
	assert.NotEmpty(t, cnf.SampleArrayString)
	assert.NotNil(t, cnf.SampleArrayFloat)
	assert.NotEmpty(t, cnf.SampleArrayFloat)
	assert.Equal(t, int64(20*1024), cnf.SampleDataSize)
	assert.Equal(t, uint32(10*1024), cnf.SampleDataSizeUint)
	assert.Equal(t, int(1*1024), cnf.SampleDataSizeInt)
	assert.NotNil(t, cnf.SampleJSON)
	if cnf.SampleJSON == nil {
		t.Fail()
		return
	}
	assert.Equal(t, uint64(2), cnf.SampleNumberUnsignedRange)
	assert.Zero(t, cnf.SampleNumberUnsignedRangeError)
	assert.Equal(t, "foo", cnf.SampleJSON.Name)
	assert.Equal(t, "bar", cnf.SampleJSON.LastName)

	assert.NotNil(t, cnf.SampleMap)
	if cnf.SampleMap == nil {
		t.Fail()
		return
	}

	assert.Equal(t, 5, cnf.SampleMap[0])
	assert.Equal(t, 10, cnf.SampleMap[30])

	assert.Equal(t, "Cannot be overridden by Input, due to 'skips' tag", cnf.CannotBeSet)
	assert.Nil(t, err)
	if inp.GetAllErrors() != nil {
		assert.Equal(t, "there are critical errors", inp.GetAllErrors())
	}
	assert.Equal(t, "value%3D", cnf.SampleURLEncoded)
	assert.Equal(t, "value=", cnf.SampleURLDecoded)

}

func TestURLInputType(t *testing.T) {
	type SampleConfig struct {
		URLHttp                   string        `name:"URL_HTTP" required:""`
		URLFtps                   string        `name:"URL_FTPS" required:"" protocols:"ftps"`
		URLHttpsUrlPkg            *url.URL      `name:"URL_HTTPS_URL" required:""`
		TimeDurWithValidation     time.Duration `name:"TIME_DUR_WITH_VALIDATION" required:"" range:"2s..1m"`
		TimeDurWithValidationErr  time.Duration `name:"TIME_DUR_WITH_VALIDATION_ERR" required:"" range:"2s..1m"`
		TimeDurWithValidationErr2 time.Duration `name:"TIME_DUR_WITH_VALIDATION_ERR_2" required:"" greaterThan:"1000ms"`
		TimeDurWithValidationErr3 time.Duration `name:"TIME_DUR_WITH_VALIDATION_ERR_3" required:"" lessThan:"20ms"`
	}
	var cnf = &SampleConfig{}
	input1Mock := inputs.NewInputMock()
	input2Mock := inputs.NewInputMock()
	input1Mock.KeysStr["URL_HTTP"] = "url::http://example.com"
	input1Mock.KeysStr["URL_FTPS"] = "url::https://example.com"
	input1Mock.KeysStr["URL_HTTPS_URL"] = "url::https://example.com/with/url/pkg"
	input1Mock.KeysStr["TIME_DUR_WITH_VALIDATION"] = "time.duration::25s"
	input1Mock.KeysStr["TIME_DUR_WITH_VALIDATION_ERR"] = "time.duration::1ms"
	input1Mock.KeysStr["TIME_DUR_WITH_VALIDATION_ERR_2"] = "time.duration::999ms"
	input1Mock.KeysStr["TIME_DUR_WITH_VALIDATION_ERR_3"] = "time.duration::19ms"

	inp := NewInputController("name", "default", input1Mock, input2Mock)
	inp.TogglePreprocessors(true)
	err := inp.FetchKeysAndMapThem(cnf)
	assert.NoError(t, err)
	assert.NotEmpty(t, inp.GetAllErrors())
	assert.Len(t, inp.GetAllErrors(), 3)

	assert.Equal(t, "http://example.com", cnf.URLHttp)
	assert.Empty(t, cnf.URLFtps)
	assert.NotNil(t, cnf.URLHttpsUrlPkg)

	if cnf.URLHttpsUrlPkg != nil {
		assert.Equal(t, "/with/url/pkg", cnf.URLHttpsUrlPkg.Path)
	}
	assert.NotZero(t, cnf.TimeDurWithValidation)
	assert.Zero(t, cnf.TimeDurWithValidationErr)
	assert.NotEmpty(t, inp.GetValidationError("TIME_DUR_WITH_VALIDATION_ERR_2", ReasonValidation))
	assert.Equal(t, time.Millisecond*19, cnf.TimeDurWithValidationErr3)

	assert.Equal(t, types.VdProtocols+": url https://example.com is not among allowed protocols ftps", inp.GetAllErrors()[0])
}

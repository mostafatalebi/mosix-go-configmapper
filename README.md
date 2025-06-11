![Version Status](https://img.shields.io/badge/version-v0.0.1--active-brightgreen
)
### Mosix Go Config Mapper
A utility which allows mapping config keys to a go struct. It has a number of helpers
which faciliates handling different types of values as well. 


### Sources
Currently this package allows you to read and map from either `OsEnv` or 
`FeatureHub`. You can easily and quickly implement your own input source
and use it with the mapped (for example FileSource).

**If you use multiple input sources**
When passing an array of inputs, they are checked
in their respective order. If a value be found, the search stops and
remaining input will not be checked.

### Installation
```shell
go get -u github.com/mostafatalebi/mosix-go-configmapper
```
### Basics
You have the following configs:
```shell
DB_USER=user
DB_PASS=pass
DB_HOST=127.0.0.1
```

You need to have the following `struct` defined in your app:
```golang
struct Config {
    DbUser string `name:"DB_USER"`
    DbPass string `name:"DB_PASS"`
    DbHost string `name:"DB_HOST"`
}
```
You can also set a default value using `default` tag name.
```golang
DbHost string `name:"DB_HOST" default:"SampleName"`
```

And you can use different validation rules:
```golang
DbConnTimeout time.Duration `name:"DB_CONN_TIMEOUT" default:"time.duration::10ms" greaterThan:"1ms" lessThan:"1s" ` // gt and lt validations
DbReadTimeout time.Duration `name:"DB_READ_TIMEOUT" default:"time.duration::10ms" range:"1ms..1000ms" ` // range with time.Duration 
MaxConnCount int `name:"MAX_CONN_COUNT" range:"2..10"` // range x..y 
DbUser string `name:"DB_USER" required=""` // required
BackoffMultiplier int `name:"BO_MULTIPLIER" default:"25" set:"0,25,50,75,100"` // limit values to a set only
StatsURL string `name:"STATS_URL" url_protocol:"https"` // allows URLs with https scheme only
```

### Usage
The following shows you how to use this package and do error handling.
```golang
osEnvInput := NewOsEnv()
configInputs = append(configInputs, osEnvInput)

var inputController = NewInputController("tag", "default", configInputs...)

// this enables processing special syntaxes 
inputController.TogglePreprocessors(true)

// maps the values to the struct
err := inputController.FetchKeysAndMapThem(c)
if err != nil {
    return err
}

// checking on errors
var errorsList = GetCriticalErrors()
if len(errorsList) > 0 {
    return errors.New(strings.Join(errorsList, ";"))
}
```

### Special Values
You can pass several values with the config being mapped or preprocessed. 
To use them, simply prepend the value of the config with the one of the
belows. For example, if you want to pass an array of integer, in your
original input source, e.g. in your OS ENV, do this:
USER_IDS=array.int::20,21,25,57

Here are is the list of available syntaxes.
```golang
const (
    // an array of integers
	SyntaxArrayInt       Syntax = "array.int::"

    // an array of float64
	SyntaxArrayFloat     Syntax = "array.float::"

    // an array of string
	SyntaxArrayStr       Syntax = "array.string::"

    // a JSON formatted text string to be parsed by the config and mapped to 
    // value of the tag
	SyntaxJsonObject     Syntax = "json.object::"

    // encodes the value before mapping it to the field
	SyntaxBase64Encoding Syntax = "base64.encode::"

    // decodes the value before mapping it to the field
	SyntaxBase64Decoding Syntax = "base64.decode::"

    // treats the value as a time.Duration. It can map the result
    // either as a time.Duration or an integer (this is specified) by
    // the type of the struct's field
	SyntaxTimeDuration   Syntax = "time.duration::"

    // data sizes. Like "20KB"
	SyntaxDataSize       Syntax = "data.size::"

    // encodes the URL before mapping it to the field
	SyntaxURLEncode      Syntax = "url.encode::"

    // decodes the URL before mapping it to the field
	SyntaxURLDecode      Syntax = "url.decode::"

    // expects the value to be a URL. It can map it both to
    // a field with type string or url.URL
	SyntaxURLParse       Syntax = "url::"
)
```

You can refer to input_test file to see more examples.


### Skipping
You might have several sources. For example you might have a file source, an OS ENV source and a FeatureHub source. You
might want your credentials to be read only from the file, you can then do:

```golang
DbPassword string `name:"DB_PASSWORD" skips:"feature-hub"
```
This way, the config mapper never reads any value for this field
from the mentioned source name.
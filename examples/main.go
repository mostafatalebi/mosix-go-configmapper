package main

import (
	"fmt"
	configmapper "mosix-go-configmapper"
	"mosix-go-configmapper/inputs"
	"os"
	"time"
)

type Params struct {
	Id       int    `json:"id"`
	ReadOnly bool   `json:"readOnly"`
	Salt     string `json:"salt"`
}

type Config struct {
	User       string        `name:"USER" required:""`
	Pass       string        `name:"PASS" required:""`
	Port       uint          `name:"PORT" range:"8000..9000"`
	Params     *Params       `name:"PARAMS"`
	Multiplier int           `name:"MULTIPLIER" set:"0,25,50,75,100"`
	Timeout    time.Duration `name:"TIMEOUT" default:"time.duration::20ms"`
}

func main() {
	os.Setenv("USER", "Joe")
	os.Setenv("PASS", "pass1234")
	os.Setenv("PORT", "8080")
	os.Setenv("PARAMS", `json.object::{ "id" : 100, "readOnly":true, "salt" : "abc"}`)
	os.Setenv("MULTIPLIER", "int.array::0,25,50,75,100")

	var exampleConfig = &Config{}

	var source1 = inputs.NewOsEnv()

	var controller = configmapper.NewInputController("name", "default", source1)
	controller.TogglePreprocessors(true)
	controller.FetchKeysAndMapThem(exampleConfig)
	var errs = controller.GetAllErrors()
	if len(errs) > 0 {
		for _, v := range errs {
			fmt.Println(v)
		}
		os.Exit(1)
	}

	if exampleConfig.User != "Joe" {
		panic("assertion failed for exampleConfig.User")
	}
	if exampleConfig.Pass != "pass1234" {
		panic("assertion failed for exampleConfig.Pass")
	}
	if exampleConfig.Port != 8080 {
		panic("assertion failed for exampleConfig.Port")
	}
	if exampleConfig.Params == nil {
		panic("assertion failed for exampleConfig.Params")
	} else {
		if exampleConfig.Params.Id != 100 {
			panic("assertion failed for exampleConfig.Params.Id")
		} else if !exampleConfig.Params.ReadOnly {
			panic("assertion failed for exampleConfig.Params.ReadOnly")
		} else if exampleConfig.Params.Salt != "abc" {
			panic("assertion failed for exampleConfig.Params.Salt")
		}
	}
	if exampleConfig.Timeout != 20*time.Millisecond {
		panic("assertion failed for exampleConfig.Timeout")
	}
	fmt.Println("all assertions passed successfully")
}

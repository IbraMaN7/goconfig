package goenv

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/h2oai/goconfig/structtag"
)

var (
	// Prefix is a string that would be placed at the beginning of the generated tags.
	Prefix string

	// Usage is the function that is called when an error occurs.
	Usage func()

	// PrintDefaultsOutput changes the default output help string
	PrintDefaultsOutput string
)

// Setup maps and variables
func Setup(tag string, tagDefault string, kebabCfgToSnakeEnv bool) {
	Usage = DefaultUsage

	structtag.Setup()
	structtag.Prefix = Prefix
	SetTag(tag)
	SetTagDefault(tagDefault)
	SetKebabCfgToSnakeEnv(kebabCfgToSnakeEnv)

	structtag.ParseMap[reflect.Int64] = reflectInt
	structtag.ParseMap[reflect.Int] = reflectInt
	structtag.ParseMap[reflect.Float64] = reflectFloat
	structtag.ParseMap[reflect.String] = reflectString
	structtag.ParseMap[reflect.Bool] = reflectBool
	structtag.ParseMap[reflect.Array] = reflectArray
	structtag.ParseMap[reflect.Slice] = reflectArray
}

// SetTag set a new tag
func SetTag(tag string) {
	structtag.Tag = tag
}

// SetTagDefault set a new TagDefault to return default values
func SetTagDefault(tag string) {
	structtag.TagDefault = tag
}

// SetKebabCfgToSnakeEnv set a new CfgToSnakeEnv to look for snakecase environment variables
func SetKebabCfgToSnakeEnv(cfgToSnakeEnv bool) {
	structtag.KebabCfgToSnakeEnv = cfgToSnakeEnv
}

// Parse configuration
func Parse(config interface{}) (err error) {
	err = structtag.Parse(config, "")
	return
}

func parseValue(datatype string, value *reflect.Value) (ret string, ok bool) {
	switch datatype {
	case "bool":
		ret = strconv.FormatBool(value.Bool())
		ok = true
	case "string":
		ret = value.String()
		ok = ret != ""
	case "int":
		ret = strconv.FormatInt(value.Int(), 10)
		ok = ret != "0"
	case "float64":
		ret = strconv.FormatFloat(value.Float(), 'f', -1, 64)
		ok = ret != "0"
	}
	return
}

func getNewValue(field *reflect.StructField, value *reflect.Value, tag string, datatype string) (ret string) {
	defaultValue := field.Tag.Get(structtag.TagDefault)

	tag = strings.ToUpper(tag)
	if structtag.KebabCfgToSnakeEnv {
		tag = strings.Replace(tag, "-", "_", -1)
	}

	sysvar := `$` + tag
	if runtime.GOOS == "windows" {
		sysvar = `%` + tag + `%`
	}

	output := fmt.Sprintf("  %v %v\n\n", sysvar, datatype)
	if defaultValue != "" {
		output = fmt.Sprintf("  %v %v\n\t(default %q)\n", sysvar, datatype, defaultValue)
	}
	PrintDefaultsOutput += output

	// get value from environment variable
	ret, ok := os.LookupEnv(tag)
	if ok {
		return
	}

	ret, ok = parseValue(datatype, value)
	if ok {
		return
	}

	// get value from default settings
	ret = defaultValue
	return
}

func reflectInt(field *reflect.StructField, value *reflect.Value, tag string) (err error) {
	newValue := getNewValue(field, value, tag, "int")
	if newValue == "" {
		return
	}
	var intNewValue int64
	intNewValue, err = strconv.ParseInt(newValue, 10, 64)
	if err != nil {
		return
	}
	value.SetInt(intNewValue)
	return
}

func reflectFloat(field *reflect.StructField, value *reflect.Value, tag string) (err error) {
	newValue := getNewValue(field, value, tag, "float64")
	if newValue == "" {
		return
	}
	var floatNewValue float64
	floatNewValue, err = strconv.ParseFloat(newValue, 64)
	if err != nil {
		return
	}
	value.SetFloat(floatNewValue)
	return
}

func reflectString(field *reflect.StructField, value *reflect.Value, tag string) (err error) {
	newValue := getNewValue(field, value, tag, "string")
	if newValue == "" {
		return
	}
	value.SetString(newValue)
	return
}

func reflectBool(field *reflect.StructField, value *reflect.Value, tag string) (err error) {
	newValue := getNewValue(field, value, tag, "bool")
	if newValue == "" {
		return
	}
	newValue = strings.ToLower(newValue)
	newBoolValue := newValue == "true" || newValue == "t" || newValue == "1"
	value.SetBool(newBoolValue)
	return
}

func reflectArray(field *reflect.StructField, value *reflect.Value, tag string) (err error) {
	// Not implemented due to serialization complexity
	return
}

// PrintDefaults print the default help
func PrintDefaults() {
	fmt.Println("Environment variables:")
	fmt.Println(PrintDefaultsOutput)
}

// DefaultUsage is assigned for Usage function by default
func DefaultUsage() {
	fmt.Println("Usage")
	PrintDefaults()
}

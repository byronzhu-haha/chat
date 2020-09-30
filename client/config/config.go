package config

import (
	"errors"
	"fmt"
	"github.com/byronzhu-haha/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
)

type Config struct {
	ServerAddr string `yaml:"ServerAddr" default:"127.0.0.1:4567"`
	Timeout    int    `yaml:"Timeout" default:"3"`
}

func (c *Config) String() string {
	return fmt.Sprintf("%+v", *c)
}

var DefaultConfig = &Config{}

func init() {
	loadConfig()
	f, err := os.Open("./client/bin/config.yaml")
	if err != nil {
		log.Warnf("open config yaml file failed, err: %+v", err)
		return
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		log.Errorf("read config failed, err: %+v", err)
		return
	}
	err = yaml.Unmarshal(buf, DefaultConfig)
	if err != nil {
		log.Errorf("read config failed, err: %+v", err)
		return
	}
	log.Infof("config; %+v", DefaultConfig)
}

func loadConfig() {
	t := reflect.TypeOf(DefaultConfig)
	v := reflect.ValueOf(DefaultConfig)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dv := field.Tag.Get("default")
		nv, err := coerce(dv, field.Type)
		if err != nil {
			log.Errorf("coerce failed, err: %+v", err)
			continue
		}
		v.Field(i).Set(nv)
	}
	log.Infof("default config: %+v", DefaultConfig)
}

func coerce(v interface{}, typ reflect.Type) (reflect.Value, error) {
	var err error
	if typ.Kind() == reflect.Ptr {
		return reflect.ValueOf(v), nil
	}
	switch typ.String() {
	case "string":
		v, err = coerceString(v)
	case "int", "int16", "int32", "int64":
		v, err = coerceInt64(v)
	default:
		v = nil
		err = fmt.Errorf("invalid type %s", typ.String())
	}
	return valueTypeCoerce(v, typ), err
}

func valueTypeCoerce(v interface{}, typ reflect.Type) reflect.Value {
	val := reflect.ValueOf(v)
	if reflect.TypeOf(v) == typ {
		return val
	}
	tval := reflect.New(typ).Elem()
	switch typ.String() {
	case "int", "int16", "int32", "int64":
		tval.SetInt(val.Int())
	default:
		tval.Set(val)
	}
	return tval
}

func coerceString(v interface{}) (string, error) {
	switch v := v.(type) {
	case string:
		return v, nil
	case int, int16, int32, int64, uint, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%f", v), nil
	}
	return fmt.Sprintf("%s", v), nil
}

func coerceInt64(v interface{}) (int64, error) {
	switch v := v.(type) {
	case string:
		return strconv.ParseInt(v, 10, 64)
	case int, int16, int32, int64:
		return reflect.ValueOf(v).Int(), nil
	case uint, uint16, uint32, uint64:
		return int64(reflect.ValueOf(v).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

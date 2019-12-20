// Package jsconfig reads configuration settings from one or more JSON files.
//
// Config files are JSON-formatted files that contain key+value pairs,
// where the value can be a string, number or bool,
// or an array (slice) or object composed of those primitives.
//
// Multiple config files can be read in succession, and later files will
// override or supplement kv pairs from earlier files.
//
// If a key occurs multiple times within a single file, later values will
// override the earlier ones.
//
//After initializing, use one of the AddRemoteStore functions of the global settings object "S", with "etcd" as the
//only supported provider, to use etcd for configuration.
//
//Pass an empty string for the "localPath" parameter to not have the remote settings written to disk.
//The AddSettingsProvider method allows users to use their own remote provider by implmenting the RemoteProvider interface.
package jsconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// Settings provides methods for looking up config values.
type Settings interface {
	FindString(key string) string
	FindNumber(key string) float64
	FindInt(key string) int
	FindBool(key string) bool
	FindSubSettings(key string) Settings
	FindStringSlice(key string) []string
	FindNumberSlice(key string) []float64
	FindBoolSlice(key string) []bool
	FindMap(key string) map[string]interface{}
	FindSubSettingsSlice(key string) []Settings
	FindInterfaceSlice(key string) []interface{}
	FindDuration(key string) time.Duration
	FindMilliseconds(key string) time.Duration
}

//RemoteProvider allows users to provide their own way to store settings.
type RemoteProvider interface {
	LookupSettings() (map[string]interface{}, error)
}

//Watcher interface has a single method to allow for key change notifications.
//This function should hang if no error is present, and returning indicates a setting
//has changed (call from go routine).  The client may call the same AddRemoteStore method to refresh settings,
//or restart if more desirable.
type Watcher interface {
	Watch() error
}

type settings map[string]interface{}

// S is the global Settings object
var S Settings

// GetSettings returns the global Settings object.
// This function is depreciated in favor of S.
func GetSettings() Settings {
	return S
}

// InitFromBytes initializes the global Settings from JSON strings.
func InitFromBytes(jsonBytes ...[]byte) error {
	s, err := newSettingsFromBytes(jsonBytes...)
	if err == nil {
		S = s
	}
	return err
}

// InitFromFiles initializes the global Settings from JSON files.
func InitFromFiles(files ...string) error {
	s, err := newSettingsFromFiles(files...)
	if err == nil {
		S = s
	}
	return err
}

const environmentVariable = "GO_ENV"

//FilesFromEnv uses the environment variabled "GO_ENV" to return
//standard config paths i.e. "./config/default.json" with the value of
//the second file returned matching the name of the environment variable.
//This can be called prior to InitFromFiles using ...
func FilesFromEnv() []string {
	env := os.Getenv(environmentVariable)
	if env == "" {
		return []string{"./config/default.json"}
	}
	return []string{"./config/default.json", fmt.Sprintf("./config/%s.json", env)}
}

// FindString looks up a string value from a Settings object.
func (s settings) FindString(key string) string {
	x, _ := findString(key, s)
	return x
}

// FindNumber looks up a float64 value from a Settings object.
func (s settings) FindNumber(key string) float64 {
	x, _ := findNumber(key, s)
	return x
}

// FindInt looks up a int value from a Settings object.
func (s settings) FindInt(key string) int {
	x, _ := findNumber(key, s)
	return int(x)
}

// FindBool looks up a bool value from a Settings object.
func (s settings) FindBool(key string) bool {
	x, _ := findBool(key, s)
	return x
}

// FindSubSettings looks up a subsection from a Settings object.
func (s settings) FindSubSettings(key string) Settings {
	i, found := findInterface(key, s)
	if !found {
		return nil
	}
	x := settings{key: i}
	return x
}

// FindMap looks up a map (ie, "SubSection") from a Settings object.
func (s settings) FindMap(key string) map[string]interface{} {
	m, _ := findMap(key, s)
	return m
}

// FindStringSlice looks up a slice of strings from a Settings object.
func (s settings) FindStringSlice(key string) []string {
	x, _ := findStringSlice(key, s)
	return x
}

// FindNumberSlice looks up a slice of float64s from a Settings object.
func (s settings) FindNumberSlice(key string) []float64 {
	x, _ := findNumberSlice(key, s)
	return x
}

// FindBoolSlice looks up a slice of bools from a Settings object.
func (s settings) FindBoolSlice(key string) []bool {
	x, _ := findBoolSlice(key, s)
	return x
}

// FindSubSettingsSlice looks up a slice of Settings objects from a Settings object.
func (s settings) FindSubSettingsSlice(key string) []Settings {
	subs, _ := findSubSettingsSlice(key, s)
	return subs
}

// FindInterfaceSlice looks up a slice of arbitrary objects from a Settings object.
// The caller is responsible for all type conversions.
func (s settings) FindInterfaceSlice(key string) []interface{} {
	x, _ := findInterfaceSlice(key, s)
	return x
}

func newSettingsFromBytes(jsonBytes ...[]byte) (Settings, error) {
	s, err := mapFromBytes(jsonBytes...)
	if err != nil {
		return nil, err
	}
	return settings(s), nil
}

func mapFromBytes(jsonBytes ...[]byte) (map[string]interface{}, error) {
	s, err := mergeStrings(jsonBytes...)
	if err != nil {
		return nil, err
	}
	return s, err
}

func newSettingsFromFiles(files ...string) (Settings, error) {
	data := make([][]byte, len(files))
	for i, fileName := range files {
		d, err := ioutil.ReadFile(fileName)
		if err != nil {
			return nil, err
		}
		data[i] = d
	}
	return newSettingsFromBytes(data...)
}

func find(key string, m map[string]interface{}) (interface{}, bool) {
	for k, o := range m {
		if k == key {
			return o, true
		}
		mapVal, isMap := o.(map[string]interface{})
		if isMap {
			oo, found := find(key, mapVal)
			if found {
				return oo, true
			}
		}
	}
	return nil, false
}

func findString(key string, m map[string]interface{}) (string, bool) {
	x, found := find(key, m)
	if found {
		val, isValType := x.(string)
		if isValType {
			return val, true
		}
	}
	return "", false
}

func findNumber(key string, m map[string]interface{}) (float64, bool) {
	x, found := find(key, m)
	if found {
		val, isValType := x.(float64)
		if isValType {
			return val, true
		}
	}
	return 0, false
}

func findBool(key string, m map[string]interface{}) (bool, bool) {
	x, found := find(key, m)
	if found {
		val, isValType := x.(bool)
		if isValType {
			return val, true
		}
	}
	return false, false
}

func findInterface(key string, m map[string]interface{}) (interface{}, bool) {
	x, found := find(key, m)
	if found {
		val, isValType := x.(interface{})
		if isValType {
			return val, true
		}
	}
	return nil, false
}

func findMap(key string, m map[string]interface{}) (map[string]interface{}, bool) {
	x, found := find(key, m)
	if found {
		val, isValType := x.(map[string]interface{})
		if isValType {
			return val, true
		}
	}
	return nil, false
}

func findStringSlice(key string, m map[string]interface{}) ([]string, bool) {
	x, found := find(key, m)
	if found {
		val, isValType := x.([]interface{})
		if isValType {
			if len(val) == 0 {
				return []string{}, true
			}
			_, isValSlice := val[0].(string)
			if !isValSlice {
				return nil, false
			}
			arr := make([]string, len(val))
			for i := range val {
				arr[i] = val[i].(string)
			}
			return arr, true
		}
	}
	return nil, false
}

func findNumberSlice(key string, m map[string]interface{}) ([]float64, bool) {
	x, found := find(key, m)
	if found {
		val, isValType := x.([]interface{})
		if isValType {
			if len(val) == 0 {
				return []float64{}, true
			}
			_, isValSlice := val[0].(float64)
			if !isValSlice {
				return nil, false
			}
			arr := make([]float64, len(val))
			for i := range val {
				arr[i] = val[i].(float64)
			}
			return arr, true
		}
	}
	return nil, false
}

func findBoolSlice(key string, m map[string]interface{}) ([]bool, bool) {
	x, found := find(key, m)
	if found {
		val, isValType := x.([]interface{})
		if isValType {
			if len(val) == 0 {
				return []bool{}, true
			}
			_, isValSlice := val[0].(bool)
			if !isValSlice {
				return nil, false
			}
			arr := make([]bool, len(val))
			for i := range val {
				arr[i] = val[i].(bool)
			}
			return arr, true
		}
	}
	return nil, false
}

func findSubSettingsSlice(key string, m map[string]interface{}) (s []Settings, found bool) {
	x, found := find(key, m)
	if !found {
		return nil, false
	}
	islice, ok := x.([]interface{})
	if !ok {
		return nil, false
	}
	for _, x = range islice {
		sub, ok := x.(map[string]interface{})
		if !ok {
			continue
		}
		s = append(s, settings(sub))
	}
	return s, true
}

func findInterfaceSlice(key string, m map[string]interface{}) ([]interface{}, bool) {
	x, found := find(key, m)
	if found {
		val, isValType := x.([]interface{})
		if isValType {
			arr := make([]interface{}, len(val))
			for i := range val {
				arr[i] = val[i]
			}
			return arr, true
		}
	}
	return nil, false
}

func mergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	mm := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			_, found := mm[k]
			if !found {
				mm[k] = v
			} else {
				switch mmType := mm[k].(type) {
				case map[string]interface{}:
					mapVal, isMap := v.(map[string]interface{})
					if !isMap {
						mm[k] = v
					} else {
						mm[k] = mergeMaps(mmType, mapVal)
					}
				default:
					mm[k] = v
				}
			}
		}
	}
	return mm
}

func mergeStrings(jsonStrings ...[]byte) (map[string]interface{}, error) {
	merged := make(map[string]interface{})
	for _, j := range jsonStrings {
		var f interface{}
		err := json.Unmarshal(j, &f)
		if err != nil {
			return nil, fmt.Errorf("Error unmarshalling '%s': %s", j, err)
		}
		m, isValid := f.(map[string]interface{})
		if !isValid {
			return nil, fmt.Errorf("Bad JSON string '%s'", j)
		}
		merged = mergeMaps(merged, m)
	}
	return merged, nil
}

func (s settings) FindDuration(key string) time.Duration {
	x := s.FindString(key)
	d, err := time.ParseDuration(x)
	if err == nil {
		return d
	}
	return 0
}

func (s settings) FindMilliseconds(key string) time.Duration {
	ms := s.FindInt(key)
	return time.Duration(ms) * time.Millisecond
}

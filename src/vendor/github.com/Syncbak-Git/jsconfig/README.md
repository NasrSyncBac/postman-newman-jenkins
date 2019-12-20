JSConfig
==========

A Go library for reading configuration settings from JSON files.

JSConfig works similarly to the node.js [config](https://github.com/lorenwest/node-config) library in that it merges configuration settings from a set of JSON files, with values from later files overriding or supplementing values from earlier files.

In the node.js world, the base settings come from a default.json file, while overrides are read from an environment-specific deployment file, found using the NODE-ENV environment variable. (node-config also has a runtime.json file that contains changes made at runtime; JSConfig does not support runtime changes).

JSConfig works somewhat differently: the application is responsible for determining which config files to use. JSConfig merges the files and handles lookup of config values.

Example files
default.json
```json
{
   "Setting1": "A value",
   "Setting2": 1234,
   "Setting3": true,
   "Setting4": [9, 8, 7, 6],
   "ASubSection": {
      "thisSettingChanges": "Default value"
   }
}
```

qa.json
```json
{
   "ASubSection": {
      "thisSettingChanges": "QA value"
   }
}
```

production.json
```json
{
   "ASubSection": {
      "thisSettingChanges": "Production value"
   }
}
```

application.go
```go
package main

import (
   "fmt"
   "github.com/Syncbak-Git/jsconfig"
)

func printSomeSettings() {
    s := jsconfig.GetSettings()
    setting2 := s.FindNumber("Setting2")
    subsection := s.FindSubSettings("ASubSection")
    variableSetting := subsection.FindString("thisSettingChanges")
    // you can also look up subsetting values directly, if the key is unique
    variableSetting2 := s.FindString("thisSettingChanges")
    fmt.Printf("%d, %s, %s\n", setting2, variableSetting, variableSetting2)
}

func main() {
    jsconfig.InitFromFiles("default.json")
    printSomeSettings() // 1234, Default value, Default value
    jsconfig.InitFromFiles("default.json", "qa.json")
    printSomeSettings() // 1234, QA value, QA value
    jsconfig.InitFromFiles("default.json", "production.json")
    printSomeSettings() // 1234, Production value, Production value
}
```


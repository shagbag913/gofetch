package main

import (
    "os/exec"
    "strings"
)

var propNames []string
var propKeys []string

func getProp(name string) string {
    // Cache prop values
    if len(propNames) == 0 || len(propKeys) == 0 {
        propList, err := exec.Command("getprop").Output()
        if err != nil {
            printDebug(err.Error())
        }

        propListSlice := strings.Split(string(propList), "\n")

        for i := range propListSlice {
            propLine := strings.Split(propListSlice[i], "]: ")
            for f := range propLine {
                propLine[f] = strings.ReplaceAll(propLine[f], "]", "")
                propLine[f] = strings.ReplaceAll(propLine[f], "[", "")
            }

            propNames = append(propNames, propLine[0])

            // Some props can be blank
            if len(propLine) < 2 {
                propKeys = append(propKeys, "")
            } else {
                propKeys = append(propKeys, propLine[1])
            }
        }
    }

    for i := range propNames {
        if propNames[i] == name {
            return propKeys[i]
        }
    }
    return ""
}

package main

import (
    "os/exec"
    "strings"
    "os"
)

var propNames []string
var propKeys []string
var isAndroid uint8

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

func isAndroidSystem() bool {
    if isAndroid == 0 {
        if _, err := os.Stat("/system/priv-app"); os.IsNotExist(err) {
            isAndroid = 1
        } else {
            isAndroid = 2
        }
    }
    return isAndroid == 2
}

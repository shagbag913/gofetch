package main

import (
    "fmt"
    "os"
    "os/exec"
    "bufio"
    "regexp"
    "strconv"
    "strings"
)

var debug bool = true
var osName string
var infoSlice [5]string
var infoSliceIter int

func printDebug(str string) {
    if debug == true {
        fmt.Println(str)
    }
}

func openNewReader(filename string) (error, *bufio.Reader, *os.File) {
    file, err := os.Open(filename)
    if err != nil {
        printDebug(err.Error())
        return err, nil, nil
    }

    return err, bufio.NewReader(file), file
}

func iterateInfoSliceNum() {
    infoSliceIter++
}

func _getOsName() string {
    if osName == "" {
        err, reader, file := openNewReader("/etc/os-release")
        if err != nil {
            return ""
        }
        defer file.Close()

        // Skip 'NAME="'
        file.Seek(6, 0)
        fileBs, err := reader.ReadBytes('"')
        if err != nil {
            fmt.Println("oof sound")
            return ""
        }

        osName = string(fileBs[:len(fileBs)-1])
    }
    return osName
}

func getOsName() {
    defer iterateInfoSliceNum()
    infoSlice[0] = "OS: " +  _getOsName()
}

func getUptime() {
    defer iterateInfoSliceNum()
    err, reader, file := openNewReader("/proc/uptime")
    if err != nil {
        return
    }
    defer file.Close()

    fileBs, err := reader.ReadBytes(' ')
    if err != nil {
        printDebug(err.Error())
        return
    }

    // Remove space and milliseconds
    uptime, err := strconv.Atoi(string(fileBs[:len(fileBs) - 4]))
    if err != nil {
        printDebug(err.Error())
        return
    }

    years := uptime / (60 * 60 * 60 * 24 * 365)
    remainder := uptime % (60 * 60 * 60 * 24 * 365)

    months := years / (60 * 60 * 24 * 31)
    remainder = remainder % (60 * 60 * 24 * 31)

    days := remainder / (60 * 60 * 24)
    remainder = remainder % (60 * 60 * 24)

    hours := remainder / (60 * 60)
    remainder = remainder % (60 * 60)

    minutes := remainder / 60

    seconds := uptime % 60

    combTime := []int{years, months, days, hours, minutes, seconds}
    combTimeString := []string{"year", "month", "day", "hour", "minute", "second"}

    var finalTimeString string

    for idx, time := range combTime {
        if time == 0 {
            continue
        }

        finalTimeString += strconv.Itoa(time) + " " + combTimeString[idx]
        if time > 1 {
            finalTimeString += "s"
        }
        finalTimeString += ", "
    }
    finalTimeString = finalTimeString[:len(finalTimeString) - 2]

    infoSlice[3] = "Uptime: " + finalTimeString
}

func getKernelVersion() {
    defer iterateInfoSliceNum()
    file, err := os.Open("/proc/version")
    if err != nil {
        return
    }
    defer file.Close()

    kVersion := make([]byte, 100)
    file.Read(kVersion)

    kRegex, err := regexp.Compile(`[0-9].+?\s`)
    if err != nil {
        return
    }

    versionSlice := kRegex.Find([]byte(kVersion))

    infoSlice[1] = "Kernel: " + string(versionSlice)
}

func getShell() {
    defer iterateInfoSliceNum()
    shell := os.Getenv("SHELL")
    version, err := exec.Command(shell, "--version").Output()
    if err != nil {
        infoSlice[2] = "Shell: " + shell
    } else {
        infoSlice[2] = "Shell: " + string(version[:len(version)-1])
    }
}

func getPackages() {
    defer iterateInfoSliceNum()
    var packagesList []byte
    var numPackages int
    var packagesString string
    var err error

    switch _getOsName() {
    case "Arch Linux":
        packagesList, err = exec.Command("pacman", "-Q", "-q").Output()
        if err != nil {
            break
        }
        numPackages = strings.Count(string(packagesList), "\n")
        packagesString = packagesString + strconv.Itoa(numPackages) + " (pacman) "
        fallthrough
    case "Ubuntu":
        packagesList, err = exec.Command("dpkg", "--list").Output()
        if err != nil {
            break
        }
        numPackages = strings.Count(string(packagesList), "\n")
        packagesString = packagesString + strconv.Itoa(numPackages) + " (dpkg) "
    }

    if packagesString != "" {
        infoSlice[4] = "Packages: " + packagesString
    }
}

func main() {
    go getUptime()
    go getOsName()
    go getKernelVersion()
    go getShell()
    go getPackages()

    for {
        if len(infoSlice) == infoSliceIter {
            break
        }
    }

    for i := 0; i < len(infoSlice); i++ {
        if infoSlice[i] != "" {
            fmt.Println(infoSlice[i])
        }
    }
}

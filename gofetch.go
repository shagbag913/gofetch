package main

import (
    "fmt"
    "os"
    "os/exec"
    "bufio"
    "strconv"
    "strings"
)

var osName string

func openNewReader(filename string) (error, *bufio.Reader, *os.File) {
    file, err := os.Open(filename)
    if err != nil {
        fmt.Println("OOF")
        return err, nil, nil
    }

    return err, bufio.NewReader(file), file
}

func _getOsName() string {
    if osName == "" {
        err, reader, file := openNewReader("/etc/os-release")
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

func getOsName(channel chan string) {
    channel <- "OS: " +  _getOsName()
}

func getUptime(channel chan string) {
    err, reader, file := openNewReader("/proc/uptime")
    defer file.Close()

    fileBs, err := reader.ReadBytes(' ')
    if err != nil {
        fmt.Println("oof sound")
        return
    }

    // Remove space and milliseconds
    uptime, err := strconv.Atoi(string(fileBs[:len(fileBs) - 4]))
    if err != nil {
        fmt.Println("oof sound")
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

    channel <- "Uptime: " + finalTimeString
}

func getKernelVersion(channel chan string) {
    _, reader,  file := openNewReader("/proc/version")
    defer file.Close()

    var versionSlice []byte
    firstNumFound := false
    for {
        vByte, _ := reader.ReadByte()
        vString := string(vByte)
        if firstNumFound == false {
            _, err := strconv.Atoi(vString)
            if err != nil {
                continue
            }
            firstNumFound = true
        }
        if vString == " " {
            break
        }
        versionSlice = append(versionSlice, vByte)
        file.Seek(1, 0)
    }

    channel <- "Kernel: " + string(versionSlice)
}

func getShell(channel chan string) {
    shell := os.Getenv("SHELL")
    version, err := exec.Command(shell, "--version").Output()
    if err != nil {
        channel <- "Shell: " + shell
    } else {
        channel <- "Shell: " + string(version[:len(version)-1])
    }
}

func getPackages(channel chan string) {
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
    channel <- "Packages: " + packagesString
}

func main() {
    channel := make(chan string)
    infoSlice := make([]string, 5)

    go getUptime(channel)
    go getOsName(channel)
    go getKernelVersion(channel)
    go getShell(channel)
    go getPackages(channel)

    for i := 0; i < 5; i++ {
        info := <-channel
        switch info[:2] {
        case "OS":
            infoSlice[0] = info
        case "Ke":
            infoSlice[1] = info
        case "Sh":
            infoSlice[2] = info
        case "Up":
            infoSlice[3] = info
        case "Pa":
            infoSlice[4] = info
        default:
            fmt.Println("N/A:", info)
        }
    }

    for i := 0; i < len(infoSlice); i++ {
        fmt.Println(infoSlice[i])
    }
}

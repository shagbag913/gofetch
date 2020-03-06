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
var infoSlice [6]string
var infoSliceIter int
var ascii [5]string

func printDebug(str string) {
    if debug == true {
        fmt.Println(str)
    }
}

func openNewReader(filename string) (error, *bufio.Reader, *os.File) {
    file, err := os.Open(filename)
    if err != nil {
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
            printDebug(err.Error())
            return ""
        }
        defer file.Close()

        // Skip 'NAME="'
        file.Seek(6, 0)
        fileBs, err := reader.ReadBytes('"')
        if err != nil {
            printDebug(err.Error())
            return ""
        }

        osName = string(fileBs[:len(fileBs)-1])
    }
    return osName
}

func getOsName() {
    defer iterateInfoSliceNum()
    if _getOsName() != "" {
        infoSlice[0] = "OS: " +  _getOsName()
    }
}

func getUptime() {
    defer iterateInfoSliceNum()
    err, reader, file := openNewReader("/proc/uptime")
    if err != nil {
        printDebug(err.Error())
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

func getCpuName() {
    defer iterateInfoSliceNum()

    var coreCount int
    var cpuName string

    file, err := os.Open("/proc/cpuinfo")
    if err != nil {
        printDebug(err.Error())
        return
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        if len(scanner.Text()) > 9 && scanner.Text()[:10] == "model name" {
            if coreCount == 0 {
                cpuName = scanner.Text()[13:]
            }
            coreCount++
        }
    }
    if cpuName == "" {
        return
    }

    newCpuName := cpuName

    // Remove double spaces from CPU name
    newCpuName = strings.ReplaceAll(newCpuName, "  ", " ")

    // Remove (R) from CPU name
    newCpuName = strings.ReplaceAll(newCpuName, "(R)", "")

    // Add core count
    newCpuName += " (" + strconv.Itoa(coreCount) + ")"

    infoSlice[5] = "CPU: " + newCpuName
}

func getAsciiLogo() []string {
    ascii := make([]string, 0)

    ascii = append(ascii, "    /\\")
    ascii = append(ascii, "   /  \\")
    ascii = append(ascii, "  / /\\ \\")
    ascii = append(ascii, " / ____ \\")
    ascii = append(ascii, "/_/    \\_\\")

    return ascii
}

func main() {
    go getUptime()
    go getOsName()
    go getKernelVersion()
    go getShell()
    go getPackages()
    go getCpuName()

    var printBuffer string

    ascii := getAsciiLogo()

    for {
        if len(infoSlice) == infoSliceIter {
            break
        }
    }

    // Filter out empty items
    validInfos := make([]string, 0)
    for i := 0; i < len(infoSlice); i++ {
        if infoSlice[i] != "" {
            validInfos = append(validInfos, infoSlice[i])
        }
    }

    // Get longest ASCII line for info spacers
    var longestAsciiLine int
    for i := 0; i < len(ascii); i++ {
        if len(ascii[i]) > longestAsciiLine {
            longestAsciiLine = len(ascii[i])
        }
    }
    infoSpacer := longestAsciiLine + 2

    var spaces int
    for i := 0; i < len(validInfos); i++ {
        if i < len(ascii) {
            spaces = infoSpacer - len(ascii[i])
            printBuffer = ascii[i] + strings.Repeat(" ", spaces) + validInfos[i]
        } else {
            printBuffer = strings.Repeat(" ", infoSpacer) + validInfos[i]
        }

        fmt.Println(printBuffer)
    }
}

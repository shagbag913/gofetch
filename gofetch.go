package main

import (
    "fmt"
    "os"
    "os/exec"
    "bufio"
    "regexp"
    "strconv"
    "strings"
    "time"
)

var debug bool = true
var osName string
var infoSlice [6]string
var infoSliceIter int
var ascii [5]string

var colorBrightWhite string = "\u001b[37;1m"
var colorBrightBlue string = "\u001b[34;1m"

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
        infoSlice[0] = "OS: " +  colorBrightWhite + _getOsName()
    }
}

func getUptime() {
    defer iterateInfoSliceNum()

    var uptime int

    err, reader, file := openNewReader("/proc/uptime")
    if err != nil {
        printDebug(err.Error())

        // Try to get uptime another way (SELinux)
        uptimeSinceBoot, err := exec.Command("uptime", "-s").Output()
        if err != nil {
            return
        }

        uptimeSinceBootString := string(uptimeSinceBoot[:len(uptimeSinceBoot)-1])

        timeSinceBoot, err := time.Parse("2006-01-02 15:04:05", uptimeSinceBootString)
        if err != nil {
            return
        }

        uptime = int(time.Now().Unix() - timeSinceBoot.Unix())
    } else {
        fileBs, err := reader.ReadBytes(' ')
        file.Close()
        if err != nil {
            printDebug(err.Error())
            return
        }

        // Remove space and milliseconds
        uptime, err = strconv.Atoi(string(fileBs[:len(fileBs) - 4]))
        if err != nil {
            printDebug(err.Error())
            return
        }
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

    infoSlice[3] = "Uptime: " + colorBrightWhite + finalTimeString
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

    infoSlice[1] = "Kernel: " + colorBrightWhite + string(versionSlice)
}

func getShell() {
    defer iterateInfoSliceNum()
    shell := os.Getenv("SHELL")
    version, err := exec.Command(shell, "--version").Output()
    if err != nil {
        infoSlice[2] = "Shell: " + colorBrightWhite + shell
    } else {
        infoSlice[2] = "Shell: " + colorBrightWhite + string(version[:len(version)-1])
    }
}

func getPackages() {
    defer iterateInfoSliceNum()
    var packagesList []byte
    var numPackages int
    var packagesString string
    var err error

    packagesList, err = exec.Command("pacman", "-Q", "-q").Output()
    if err == nil {
        numPackages = strings.Count(string(packagesList), "\n")
        packagesString = packagesString + strconv.Itoa(numPackages) + " (pacman) "
    }

    packagesList, err = exec.Command("dpkg-query", "-f", "\n", "-W").Output()
    if err == nil {
        numPackages = strings.Count(string(packagesList), "\n")
        packagesString = packagesString + strconv.Itoa(numPackages) + " (dpkg) "
    }

    if packagesString != "" {
        infoSlice[4] = "Packages: " + colorBrightWhite + packagesString
    }
}

func getCpuInfoFromProc(scanner bufio.Scanner) (string, int) {
    var coreCount int
    var cpuName string

    cpuNameProcList := []string{"model name", "Hardware"}

    for scanner.Scan() {
        for i := 0; i < len(cpuNameProcList); i++ {
            length := len(cpuNameProcList[i])
            if len(scanner.Text()) >= length && scanner.Text()[:length] == cpuNameProcList[i] {
                cpuName = scanner.Text()[length+3:]
                break
            }
        }
        if len(scanner.Text()) >= 9 && scanner.Text()[:9] == "processor" {
            coreCount++
        }
    }
    return cpuName, coreCount
}

func getCpuName() {
    defer iterateInfoSliceNum()

    file, err := os.Open("/proc/cpuinfo")
    if err != nil {
        printDebug(err.Error())
        return
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    cpuName, coreCount := getCpuInfoFromProc(*scanner)

    if cpuName == "" {
        return
    }

    // Remove double spaces from CPU name
    cpuName = strings.ReplaceAll(cpuName, "  ", " ")

    // Remove (R) from CPU name
    cpuName = strings.ReplaceAll(cpuName, "(R)", "")

    // Add core count
    cpuName += " (" + strconv.Itoa(coreCount) + ")"

    infoSlice[5] = "CPU: " + colorBrightWhite + cpuName
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
    fmt.Println("")
    for i := 0; i < len(validInfos); i++ {
        if i < len(ascii) {
            spaces = infoSpacer - len(ascii[i])
            printBuffer = "  " + colorBrightBlue + ascii[i] + strings.Repeat(" ", spaces) + validInfos[i]
        } else {
            printBuffer = "  " + colorBrightBlue + strings.Repeat(" ", infoSpacer) + validInfos[i]
        }

        fmt.Println(printBuffer)
    }
    fmt.Println("")
}

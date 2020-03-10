package main

import (
    "fmt"
    "os"
    "os/user"
    "os/exec"
    "bufio"
    "regexp"
    "strconv"
    "strings"
    "time"
)

var debug bool
var osName string
var infoSlice [9]string
var infoSliceIter int

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
        if isAndroidSystem() {
            osName = "Android"
            return osName
        }
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
        infoSlice[2] = "OS: " +  colorBrightWhite + _getOsName() + " " + getOsVersion()
    }
}

func getOsVersion() string {
    if isAndroidSystem() {
        return getProp("ro.build.version.release")
    }
    return ""
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

    combTime := []int{years, months, days, hours, minutes}
    combTimeString := []string{"yr", "mo", "day", "hr", "min"}

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

    infoSlice[5] = "Uptime: " + colorBrightWhite + finalTimeString
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

    infoSlice[3] = "Kernel: " + colorBrightWhite + string(versionSlice)
}

func getShell() {
    defer iterateInfoSliceNum()
    shell := os.Getenv("SHELL")
    shellRegex, err := regexp.Compile(`.*/`)
    if err == nil {
        shell = string(shellRegex.ReplaceAllString(shell, ""))
    }
    infoSlice[4] = "Shell: " + colorBrightWhite + shell
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
        infoSlice[6] = "Packages: " + colorBrightWhite + packagesString
    }
}

func getCpuInfoFromProc(scanner bufio.Scanner) (string, int, int) {
    var coreCount int
    var threadCount int
    var cpuName string
    var err error

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
            threadCount++
        }
        if coreCount == 0 && len(scanner.Text()) >= 9 && scanner.Text()[:9] == "cpu cores" {
            coreCount, err = strconv.Atoi(scanner.Text()[12:])
            if err != nil {
                printDebug(err.Error())
                coreCount = 0
                continue
            }
        }
    }
    if coreCount == 0 && threadCount > 0 {
        coreCount = threadCount
    }
    return cpuName, coreCount, threadCount
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
    cpuName, coreCount, threadCount := getCpuInfoFromProc(*scanner)

    if cpuName == "" {
        return
    }

    // Remove (R) from CPU name
    cpuName = strings.ReplaceAll(cpuName, "(R)", "")

    // Shorten QCOM CPU/SoC name
    cpuName = strings.ReplaceAll(cpuName, "Qualcomm Technologies, Inc", "QCOM")

    // Remove "CPU"
    cpuName = strings.ReplaceAll(cpuName, "CPU", "")

    // Remove text-based core count
    cpuRegex, err := regexp.Compile(` [a-zA-Z]+-[cC]ore `)
    if err == nil {
        cpuName = cpuRegex.ReplaceAllString(cpuName, "")
    }

    // Remove "[Pp]rocessor"
    cpuName = strings.ReplaceAll(cpuName, "Processor", "")
    cpuName = strings.ReplaceAll(cpuName, "processor", "")

    cpuRegex, err = regexp.Compile(`\s+`)
    if err == nil {
        cpuName = cpuRegex.ReplaceAllString(cpuName, " ")
    }

    // Add core count
    cpuName += " (" + strconv.Itoa(coreCount) + "c " + strconv.Itoa(threadCount) + "t)"

    infoSlice[7] = "CPU: " + colorBrightWhite + cpuName
}

func getAsciiLogo() []string {
    ascii := make([]string, 0)
    configHome := os.Getenv("XDG_CONFIG_HOME")
    if configHome == "" {
        configHome = os.Getenv("HOME") + "/.config"
    }
    asciiFile := strings.ReplaceAll(strings.ToLower(_getOsName()), " ", "_")
    asciiPath := configHome + "/gofetch/ascii/" + asciiFile

    file, err := os.Open(asciiPath)
    if err != nil {
        printDebug("Couldn't find ASCII for OS " + _getOsName() + " at path " + asciiPath)
        return []string{""}
    }
    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        ascii = append(ascii, scanner.Text())
    }

    return ascii
}

func getMemCapacity() {
    defer iterateInfoSliceNum()
    err, reader, file := openNewReader("/proc/meminfo")
    if err != nil {
        printDebug(err.Error())
        return
    }
    defer file.Close()

    memTotalProc, err := reader.ReadBytes('\n')
    if err != nil {
        printDebug(err.Error())
        return
    }

    memRegex, err := regexp.Compile(`[0-9]+`)
    if err != nil {
        printDebug(err.Error())
        return
    }
    memTotal, err := strconv.Atoi(string(memRegex.Find(memTotalProc)))
    if err != nil {
        printDebug(err.Error())
        return
    }
    memTotalMB := strconv.Itoa(memTotal / 1000)

    infoSlice[8] = "Memory: " + colorBrightWhite + memTotalMB + "MB"
}

func getHostUserHeader() {
    defer iterateInfoSliceNum()
    defer iterateInfoSliceNum()
    user, err := user.Current()
    if err != nil {
        printDebug(err.Error())
        return
    }

    hostname, err := os.Hostname()
    if err != nil {
        printDebug(err.Error())
        return
    }

    infoSlice[0] = colorBrightBlue + user.Username + colorBrightWhite + "@" + colorBrightBlue + hostname
    infoSlice[1] = colorBrightWhite + strings.Repeat("-", len(user.Username) + len(hostname) + 1)
}

func main() {
    go getUptime()
    go getOsName()
    go getKernelVersion()
    go getShell()
    go getPackages()
    go getCpuName()
    go getMemCapacity()
    go getHostUserHeader()

    var printBuffer string

    for {
        if len(infoSlice) == infoSliceIter {
            break
        }
    }

    // Filter out empty items
    validInfos := make([]string, 0)
    for iter := range infoSlice {
        if infoSlice[iter] != "" {
            validInfos = append(validInfos, infoSlice[iter])
        }
    }

    // Get longest ASCII line for info spacers
    ascii := getAsciiLogo()
    var longestAsciiLine int
    for iter := range ascii {
        if len(ascii[iter]) > longestAsciiLine {
            longestAsciiLine = len(ascii[iter])
        }
    }
    infoSpacer := longestAsciiLine + 2

    infoThreshold := -1
    if len(ascii)+1 > len(validInfos) {
        infoThreshold = (len(ascii) - len(validInfos)) / 2
    }

    asciiThreshold := -1
    if len(ascii) < len(validInfos)+1 {
        asciiThreshold = (len(validInfos) - len(ascii)) / 2
    }

    fmt.Println("")
    iter := 0
    infoIter := 0
    asciiIter := 0
    for {
        spacer := infoSpacer
        printBuffer = "  " + colorBrightBlue
        if asciiIter < len(ascii) && iter >= asciiThreshold {
            printBuffer += ascii[asciiIter]
            spacer -= len(ascii[asciiIter])
            asciiIter++
        }
        printBuffer += strings.Repeat(" ", spacer)
        if infoIter < len(validInfos) && iter >= infoThreshold {
            printBuffer += validInfos[infoIter]
            infoIter++
        }

        // If there's only color escape sequences and spaces then we're done here
        if strings.ReplaceAll(printBuffer, " ", "") == colorBrightBlue {
            break
        }

        fmt.Println(printBuffer)
        iter++
    }
    fmt.Println("")
}

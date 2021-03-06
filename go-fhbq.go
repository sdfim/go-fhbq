package main

import (
    "fmt"
    "flag"
    "github.com/tarm/serial"
    "io"
    "strings"
    "time"
    "bufio"
    "reflect"
    "github.com/fatih/color"
    "strconv"
)

//map is used to check the validity of the commands
//as well as to display a hint of possible/valid commands
var modes = map[string]string{
    "status":    "view current status",
    "n 1 auto":  "mode: normal; speed: 1; bypass: auto;",
    "n 2 auto":  "mode: normal; speed: 2; bypass: auto;",
    "n 3 auto":  "mode: normal; speed: 3; bypass: auto;",
    "n 1 on":    "mode: normal; speed: 1; bypass: on;",
    "n 2 on":    "mode: normal; speed: 2; bypass: on;",
    "n 3 on":    "mode: normal; speed: 3; bypass: on;",
    "n 1 off":   "mode: normal; speed: 1; bypass: off;",
    "n 2 off":   "mode: normal; speed: 2; bypass: off;",
    "n 3 off":   "mode: normal; speed: 3; bypass: off;",
    "ne 1 auto": "mode: normal exhaust; speed: 1; bypass: auto;",
    "ne 3 auto": "mode: normal exhaust; speed: 3; bypass: auto;",
    "ne 1 on":   "mode: normal exhaust; speed: 1; bypass: on;",
    "ne 3 on":   "mode: normal exhaust; speed: 3; bypass: on;",
    "ne 1 off":  "mode: normal exhaust; speed: 1; bypass: off;",
    "ne 3 off":  "mode: normal exhaust; speed: 3; bypass: off;",
    "ns 1 auto": "mode: normal supply; speed: 1; bypass: auto;",
    "ns 3 auto": "mode: normal supply; speed: 3; bypass: auto;",
    "ns 1 on":   "mode: normal supply; speed: 1; bypass: on;",
    "ns 3 on":   "mode: normal supply; speed: 3; bypass: on;",
    "ns 1 off":  "mode: normal supply; speed: 1; bypass: off;",
    "ns 3 off":  "mode: normal supply; speed: 3; bypass: off;",
    "s 1 auto":  "mode: save; speed: 1; bypass: auto;",
    "s 2 auto":  "mode: save; speed: 2; bypass: auto;",
    "s 3 auto":  "mode: save; speed: 3; bypass: auto;",
    "s 1 on":    "mode: save; speed: 1; bypass: on;",
    "s 2 on":    "mode: save; speed: 2; bypass: on;",
    "s 3 on":    "mode: save; speed: 3; bypass: on;",
    "s 1 off":   "mode: save; speed: 1; bypass: off;",
    "s 2 off":   "mode: save; speed: 2; bypass: off;",
    "s 3 off":   "mode: save; speed: 3; bypass: off;",
    "se 1 auto": "mode: save exhaust; speed: 1; bypass: auto;",
    "se 3 auto": "mode: save exhaust; speed: 3; bypass: auto;",
    "se 1 on":   "mode: save exhaust; speed: 1; bypass: on;",
    "se 3 on":   "mode: save exhaust; speed: 3; bypass: on;",
    "se 1 off":  "mode: save exhaust; speed: 1; bypass: off;",
    "se 3 off":  "mode: save exhaust; speed: 3; bypass: off;",
    "ss 1 auto": "mode: save supply; speed: 1; bypass: auto;",
    "ss 3 auto": "mode: save supply; speed: 3; bypass: auto;",
    "ss 1 on":   "mode: save supply; speed: 1; bypass: on;",
    "ss 3 on":   "mode: save supply; speed: 3; bypass: on;",
    "ss 1 off":  "mode: save supply; speed: 1; bypass: off;",
    "ss 3 off":  "mode: save supply; speed: 3; bypass: off;",
    "off":       "off",
    "rhoff":     "rhoff",
    "rhon":      "rhon",
}

//read k byte
func  readPack(s io.ReadWriteCloser, k int)  []byte {
    var bf []byte
    reader := bufio.NewReader(s)
    var rep byte
    for i:=0; i<k; i++ {
        rep, _ = reader.ReadByte()
        bf = append(bf, rep)
    }
    //fmt.Printf("read <-  % x\n", bf)
    return bf
}

var blockT, unitT, checkT []byte

//get start position
func startPosition(s io.ReadWriteCloser) {
    fmt.Println("startPosition start")
    reader := bufio.NewReader(s)
    rep, _ := reader.ReadBytes('\x7e')
    fmt.Println("rep", rep)
    buf := readPack(s, 17)
    n := 0
    for {
        n++
        fmt.Println(n, buf)
        if reflect.DeepEqual(buf[:4], []byte{0x7e, 0x7e, 0xa0, 0x00}) {
            buf = readPack(s, 17)
            break
        }else  {buf = readPack(s, 17) }
    }
    fmt.Println("startPosition end")
}

func  readTelegram(s io.ReadWriteCloser, z string) {
    buf := readPack(s, 17)
    if echoTelegram { fmt.Printf("read <-  % x\n", buf) }
    for {
        if (z == "start" || z == "check") && reflect.DeepEqual(buf[:4], []byte{0x7e, 0x7e, 0xc0, 0xff}) {
            break
        }else if z == "block" && reflect.DeepEqual(buf[:4], []byte{0x7e, 0x7e, 0x00, 0xa0}) {
            break
        }else if z == "unit" && reflect.DeepEqual(buf[:4], []byte{0x7e, 0x7e, 0xa0, 0x00}) {
                break
            }else {
            buf = readPack(s, 17)
            if echoTelegram { fmt.Printf("read <-  % x\n", buf) }
        }
    }
    if z == "start" || z == "check" { checkT = buf 
    }else if z == "block" { blockT = buf 
    }else if z == "unit" { unitT = buf }
}

//read current status
func getStatus() string {
    var bypass, mode string
    status := ""
    if len(checkT) > 0 {
        rx := checkT
        if rx[9] == '\x0a' || rx[9] == '\x2a' || rx[9] == '\x4a' {
            status = "off"
        }else {
            if rx[9] == '\x8a'  { bypass = "bypass: auto; "}
            if rx[9] == '\xaa'  { bypass = "bypass: on; "}
            if rx[9] == '\xca'  { bypass = "bypass: off; "}
            if (rx[13] == '\x00' || rx[13] == '\x20'){
                if rx[10] == '\x0c'  { mode = "mode: normal; speed: 1; "}
                if rx[10] == '\x12'  { mode = "mode: normal; speed: 2; "}
                if rx[10] == '\x21'  { mode = "mode: normal; speed: 3; "}
                if rx[10] == '\x4a'  { mode = "mode: normal exhaust; speed: 1; "}
                if rx[10] == '\x51'  { mode = "mode: normal exhaust; speed: 3; "}
                if rx[10] == '\x94'  { mode = "mode: normal supply; speed: 1; "}
                if rx[10] == '\xa2'  { mode = "mode: normal supply; speed: 3; "}
            }
            if rx[13] == '\x10'  {
                if rx[10] == '\x0c'  { mode = "mode: save; speed: 1; "}
                if rx[10] == '\x12'  { mode = "mode: save; speed: 2; "}
                if rx[10] == '\x21'  { mode = "mode: save; speed: 3; "}
                if rx[10] == '\x4a'  { mode = "mode: save exhaust; speed: 1; "}
                if rx[10] == '\x51'  { mode = "mode: save exhaust; speed: 3; "}
                if rx[10] == '\x94'  { mode = "mode: save supply; speed: 1; "}
                if rx[10] == '\xa2'  { mode = "mode: save supply; speed: 3; "}
            }
            status = mode + bypass
        }
    }
    if status == "" {status = "error get status"}
    //fmt.Println("status:", status)
    return status
}

//calculate checksum
func checkSum (p []byte) byte {
    var sum byte
    for _, e := range p {
        sum ^= e
    }
    return sum
}

//command execution
func runCommand (s io.ReadWriteCloser, cm []string) string {
    readTelegram(s, "check")
    rx := checkT
    rx[2] = '\x00'
    rx[3] = '\xa0'
    if cm[0] == "off" {
        if rx[9] == '\x8a'{  rx[9] = '\x0a' }
        if rx[9] == '\xaa'{  rx[9] = '\x2a' }
        if rx[9] == '\xca'{  rx[9] = '\x4a' }
    }else if cm[0] == "rhoff" && (rx[9] == '\x8a' || rx[9] == '\xaa' || rx[9] == '\xca') {
        fmt.Println ("use rhoff")
        rx[11] = '\x40'
    }else if cm[0] == "rhon" && (rx[9] == '\x8a' || rx[9] == '\xaa' || rx[9] == '\xca') {
        fmt.Println ("use rhon")
        rx[11] = '\xd0'
    }else {
        if cm[2] == "auto"{  rx[9] = '\x8a'    }        //bypass: auto;
        if cm[2] == "on"{  rx[9] = '\xaa'   }           //bypass: on;
        if cm[2] == "off"{  rx[9] = '\xca'  }           //bypass: off;
        if cm[0] == "n" || cm[0] == "ne" || cm[0] == "ns" {
            rx[13] = '\x20'
            //rx[13] = '00'
            if cm[0] == "n" && cm[1] == "1" { rx[10] = '\x0c' }            //mode: normal; speed: 1;
            if cm[0] == "n" && cm[1] == "2" { rx[10] = '\x12' }            //mode: normal; speed: 2;
            if cm[0] == "n" && cm[1] == "3" { rx[10] = '\x21' }            //mode: normal; speed: 3;
            if cm[0] == "ne" && cm[1] == "1" { rx[10] = '\x4a' }           //mode: normal exhaust; speed: 1;
            if cm[0] == "ne" && cm[1] == "3" { rx[10] = '\x51' }           //mode: normal exhaust; speed: 3;
            if cm[0] == "ns" && cm[1] == "1" { rx[10] = '\x94' }           // mode: normal supply; speed: 1;
            if cm[0] == "ns" && cm[1] == "3" { rx[10] = '\xa2' }           //mode: normal supply; speed: 3;
        }
        if cm[0] == "s" || cm[0] == "se" || cm[0] == "ss" {
            rx[13] = '\x10'
            if cm[0] == "s" && cm[1] == "1" { rx[10] = '\x0c' }            //mode: save; speed: 1;
            if cm[0] == "s" && cm[1] == "2" { rx[10] = '\x12' }            //mode: save; speed: 2;
            if cm[0] == "s" && cm[1] == "3" { rx[10] = '\x21' }            //mode: save; speed: 3;
            if cm[0] == "se" && cm[1] == "1" { rx[10] = '\x4a' }           //mode: save exhaust; speed: 1;
            if cm[0] == "se" && cm[1] == "3" { rx[10] = '\x51' }           //mode: save exhaust; speed: 3;
            if cm[0] == "ss" && cm[1] == "1" { rx[10] = '\x94' }           //mode: save supply; speed: 1;
            if cm[0] == "ss" && cm[1] == "3" { rx[10] = '\xa2' }           //mode: save supply; speed: 3;
        }
    }
    rx[16] = checkSum (rx[:16])

    return writeTelegram(s, rx, cm)
}

func writeTelegram (s io.ReadWriteCloser, tx []byte, cm []string) string {
    m := 0
    repeat:
    readTelegram(s, "block")
    _, err := s.Write(tx)
    if err != nil {
    fmt.Printf("s.Write: %v", err)
    }
    if echoTelegram { fmt.Printf("write -> % x\n", tx) }
    //test
    readTelegram(s, "check")
    checkStatus := getStatus()
    checkStatus = strings.TrimRight(checkStatus, " ")
    cmString := strings.Join(cm, " ")
    fmt.Println(checkStatus)
    var ret string
    
    for {
        m++
        fmt.Println("probe:", m)
        if modes[cmString] == checkStatus {
            ret = "DONE"
            break
        }else if m%2 == 0 {
            startPosition(s)
        }else if  m > 5 {
            ret = "ERROR"
            break
        }else {
            goto repeat
        }
        
    }
    return ret
}

//string(csv) to json
func getJSONStatus (str string) string {
    str = strings.Replace(str, "; ", "\", \"", -1)
    str = strings.Replace(str, ": ", "\": \"", -1)
    str = strings.TrimRight(str, ", \"")
    str = "{\""+str+"\"}"
    return str
}

func snifferFunc (s io.ReadWriteCloser, np int) {
    readTelegram(s, "start")
        var com string
        for sn:=1; sn<=np; sn++ {
            com = getStatus()
            if len(com) == 0 { com = "???"    }
            color.Green("%s", "paket: "+strconv.Itoa(sn)+" COMMAND   ->  "+com)
            readTelegram(s, "block")
            fmt.Printf("block:   % x\n", blockT)
            readTelegram(s, "unit")
            fmt.Printf("unit:    % x\n", unitT)
            readTelegram(s, "check")
            fmt.Printf("check:   % x\n", checkT)
        }
}

func snifferFullFunc (s io.ReadWriteCloser, np int, ignore bool) {
    readTelegram(s, "start")
    var buf, bufAll []byte
    var com string
    sn := 1
    sm := 1
    var x int
    var blockPrint,    unitPrint, checkPrint bool
    for {
        blockPrint = true
        unitPrint = true
        checkPrint = true
        x = 0
        for {
            buf = readPack(s, 17)
            bufAll = append(bufAll, buf...)
            x++
            if reflect.DeepEqual(buf[:4], []byte{0x7e, 0x7e, 0x00, 0xa0}) { 
                if reflect.DeepEqual(buf, blockT) { blockPrint = false}
                blockT = buf
            }
            if reflect.DeepEqual(buf[:4], []byte{0x7e, 0x7e, 0xa0, 0x00}) {
                if ignore && unitT != nil {
                    if reflect.DeepEqual(buf[:16], unitT[:16]) {
                        unitPrint = false
                    }
                }else if reflect.DeepEqual(buf, unitT) {
                    unitPrint = false
                }
                unitT = buf
            }
            if reflect.DeepEqual(buf[:4], []byte{0x7e, 0x7e, 0xc0, 0xff}) { 
                if reflect.DeepEqual(buf, checkT) { checkPrint = false}
                checkT = buf
                buf = bufAll
                bufAll = nil
                break 
            }
        }
        if blockPrint || unitPrint || checkPrint {
            for i := 0; i < x; i++ {
                if reflect.DeepEqual(buf[i*17:i*17+4], []byte{0x7e, 0x7e, 0xc0, 0xff}) {
                    fmt.Printf("check:   % x\n", buf[i*17:i*17+17])
                    com = getStatus()
                    if len(com) == 0 { com = "???"    }
                    color.Green("%s", "paket: "+strconv.Itoa(sn)+" COMMAND   ->  "+com)
                    fmt.Println("time:   ", time.Now().Format("15:04:05.000000"))
                    fmt.Println("--------------------------")
                    sn++
                }else if reflect.DeepEqual(buf[i*17:i*17+4], []byte{0x7e, 0x7e, 0x00, 0xa0}) {
                    fmt.Printf("block:   % x\n", buf[i*17:i*17+17])
                }else if reflect.DeepEqual(buf[i*17:i*17+4], []byte{0x7e, 0x7e, 0xa0, 0x00}) {
                    fmt.Printf("unit:    % x\n", buf[i*17:i*17+17])
                }else { fmt.Printf("?????:   % x\n", buf[i*17:i*17+17]) }
            }
        }
        if sn > np { break }
        if sm > np*5 { break }
        sm++
    }
}

func snifferDifFunc (s io.ReadWriteCloser, np int, ignore bool) {
    color.Blue("%s", "snifferDif run")
        readTelegram(s, "start")
        
        readTelegram(s, "block")
        b := blockT
        readTelegram(s, "unit")
        u := unitT
        readTelegram(s, "check")
        c := checkT
        com := getStatus()
        if len(com) == 0 {
            com = "???"
        }
        color.Green("%s", "paket START  ->  "+com)
        fmt.Printf("block:   % x\n", b)
        fmt.Printf("unit:    % x\n", u)
        fmt.Printf("check:   % x\n", c)
        num := 0
        for sn:=1; sn<=np; sn++ {
            var dif bool
            readTelegram(s, "block")
            bc := blockT
            readTelegram(s, "unit")
            uc := unitT
            readTelegram(s, "check")
            cc := checkT
            if ignore {
                dif = reflect.DeepEqual(uc[:16], u[:16])
            }else {
                dif = reflect.DeepEqual(uc, u)
            }
            if !reflect.DeepEqual(bc, b) || !dif || !reflect.DeepEqual(cc, c) {
                num++
                com := getStatus()
                if len(com) == 0 {
                    com = "???"
                }
                b = bc
                u = uc
                c = cc
                //fmt.Println("paket:", sn, "NEW COMMAND -> ", com)
                color.Green("%s", "paket: "+strconv.Itoa(sn)+" NEW COMMAND "+strconv.Itoa(num)+"  ->  "+com)
                //fmt.Println(time.Now().Format(time.RFC850))
                fmt.Println("time:   ", time.Now().Format("15:04:05.000000"))
                fmt.Printf("block:   % x\n", b)
                fmt.Printf("unit:    % x\n", u)
                fmt.Printf("check:   % x\n", c)
            }
        }
}

var echoTelegram bool

func main() {
    start := time.Now()

    var rpinttel, rpinterr, valid, json, timer, ignore bool
    flag.BoolVar(&rpinttel, "rpinttel", false, "print a telegram")
    flag.BoolVar(&rpinterr, "rpinterr", false, "print of write error telegrem")
    flag.BoolVar(&valid, "valid", false, "print a valid commands")
    flag.BoolVar(&json, "json", false, "print a status in json formate")
    flag.BoolVar(&timer, "timer", false, "print script execution time")
    flag.BoolVar(&ignore, "ignore", false, "ignore unit in snifferDif")
    flag.BoolVar(&echoTelegram, "echoTelegram", false, "ignore unit in snifferDif")
    var sniffer, snifferDif, snifferFull int
    flag.IntVar(&sniffer, "sniffer", 0, "run sniffer for N packet, one packet = 3 telegrams")
    flag.IntVar(&snifferDif, "snifferDif", 0, "run sniffer view renge command fo N packet, one packet = 3 telegrams")
    flag.IntVar(&snifferFull, "snifferFull", 0, "run sniffer full hfcrtts")
    var usb string
    flag.StringVar(&usb, "port", "/dev/ttyUSB1", "used USB example: /dev/ttyUSB0")
    
    flag.Parse()

    c := &serial.Config{Name: usb, Baud: 9600 }
    s, err := serial.OpenPort(c)
    if err != nil {    fmt.Println(err) }

    defer s.Close()
    
    //for test
    //startPosition(s)

    //logic body
    //getting arguments and run commands
    
    //stream full sniffer
    if snifferFull > 0 {
        snifferFullFunc(s, snifferFull, ignore)
        fmt.Println(time.Now().Sub(start))
    }

    //stream sniffer
    if sniffer > 0 {
        snifferFunc(s, sniffer)
        fmt.Println(time.Now().Sub(start))
    }
    //sniffer outputs packets of telegrams, if there were changes in the stream
    //if is set flag -ignore, change with unit telegram will be ignore
    if snifferDif > 0 {
        snifferDifFunc(s, snifferDif, ignore)
        fmt.Println(time.Now().Sub(start))
    }

    //print a valid commands
    if valid {
        for i, n := range modes {
            fmt.Println( i, "\t", n)
        }
    }
    
    //get a command and run 
    command := strings.Join(flag.Args(), " ")

    if _, ok := modes[command]; ok {
        if flag.Args()[0] == "status" {
            readTelegram(s, "start")
            ansSt:= getStatus()
            if json {
                fmt.Println(getJSONStatus(ansSt))
            }else {
                if timer { fmt.Println(time.Now().Sub(start)) }
                fmt.Println(ansSt)
            }
        }else {
            ansCom := runCommand(s, flag.Args())
            if timer { fmt.Println(time.Now().Sub(start)) }
            fmt.Println(ansCom)
        }
    } else if sniffer == 0 && snifferDif == 0 && snifferFull == 0{
        fmt.Println("your command is not valid, see possible/valid commands with flag -valid and -help for used flag")
    }
    
}

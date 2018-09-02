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
	bf := make([]byte, 0)
	reader := bufio.NewReader(s)
	var rep byte
	for i:=0; i<k; i++ {
		rep, _ = reader.ReadByte()
		bf = append(bf, rep)
	}
	//fmt.Printf("\nread: % x\n", bf)
	return bf
}

var blockT, unitT, checkT []byte

//get start position
func  readTelegram(s io.ReadWriteCloser, z string) {
	buf := readPack(s, 17)
    for {
		if (z == "start" || z == "check") && reflect.DeepEqual(buf[:4], []byte{0x7e, 0x7e, 0xc0, 0xff}) {
            break
        }else if z == "block" && reflect.DeepEqual(buf[:4], []byte{0x7e, 0x7e, 0x00, 0xa0}) {
            break
        }else if z == "unit" && reflect.DeepEqual(buf[:4], []byte{0x7e, 0x7e, 0xa0, 0x00}) {
				break
			}else {
			buf = readPack(s, 17)
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
        if cm[2] == "auto"{  rx[9] = '\x8a'	}		//'bypass: auto; '
        if cm[2] == "on"{  rx[9] = '\xaa'   }		//'bypass: on; '
        if cm[2] == "off"{  rx[9] = '\xca'  }		//'bypass: off; '
        if cm[0] == "n" || cm[0] == "ne" || cm[0] == "ns" {
            rx[13] = '\x20'
            //rx[13] = '00'
            if cm[0] == "n" && cm[1] == "1" { rx[10] = '\x0c' }			//'mode: normal; speed: 1; '
            if cm[0] == "n" && cm[1] == "2" { rx[10] = '\x12' }			//'mode: normal; speed: 2; '
            if cm[0] == "n" && cm[1] == "3" { rx[10] = '\x21' }			//'mode: normal; speed: 3; '
            if cm[0] == "ne" && cm[1] == "1" { rx[10] = '\x4a' }		//'mode: normal exhaust; speed: 1; '
            if cm[0] == "ne" && cm[1] == "3" { rx[10] = '\x51' }		//'mode: normal exhaust; speed: 3; '
            if cm[0] == "ns" && cm[1] == "1" { rx[10] = '\x94' }		// 'mode: normal supply; speed: 1; '
			if cm[0] == "ns" && cm[1] == "3" { rx[10] = '\xa2' }		//'mode: normal supply; speed: 3; '
		}
        if cm[0] == "s" || cm[0] == "se" || cm[0] == "ss" {
            rx[13] = '\x10'
            if cm[0] == "s" && cm[1] == "1" { rx[10] = '\x0c' }			//'mode: save; speed: 1; '
            if cm[0] == "s" && cm[1] == "2" { rx[10] = '\x12' }			//'mode: save; speed: 2; '
            if cm[0] == "s" && cm[1] == "3" { rx[10] = '\x21' }			//'mode: save; speed: 3; '
            if cm[0] == "se" && cm[1] == "1" { rx[10] = '\x4a' }		//'mode: save exhaust; speed: 1; '
            if cm[0] == "se" && cm[1] == "3" { rx[10] = '\x51' }		//'mode: save exhaust; speed: 3; '
            if cm[0] == "ss" && cm[1] == "1" { rx[10] = '\x94' }		//'mode: save supply; speed: 1; '
            if cm[0] == "ss" && cm[1] == "3" { rx[10] = '\xa2' }		//'mode: save supply; speed: 3; '
		}
	}
	rx[16] = checkSum (rx[:16])

	m := 0
	repeat:
		//выходим на позицию записи
		readTelegram(s, "block")
		_, err := s.Write(rx)
		if err != nil {
		fmt.Printf("s.Write: %v", err)
		}
		checkStatus := getStatus()
		checkStatus = strings.TrimRight(checkStatus, " ")
		cmString := strings.Join(cm, " ")
		fmt.Println(checkStatus)

		var ret string
		
		for {
			m++
			if modes[cmString] == checkStatus {
				ret = "DONE"
				break
			}else if  m > 3 {
				ret = "ERROR"
				break
			}else {
				goto repeat
			}
			
		}
	return ret
}

//string(csv) to json
func getJsonStatus (str string) string {
	str = strings.Replace(str, "; ", "\", \"", -1)
	str = strings.Replace(str, ": ", "\": \"", -1)
	str = strings.TrimRight(str, ", \"")
	str = "{\""+str+"\"}"
	return str
}

func main() {
    start := time.Now()

	var rpinttel, rpinterr, valid, json, timer, ignore bool
	flag.BoolVar(&rpinttel, "rpinttel", false, "print a telegram")
	flag.BoolVar(&rpinterr, "rpinterr", false, "print of write error telegrem")
	flag.BoolVar(&valid, "valid", false, "print a valid commands")
	flag.BoolVar(&json, "json", false, "print a status in json formate")
	flag.BoolVar(&timer, "timer", false, "print script execution time")
	flag.BoolVar(&ignore, "ignore", false, "ignore unit in sniffer_dif")
	var sniffer, sniffer_dif int
	flag.IntVar(&sniffer, "sniffer", 0, "run sniffer for N packet, one packet = 3 telegrams")
	flag.IntVar(&sniffer_dif, "sniffer_dif", 0, "run sniffer view renge command fo N packet, one packet = 3 telegrams")
	var usb string
	flag.StringVar(&usb, "port", "/dev/ttyUSB1", "used USB example: /dev/ttyUSB0")
    
    flag.Parse()

	c := &serial.Config{Name: usb, Baud: 9600 }
	s, err := serial.OpenPort(c)
	if err != nil {	fmt.Println(err) }

	defer s.Close()
    
    //logic body
	//getting arguments and run commands
	
	//stream sniffer
	if sniffer > 0 {
		readTelegram(s, "start")
		var com string
		for sn:=1; sn<=sniffer; sn++ {
			com = getStatus()
			if len(com) == 0 { com = "???"	}
			color.Green("%s", "paket: "+strconv.Itoa(sn)+" COMMAND   ->  "+com)
			readTelegram(s, "block")
			fmt.Printf("block:   % x\n", blockT)
			readTelegram(s, "unit")
			fmt.Printf("unit:    % x\n", unitT)
			readTelegram(s, "check")
			fmt.Printf("check:   % x\n", checkT)
		}
		fmt.Println(time.Now().Sub(start))
	}
	//sniffer outputs packets of telegrams, if there were changes in the stream
	if sniffer_dif > 0 {
		color.Blue("%s", "sniffer_dif run")
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
		for sn:=1; sn<=sniffer_dif; sn++ {
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
		fmt.Println(time.Now().Sub(start))
	}

	if valid {
		for i, n := range modes {
			fmt.Println( i, "\t", n)
		}
	}
	
	command := strings.Join(flag.Args(), " ")

	if _, ok := modes[command]; ok {
		readTelegram(s, "start")
		if flag.Args()[0] == "status" {
			ansSt:= getStatus()
			if json {
				fmt.Println(getJsonStatus(ansSt))
			}else {
				if timer { fmt.Println(time.Now().Sub(start)) }
				fmt.Println(ansSt)
			}
		}else {
			ansCom := runCommand(s, flag.Args())
			if timer { fmt.Println(time.Now().Sub(start)) }
			fmt.Println(ansCom)
		}
	} else if sniffer == 0 && sniffer_dif == 0 {
		fmt.Println("your command is not valid, see possible/valid commands with flag -valid and -help for used flag")
	}
	
    
    
    //fmt.Println(time.Now().Sub(start))
}


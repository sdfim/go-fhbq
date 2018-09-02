[![GoDoc](https://godoc.org/github.com/sdfim/go-fhbq?status.svg)](https://godoc.org/github.com/sdfim/go-fhbq)

go-fhbq
=======
 control of recuperators FHBQ
 
 это версия скрипта python (https://github.com/sdfim/FHBQ-D) только на go

использование
-------------
соответственно аналогично

`$ go run go-fhbq.go <режим>`
наприклад: 
`$ go run go-fhbq.go n 1 auto` 
тобто нормальный мод, 1-а швидкість, байпас - авто 

перевірка поточного статусу: 
`$ go run go-fhbq.go status` 

всі доступні режими можна подивитися так: 
`$ go run go-fhbq.go -valid` 

new
---
добавлены флаги, которые можно посмотреть: 
`$ go run go-fhbq.go -help` 

есть два сниффера:
`$ ./go-fhbq.go -sniffer` 
и 
`$ ./go-fhbq.go -snifferDif` 
и дополнительным флагом -ignore
`$ ./go-fhbq.go -snifferDif -ignore` 


[![GoDoc](https://godoc.org/github.com/sdfim/go-fhbq?status.svg)](https://godoc.org/github.com/sdfim/go-fhbq)

go-fhbq
=======
 control of recuperators FHBQ
 
 это версия скрипта python (https://github.com/sdfim/FHBQ-D) только на go

usage
-------------
соответственно аналогично

`$ go run go-fhbq.go <режим>` <br>
наприклад:  <br>
`$ go run go-fhbq.go n 1 auto`  <br>
тобто нормальный мод, 1-а швидкість, байпас - авто  <br>
 <br>
перевірка поточного статусу:  <br>
`$ go run go-fhbq.go status`  <br>
 <br>
всі доступні режими можна подивитися так:  <br>
`$ go run go-fhbq.go -valid`  <br>

new
---
добавлены флаги, которые можно посмотреть:  <br>
`$ go run go-fhbq.go -help`  <br>
 <br>
есть два сниффера: <br>
`$ ./go-fhbq.go -sniffer`  <br>
и  <br>
`$ ./go-fhbq.go -snifferDif`  <br>
и с дополнительным флагом -ignore <br>
`$ ./go-fhbq.go -snifferDif -ignore`  <br>
 <br>
есть  возможность получить status в json формате: <br>
`$ ./go-fhbq.go -json status`  <br>
используя флаг -timer можно смотреть длительность исполнения: <br>
`$ ./go-fhbq.go -timer status`  <br>
`$ ./go-fhbq.go -timer n 3 auto`  <br>

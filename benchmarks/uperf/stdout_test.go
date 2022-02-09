package uperf

import (
	"encoding/json"
	"fmt"
	"testing"
)

// uperf -T -t -f -g -k -p -e -E -a -v  -m iperf.xml
var UPERF_TEST_STDOUT_ALL_ARGS string = `Error getting SSL CTX:1
Allocating shared memory of size 156624 bytes
Completed handshake phase 1
Starting handshake phase 2
Handshake phase 2 with 127.0.0.1
  Done preprocessing accepts
  Sent handshake header
  Sending workorder
    Sent workorder
    Sent transaction
    Sent flowop
    Sent transaction
    Sent flowop
    Sent transaction
    Sent flowop
TX worklist success  Sent workorder
Handshake phase 2 with 127.0.0.1 done
Completed handshake phase 2
Starting 1 threads running profile:iperf ...   0.00 seconds
TX command [UPERF_CMD_NEXT_TXN, 0] to 127.0.0.1
Txn1            0 /   1.00(s) =            0           1op/s                                                                   
TX command [UPERF_CMD_NEXT_TXN, 1] to 127.0.0.1
Txn2      63.93GB /  29.03(s) =    18.92Gb/s      288655op/s Sending signal SIGUSR2 to 140042707740224                         
called out
Txn2      66.11GB /  30.23(s) =    18.78Gb/s      286609op/s                                                                   
TX command [UPERF_CMD_NEXT_TXN, 2] to 127.0.0.1
Txn3            0 /   0.00(s) =            0           0op/s                                                                   
-------------------------------------------------------------------------------------------------------------------------------
TX command [UPERF_CMD_SEND_STATS, 0] to 127.0.0.1
** Warning: Send buffer: 100.00KB (Requested:50.00KB)  
** Warning: Recv buffer: 100.00KB (Requested:50.00KB)  
Total     66.11GB /  32.33(s) =    17.56Gb/s      267977op/s 

Group Details
-------------------------------------------------------------------------------------------------------------------------------
Group0     6.61GB /  31.23(s) =     1.82Gb/s      277423op/s 


Strand Details
-------------------------------------------------------------------------------------------------------------------------------
Thr0      66.11GB /  32.23(s) =    17.62Gb/s      268809op/s 


Txn                Count         avg         cpu         max         min 
-------------------------------------------------------------------------------------------------------------------------------
Txn0                   1    107.04us      0.00ns    107.04us    107.04us 
Txn1              866453     34.61us      0.00ns      4.11ms     22.93us 
Txn2                   1      1.00us      0.00ns      1.00us      1.00us 


Flowop             Count         avg         cpu         max         min 
-------------------------------------------------------------------------------------------------------------------------------
connect                1    106.60us      0.00ns    106.60us    106.60us 
write            8664520      3.46us      0.00ns      4.11ms     22.89us 
disconnect             1    701.00ns      0.00ns    701.00ns    701.00ns 


Netstat statistics for this run
-------------------------------------------------------------------------------------------------------------------------------
Nic       opkts/s     ipkts/s      obits/s      ibits/s
lo         267977      267977    17.62Gb/s    17.62Gb/s 
tap0            0           0     17.32b/s            0 
-------------------------------------------------------------------------------------------------------------------------------

Run Statistics
Hostname            Time       Data   Throughput   Operations      Errors
-------------------------------------------------------------------------------------------------------------------------------
127.0.0.1         32.33s    58.40GB    15.51Gb/s      7654453        0.00
master            32.33s    66.11GB    17.56Gb/s      8664523        0.00
-------------------------------------------------------------------------------------------------------------------------------
Difference(%)     -0.00%     11.66%       11.66%       11.66%       0.00%

** Warning: Send buffer: 100.00KB (Requested:50.00KB)  
** Warning: Recv buffer: 100.00KB (Requested:50.00KB)  
** [127.0.0.1] Warning: Send buffer: 100.00KB (Requested:50.00KB)  
Warning: Recv buffer: 100.00KB (Requested:50.00KB)  
  

`

// uperf -m iperf.xml
var UPERF_TEST_STDOUT_MINIMAL_ARGS string = `Error getting SSL CTX:1
Starting 1 threads running profile:iperf ...   0.00 seconds
Txn1            0 /   1.00(s) =            0           1op/s                                                                   
Txn2      67.14GB /  30.23(s) =    19.08Gb/s      291075op/s                                                                   
Txn3            0 /   0.00(s) =            0           0op/s                                                                   
-------------------------------------------------------------------------------------------------------------------------------
** Warning: Send buffer: 100.00KB (Requested:50.00KB)  
** Warning: Recv buffer: 100.00KB (Requested:50.00KB)  
Total     67.14GB /  32.34(s) =    17.84Gb/s      272158op/s 

Netstat statistics for this run
-------------------------------------------------------------------------------------------------------------------------------
Nic       opkts/s     ipkts/s      obits/s      ibits/s
lo         272158      272158    17.90Gb/s    17.90Gb/s 
-------------------------------------------------------------------------------------------------------------------------------

Run Statistics
Hostname            Time       Data   Throughput   Operations      Errors
-------------------------------------------------------------------------------------------------------------------------------
127.0.0.1         32.34s    59.48GB    15.80Gb/s      7795963        0.00
master            32.34s    67.14GB    17.84Gb/s      8800383        0.00
-------------------------------------------------------------------------------------------------------------------------------
Difference(%)     -0.00%     11.41%       11.41%       11.41%       0.00%

** Warning: Send buffer: 100.00KB (Requested:50.00KB)  
** Warning: Recv buffer: 100.00KB (Requested:50.00KB)  
** [127.0.0.1] Warning: Send buffer: 100.00KB (Requested:50.00KB)  
Warning: Recv buffer: 100.00KB (Requested:50.00KB)  
  

`

// uperf -T -t -f -g -k -p -e -E -a -v -R -m iperf.xml
var UPERF_TEST_STDOUT_ALL_ARGS_RAW string = `Error getting SSL CTX:1
Allocating shared memory of size 156624 bytes
Completed handshake phase 1
Starting handshake phase 2
Handshake phase 2 with 127.0.0.1
  Done preprocessing accepts
  Sent handshake header
  Sending workorder
    Sent workorder
    Sent transaction
    Sent flowop
    Sent transaction
    Sent flowop
    Sent transaction
    Sent flowop
TX worklist success  Sent workorder
Handshake phase 2 with 127.0.0.1 done
Completed handshake phase 2
Starting 1 threads running profile:iperf ...   0.00 seconds
TX command [UPERF_CMD_NEXT_TXN, 0] to 127.0.0.1
timestamp_ms:1644254595992.8884 name:Txn1 nr_bytes:0 nr_ops:0
timestamp_ms:1644254596993.8118 name:Txn1 nr_bytes:0 nr_ops:1

TX command [UPERF_CMD_NEXT_TXN, 1] to 127.0.0.1
timestamp_ms:1644254596993.8628 name:Txn2 nr_bytes:0 nr_ops:0
timestamp_ms:1644254597994.8152 name:Txn2 nr_bytes:2338242560 nr_ops:285430
timestamp_ms:1644254598997.6321 name:Txn2 nr_bytes:4677222400 nr_ops:570950
timestamp_ms:1644254599997.8281 name:Txn2 nr_bytes:7156449280 nr_ops:873590
timestamp_ms:1644254600998.8230 name:Txn2 nr_bytes:9516318720 nr_ops:1161660
timestamp_ms:1644254601999.8264 name:Txn2 nr_bytes:11885117440 nr_ops:1450820
timestamp_ms:1644254603000.8113 name:Txn2 nr_bytes:14316994560 nr_ops:1747680
timestamp_ms:1644254604001.8318 name:Txn2 nr_bytes:16677027840 nr_ops:2035770
timestamp_ms:1644254605002.8262 name:Txn2 nr_bytes:19024445440 nr_ops:2322320
timestamp_ms:1644254606002.8989 name:Txn2 nr_bytes:21376860160 nr_ops:2609480
timestamp_ms:1644254607003.9443 name:Txn2 nr_bytes:23732879360 nr_ops:2897080
timestamp_ms:1644254608004.8228 name:Txn2 nr_bytes:26078167040 nr_ops:3183370
timestamp_ms:1644254609005.8274 name:Txn2 nr_bytes:28437381120 nr_ops:3471360
timestamp_ms:1644254610006.8308 name:Txn2 nr_bytes:30779637760 nr_ops:3757280
timestamp_ms:1644254611007.0103 name:Txn2 nr_bytes:33119436800 nr_ops:4042900
timestamp_ms:1644254612007.8311 name:Txn2 nr_bytes:35455221760 nr_ops:4328030
timestamp_ms:1644254613008.8909 name:Txn2 nr_bytes:37793546240 nr_ops:4613470
timestamp_ms:1644254614416.5181 name:Txn2 nr_bytes:51768197120 nr_ops:6319360
timestamp_ms:1644254615417.5188 name:Txn2 nr_bytes:54116679680 nr_ops:6606040
timestamp_ms:1644254616418.5320 name:Txn2 nr_bytes:56451235840 nr_ops:6891020
timestamp_ms:1644254617419.5195 name:Txn2 nr_bytes:58805043200 nr_ops:7178350
timestamp_ms:1644254618420.5459 name:Txn2 nr_bytes:61165977600 nr_ops:7466550
timestamp_ms:1644254619421.5198 name:Txn2 nr_bytes:63519784960 nr_ops:7753880
timestamp_ms:1644254620422.5144 name:Txn2 nr_bytes:65871380480 nr_ops:8040940
timestamp_ms:1644254621423.5200 name:Txn2 nr_bytes:68204625920 nr_ops:8325760
timestamp_ms:1644254622424.5208 name:Txn2 nr_bytes:70549585920 nr_ops:8612010
timestamp_ms:1644254623425.5730 name:Txn2 nr_bytes:72889630720 nr_ops:8897660
timestamp_ms:1644254624426.6221 name:Txn2 nr_bytes:75243929600 nr_ops:9185050
timestamp_ms:1644254625427.6726 name:Txn2 nr_bytes:77596590080 nr_ops:9472240
Sending signal SIGUSR2 to 140570986473024
called out
timestamp_ms:1644254626628.8372 name:Txn2 nr_bytes:79955230720 nr_ops:9760160

TX command [UPERF_CMD_NEXT_TXN, 2] to 127.0.0.1
timestamp_ms:1644254626628.8818 name:Txn3 nr_bytes:0 nr_ops:0
timestamp_ms:1644254626628.8894 name:Txn3 nr_bytes:0 nr_ops:0

-------------------------------------------------------------------------------------------------------------------------------
TX command [UPERF_CMD_SEND_STATS, 0] to 127.0.0.1
** Warning: Send buffer: 100.00KB (Requested:50.00KB)  
** Warning: Recv buffer: 100.00KB (Requested:50.00KB)  
timestamp_ms:1644254626729.0642 name:Total nr_bytes:79955230720 nr_ops:9760161

Group Details
-------------------------------------------------------------------------------------------------------------------------------
timestamp_ms:1644254626628.9016 name:Group0 nr_bytes:7995523072 nr_ops:9760162


Strand Details
-------------------------------------------------------------------------------------------------------------------------------
timestamp_ms:1644254626628.9026 name:Thr0 nr_bytes:79955230720 nr_ops:9760163


Txn                Count         avg         cpu         max         min 
-------------------------------------------------------------------------------------------------------------------------------
Txn0                   1     69.90us      0.00ns     69.90us     69.90us 
Txn1              976017     30.11us      0.00ns 184467440ms     14.08us 
Txn2                   1      1.30us      0.00ns      1.30us      1.30us 


Flowop             Count         avg         cpu         max         min 
-------------------------------------------------------------------------------------------------------------------------------
connect                1     69.62us      0.00ns     69.62us     69.62us 
write            9760160      3.00us      0.00ns 184467440ms     22.88us 
disconnect             1      1.01us      0.00ns      1.01us      1.01us 


Netstat statistics for this run
-------------------------------------------------------------------------------------------------------------------------------
Nic       opkts/s     ipkts/s      obits/s      ibits/s
lo         307537      307537    20.22Gb/s    20.22Gb/s 
-------------------------------------------------------------------------------------------------------------------------------

Run Statistics
Hostname            Time       Data   Throughput   Operations      Errors
-------------------------------------------------------------------------------------------------------------------------------
127.0.0.1         31.74s    69.67GB    18.86Gb/s      9132343        0.00
master            31.74s    74.46GB    20.15Gb/s      9760163        0.00
-------------------------------------------------------------------------------------------------------------------------------
Difference(%)     -0.00%      6.43%        6.43%        6.43%       0.00%

** Warning: Send buffer: 100.00KB (Requested:50.00KB)  
** Warning: Recv buffer: 100.00KB (Requested:50.00KB)  
** [127.0.0.1] Warning: Send buffer: 100.00KB (Requested:50.00KB)  
Warning: Recv buffer: 100.00KB (Requested:50.00KB)  
  

`

func TestParseUperfStdout(t *testing.T) {
	for _, stdout_test := range []string{UPERF_TEST_STDOUT_ALL_ARGS, UPERF_TEST_STDOUT_ALL_ARGS_RAW, UPERF_TEST_STDOUT_MINIMAL_ARGS} {
		out, err := ParseUperfStdout(stdout_test)
		if err != nil {
			t.Errorf("Unexpected error when parsing stdout: %s", err)
		}
		marshalled, _ := json.MarshalIndent(out, "", "    ")
		fmt.Println(string(marshalled))
	}
}

// go build server.go frame.go robot.go egg.go parseConfig.go client.go web.go
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	//"encoding/hex"
	"strings"
	//"strconv"
	"bufio"
	//"strconv"
	//"time"
	"sync"
)

var (
	VERSION = "SELFBUILD" // injected by buildflags

	localAddr   = flag.String("l", "192.168.86.187:45570", "")
	verbosity   = flag.Int("v", 3, "verbosity")
	userData    = flag.String("d", "robot.txt", "data for list")
	extraData   = flag.String("ex", "extra.txt", "extra data for testing")
	eggPoolData = flag.String("egg", "egg.txt", "egg pool data")

	webAddr = flag.String("web", ":8080", "http server port, empty means not to start http server")
)

var user = NewUserInfo()

var grid = NewGrid()
var eggPool = NewEggPool()

type Session struct {
	mx   sync.Mutex
	list map[*Client]*Client
}

func (s *Session) Add(cl *Client) {
	s.mx.Lock()
	s.list[cl] = cl
	s.mx.Unlock()
}

func (s *Session) Del(cl *Client) {
	s.mx.Lock()
	delete(s.list, cl)
	s.mx.Unlock()
}

func (s *Session) Flush() {
	s.mx.Lock()
	for _, cl := range s.list {
		cl.Flush()
	}
	s.mx.Unlock()
}

var clients = &Session{
	list: make(map[*Client]*Client),
}

var PageFriends = Raw2Byte("0A 07 85 35 00 00 08 27 00 00 00 03 00 00 00 0C 00 01 00 00 00 CF C4 D1 C7 00 00 00 00 00 00 00 00 00 00 00 00 00 01 01 00 00 00 00 00 00 00 00 00 00 00 3A 23 6F 24 00 00 00 00 00 00 00 00 00 00 00 00 5E 34 00 00 0A 00 01 00 00 00 CE E1 C3 FB CB C0 CD F6 D6 AE D2 ED 00 00 00 00 00 01 02 00 00 00 00 00 00 00 00 00 00 00 2F 23 5F 24 00 00 00 00 00 00 00 00 00 00 00 00 DD 13 00 00 0A 00 01 00 00 00 4A 6F 6B 65 00 00 00 00 00 00 00 00 00 00 00 00 00 07 03 00 00 00 1C 00 00 00 00 00 00 00 2B 23 5B 24 00 00 00 00 00 00 00 00 05 00 00 00 49 FB 00 00 0A 00 01 00 00 00 45 6E 64 6A 6F 62 58 58 00 00 00 00 00 00 00 00 00 07 03 00 00 00 00 00 00 00 53 00 00 00 33 23 55 24 00 00 00 00 00 00 00 00 3A 00 00 00 4A 1C 00 00 0C 00 01 00 00 00 B0 D6 B0 D6 D4 D9 B4 F2 CE D2 D2 BB B4 CE 00 00 00 07 03 00 00 00 00 00 00 00 6F 00 00 00 33 23 55 24 00 00 00 00 00 00 00 00 00 00 00 00 69 0E 00 00 0A 00 01 00 00 00 CB C4 B4 FA D6 D8 BC DF BC DF 32 30 00 00 00 00 00 02 01 00 00 00 00 00 00 00 00 00 00 00 33 23 55 24 00 00 00 00 00 00 00 00 3A 00 00 00 02 00 00 00 05 00 01 00 00 00 B0 A2 C4 B7 C2 DE 00 00 00 00 00 00 00 00 00 00 00 01 01 00 00 00 00 00 00 00 00 00 00 00 31 23 55 24 00 00 00 00 00 00 00 00 00 00 00 00 4B 06 00 00 0A 00 01 00 00 00 CD B8 D6 A7 B5 C4 BB D8 D2 E4 00 00 00 00 00 00 00 01 01 00 00 00 00 00 00 00 00 00 00 00 33 23 55 24 00 00 00 00 00 00 00 00 3A 00 00 00 ")

func handleConn(p1 net.Conn) {
	defer p1.Close()
	buffer := make([]byte, (1<<16)+headerSize)

	Respond(p1, "version", versionPacket)

	for {
		f, err := readFrame(p1, buffer)
		if err != nil {
			return
		}

		LogIncomeMessage(f)

		switch f.cmd {
		case cmdHERT:
			// [6F][0002][03F0]02 00 F0 03 6F 1F
			Respond(p1, "heartReply", heartReply)

		case 0x076B:
			// 登入頁面ok
			// [6B][0010][03F0]0A 00 F0 03 6B 07 00 00 00 00 01 00 00 00
			buf := Raw2Byte("0A 00 F0 03 37 06 85 35 00 00 01 00 00 00")
			Respond(p1, "0x076B", buf)
			// client would follow with
			//  Hex stream: 00 00 00 00 00 00

		case cmdLOGIN: // login ID+pass
			// ID = '11111111111', PWD = 'a'
			// [29][0075][03F0]4B 00 F0 03 29 23 00 00 00 00 31 31 31 31 31 31 31 31 31 31 31 00 00 30 63 63 31 37 35 62 39 63 30 66 31 62 36 61 38 33 31 63 33 39 39 65 32 36 39 37 37 32 36 36 31 00 00 55 BC 00 00 00 79 3F 00 00 00 00 00 00 00 00 00 00 00 00 BB 00 00
			Vln(3, "[login, handling user]", "")

			// TODO: check login & get user data
			// this looks like just handles one user, since it has given this conn to class NewClient, future should write multiple user support
			client := NewClient(p1, grid)
			handleUser(client, buffer)
			return

		default:
			Vln(3, "[????]", f)
		}
	}
}

func handleUser(p1 *Client, buffer []byte) {
	defer p1.Close()

	clients.Add(p1)
	defer clients.Del(p1)

	first := true

	Respond(p1, "logData1", logData1)
	Respond(p1, "logData2", logData2)
	Respond(p1, "logData3", logData3)

	for {
		f, err := p1.ReadFrame(buffer)
		if err != nil {
			return
		}

		LogIncomeMessage(f)

		switch f.cmd {
		case cmdHERT:
			// [6F][0002][03F0]02 00 F0 03 6F 1F
			Respond(p1, "[heartReply]", heartReply)

		case 0x076B:
			buf := Raw2Byte("0A 00 F0 03 37 06 85 35 00 00 02 00 00 00")
			buf[10] = f.data[6]
			Respond(p1, "0x076B", buf)

			// 登入頁面ok
			// [076B][0010][03F0]0A 00 F0 03 6B 07 00 00 00 00 01 00 00 00

			// 進入我的房間
			// [076B][0010][03F0]0A 00 F0 03 6B 07 00 00 00 00 02 00 00 00

			// 進入任務頻道
			// [076B][0010][03F0]0A 00 F0 03 6B 07 00 00 00 00 03 00 00 00

			// 進入抽蛋
			// [076B][0010][03F0]0A 00 F0 03 6B 07 00 00 00 00 07 00 00 00

			// 進入勳章
			// [076B][0010][03F0]0A 00 F0 03 6B 07 00 00 00 00 0C 00 00 00

			// 右上角X
			// [076B][0010][03F0]0A 00 F0 03 6B 07 00 00 00 00 0A 00 00 00
			// TODO: 有資料要重傳, 會卡loading
			if f.data[6] == byte(0x0A) {
				Vf(4, "[logout]\n")
			}

		case 0x9C43:
			// [43][0010][03F0]0A 00 F0 03 43 9C 00 00 00 00 05 BC 56 C2
			p1.RespondRawFrame("logData4", logData4)

		case 0x9C49:
			// [49][0006][03F0]06 00 F0 03 49 9C 00 00 00 00
			caseReply := "08 00 F0 03 48 9C 85 35 00 00 E9 03"
			p1.RespondRawFrame("caseReply", caseReply)

		case 0x0A4D:
			// [4D][0020][03F0]14 00 F0 03 4D 0A 00 00 00 00 00 00 00 00 00 00 00 00 2F 61 A6 83 30 FD
			caseReply := "0D 00 F0 03 C7 08 00 00 00 00 7A 60 2A 3F E0 59 00"
			p1.RespondRawFrame("0x0A4d", caseReply) // 122.96.42.63:23008 ?
			//p1.WriteRawFrame("0D 00 F0 03 C7 08 00 00 00 00 7F 00 00 01 A4 0F 00") // 127.0.0.1:4004
			//p1.WriteRawFrame("0D 00 F0 03 C7 08 00 00 00 00 C0 A8 01 91 A4 0F 00") // 192.168.1.145:4004

		case 0x054F: // user IP&port (內網?)  == (0x??C7) ?
			// [4F][0012][03F0]0C 00 F0 03 4F 05 00 00 00 00 0A 08 09 E6 0B B9
			// (X) [4F][0010][03F0]0A 00 F0 03 4F 9C 00 00 00 00 01 46 2C F8
			Vln(4, "0x054F_too_long\n")
			p1.RespondFrame("0x054F", UNKNOWN_RESPONSE1)

		case 0x0740: // 初始訊息? 出擊機體
			// [0740][0006][03F0]06 00 F0 03 40 07 00 00 00 00
			garage_message := p1.GetInfo1Bytes()
			p1.RespondFrame("[0x0740 garage_page]", p1.GetInfo1Bytes())

		case 0x07E8:
			// [E8][0006][03F0]06 00 F0 03 E8 07 00 00 00 00
			p1.RespondRawFrame("[0x07E8]", "0E 00 F0 03 2F 23 85 35 00 00 00 00 00 00 01 00 00 00")
			p1.RespondRawFrame("[0x07E8]", UNKNOWN_RESPONSE2)
			//p1.WriteRawFrame("0A 00 F0 03 EE 05 85 35 00 00 00 00 00 00")

		case 0x0722:
			// [0722][0006][03F0]06 00 F0 03 22 07 00 00 00 00
			res := "0A 00 F0 03 EE 05 85 35 00 00 00 00 00 00"
			p1.WriteRawFrame(res)

		case 0x0A35:
			// user keybind initilization
			// [35][0006][03F0]06 00 F0 03 35 0A 00 00 00 00
			if user.KeyBind == nil {
				p1.RespondRawFrame("[0x0A35_default_keybind]", USER_DEFAULT_KEYBIND)
			} else {
				/// todo:change this to custome keybind after user function is dblized
				p1.RespondFrame("[0x0A35]", append(Raw2Byte("01 09 85 35 00 00"), user.KeyBind...))
			}

		case 0x062C:
			// change of keybind?
			// [062C][0058][03F0]3A 00 F0 03 2C 06 00 00 00 00 11 01 1F 01 20 01 1E 01 10 01 12 01 39 01 2A 01 1D 01 00 02 01 02 03 02 02 02 0F 01 02 01 03 01 04 01 05 01 06 01 2F 01 00 00 00 00 00 00 00 00 00 00 00 00
			keys := f.data[6:]
			if len(keys) == 52 {
				k := make([]byte, len(keys), len(keys))
				copy(k, keys)

				user.Mx.Lock()
				user.KeyBind = k
				user.Mx.Unlock()
				p1.WriteFrame(append(Raw2Byte("01 09 85 35 00 00"), user.KeyBind...))
			}
			Vf(4, "[keys]%d, [%02X]\n", len(keys), keys)

		case 0x0A2E:
			// [2E][0006][03F0]06 00 F0 03 2E 0A 00 00 00 00
			p1.RespondRawFrame("[0x0A2E]", UNKNOWN_RESPONSE3)
			//p1.WriteRawFrame("0A 00 F0 03 1F 00 4F 0F 00 00 4A 3B 40 5C")

		case 0x068E:
			// [068E][0010][03F0]0A 00 F0 03 8E 06 00 00 00 00 00 00 00 00
			res := "0E 00 F0 03 EE 08 85 35 00 00 E3 07 01 00 11 00 10 00"
			p1.RespondRawFrame("[0x068E]", res)
			//p1.WriteRawFrame(UNKNOWN_RESPONSE4)

			// logout
			// [068E][0010][03F0]0A 00 F0 03 8E 06 00 00 00 00 00 12 35 00
			// 0A 00 F0 03 8E 06 00 00 00 00 A1 39 31 00
			// 0A 00 F0 03 8E 06 00 00 00 00 B0 39 31 00
			// 0A 00 F0 03 8E 06 00 00 00 00 C1 39 31 00
			// 0A 00 F0 03 8E 06 00 00 00 00 00 12 35 00
			// 0A 00 F0 03 8E 06 00 00 00 00 10 12 35 00
			// 0A 00 F0 03 8E 06 00 00 00 00 66 4E 36 00
			// f.data[6:10]

		case 0x083E: // 好友列表, 待分析
			// [083E][0008][03F0]08 00 F0 03 3E 08 00 00 00 00 01 00 (start, p1)
			// [083E][0008][03F0]08 00 F0 03 3E 08 00 00 00 00 02 00 (p2)
			p1.RespondFrame("[friend_list]", PageFriends)
			//p1.WriteRawFrame(PAGE_FRIEND_LIST)

		case 0x0668: // 初始機格快取?
			// [0668][0014][03F0]0E 00 F0 03 68 06 00 00 00 00 01 00 00 00 00 00 00 00
			// 機體清單
			buf := p1.GetAll()
			typ := fmt.Sprintf("[all][%04d]% 02X\n", len(buf), buf)
			p1.RespondFrame(typ, buf)
			//p1.WriteAllPage() // not work

			// ret error
			p1.RespondRawFrame(typ, RET_ERROR1)
			// ret error
			p1.RespondRawFrame(typ, RET_ERROR2)
			// ret error
			p1.RespondRawFrame(typ, RET_ERROR3)
			// ret error
			p1.RespondRawFrame(typ, RET_ERROR4)
			//p1.RespondRawFrame(typ, POSSIBLE_RET_ERROR5)

		case 0x08B3:
			// not "7F 07" !!!
			//p1.WriteRawFrame("28 03 F0 03 7F 07 00 00 00 00 32 00 7A 63 01 00 E2 07 0C 00 08 00 09 00 01 00 00 00 7D 63 01 00 E2 07 0C 00 02 00 0C 00 04 00 00 00 84 63 01 00 E2 07 0C 00 0A 00 0C 00 01 00 00 00 85 63 01 00 E2 07 0C 00 05 00 13 00 01 00 00 00 87 63 01 00 E2 07 0C 00 08 00 0B 00 01 00 00 00 97 63 01 00 E2 07 0C 00 02 00 0C 00 01 00 00 00 98 63 01 00 E2 07 0C 00 02 00 13 00 01 00 00 00 99 63 01 00 E2 07 0C 00 02 00 13 00 01 00 00 00 9A 63 01 00 E2 07 0C 00 1E 00 10 00 01 00 00 00 9B 63 01 00 E2 07 0C 00 02 00 12 00 01 00 00 00 9D 63 01 00 E2 07 0C 00 02 00 13 00 01 00 00 00 9E 63 01 00 E2 07 0C 00 08 00 0B 00 01 00 00 00 9F 63 01 00 E2 07 0C 00 19 00 15 00 01 00 00 00 AC 63 01 00 E2 07 0C 00 02 00 13 00 01 00 00 00 AD 63 01 00 E2 07 0C 00 02 00 14 00 01 00 00 00 AE 63 01 00 E2 07 0C 00 02 00 14 00 01 00 00 00 AF 63 01 00 E2 07 0C 00 02 00 14 00 01 00 00 00 BA 63 01 00 E2 07 0C 00 09 00 10 00 05 00 00 00 C4 63 01 00 E2 07 0C 00 02 00 17 00 01 00 00 00 C5 63 01 00 E2 07 0C 00 03 00 08 00 05 00 00 00 C6 63 01 00 E2 07 0C 00 03 00 0B 00 0A 00 00 00 C7 63 01 00 E2 07 0C 00 04 00 14 00 05 00 00 00 C8 63 01 00 E2 07 0C 00 04 00 15 00 0A 00 00 00 C9 63 01 00 E2 07 0C 00 04 00 16 00 05 00 00 00 CA 63 01 00 E2 07 0C 00 04 00 17 00 03 00 00 00 CB 63 01 00 E2 07 0C 00 09 00 16 00 0A 00 00 00 CC 63 01 00 E2 07 0C 00 0E 00 07 00 01 00 00 00 CD 63 01 00 E2 07 0C 00 0E 00 07 00 01 00 00 00 CE 63 01 00 E2 07 0C 00 0E 00 0C 00 05 00 00 00 D5 63 01 00 E3 07 01 00 02 00 11 00 0A 00 00 00 D6 63 01 00 E3 07 01 00 03 00 15 00 0A 00 00 00 D7 63 01 00 E3 07 01 00 04 00 0E 00 0A 00 00 00 D8 63 01 00 E2 07 0C 00 08 00 0C 00 03 00 00 00 D9 63 01 00 E2 07 0C 00 08 00 15 00 03 00 00 00 DA 63 01 00 E2 07 0C 00 09 00 0C 00 03 00 00 00 DE 63 01 00 E2 07 0C 00 07 00 17 00 01 00 00 00 DF 63 01 00 E2 07 0C 00 07 00 17 00 01 00 00 00 E4 63 01 00 E3 07 01 00 03 00 09 00 14 00 00 00 E5 63 01 00 E3 07 01 00 04 00 14 00 14 00 00 00 E6 63 01 00 E3 07 01 00 06 00 0F 00 14 00 00 00 E8 63 01 00 E2 07 0C 00 0A 00 09 00 03 00 00 00 E9 63 01 00 E2 07 0C 00 0C 00 0C 00 03 00 00 00 EA 63 01 00 E2 07 0C 00 09 00 16 00 01 00 00 00 EB 63 01 00 E2 07 0C 00 09 00 16 00 01 00 00 00 EC 63 01 00 E2 07 0C 00 09 00 17 00 01 00 00 00 ED 63 01 00 E2 07 0C 00 0A 00 09 00 01 00 00 00 EE 63 01 00 E2 07 0C 00 0A 00 09 00 01 00 00 00 EF 63 01 00 E2 07 0C 00 0A 00 09 00 01 00 00 00 F0 63 01 00 E2 07 0C 00 0A 00 09 00 01 00 00 00 F1 63 01 00 E2 07 0C 00 0A 00 0C 00 01 00 00 00 ")

			// [08B3][0010][03F0]0A 00 F0 03 B3 08 00 00 00 00 00 00 00 00
			if f.data[8] == byte(0x00) {
				p1.RespondRawFrame("[0x08B3][0x00]", "16 00 F0 03 2C 07 85 35 00 00 E3 07 01 00 11 00 10 00 16 00 25 00 FE 01 00 00")
			}
			// [B3][0010][03F0]0A 00 F0 03 B3 08 00 00 00 00 F1 63 01 00
			if f.data[6] == byte(0xF1) {
				//p1.WriteRawFrame("09 00 F0 03 18 0B 85 35 00 00 03 00 00")
				p1.RespondRawFrame("[0x08B3][0xF1]", UNKNOWN_RESPONSE5)
			}

		case 0x05DB:
			// [05DB][0015][03F0]0F 00 F0 03 DB 05 00 00 00 00 02 00 00 00 00 00 00 00 00
			p1.RespondRawFrame("[0x05DB]", "09 00 F0 03 18 0B 85 35 00 00 1A 00 00")

		case 0x073C: // 機格數量(格數, 總數=24+N)
			// [073C][0006][03F0]06 00 F0 03 3C 07 00 00 00 00
			p1.RespondFrame("[0x073C]", p1.GetPageCountPack())

		case 0x0621:
			// [0621][0010][03F0]0A 00 F0 03 21 06 00 00 00 00 E1 15 00 00
			// [0621][0010][03F0]0A 00 F0 03 21 06 00 00 00 00 21 1C 00 00
			if f.data[6] == byte(0x21) {
				/*p1.WriteRawFrame(
				"28 03 F0 03 7F 07 00 00 00 00 1D 00 F2 63 01 00 E2 07 0C 00 0B 00 15 00 01 00 00 00 F3 63 01 00 E2 07 0C 00 0B 00 15 00 01 00 00 00 F6 63 01 00 E2 07 0C 00 10 00 0C 00 03 00 00 00 F7 63 01 00 E2 07 0C 00 10 00 0C 00 03 00 00 00 F8 63 01 00 E2 07 0C 00 10 00 08 00 03 00 00 00 FE 63 01 00 E2 07 0C 00 0C 00 13 00 01 00 00 00 FF 63 01 00 E2 07 0C 00 0C 00 13 00 01 00 00 00 00 64 01 00 E2 07 0C 00 0C 00 13 00 01 00 00 00 01 64 01 00 E2 07 0C 00 0C 00 11 00 01 00 00 00 02 64 01 00 E2 07 0C 00 0C 00 11 00 01 00 00 00 03 64 01 00 E2 07 0C 00 0C 00 11 00 01 00 00 00 04 64 01 00 E2 07 0C 00 0C 00 16 00 01 00 00 00 05 64 01 00 E2 07 0C 00 0C 00 17 00 01 00 00 00 06 64 01 00 E2 07 0C 00 0C 00 17 00 01 00 00 00 0A 64 01 00 E2 07 0C 00 0F 00 0C 00 02 00 00 00 0B 64 01 00 E2 07 0C 00 12 00 12 00 02 00 00 00 0C 64 01 00 E2 07 0C 00 16 00 0C 00 02 00 00 00 0D 64 01 00 E2 07 0C 00 16 00 0F 00 02 00 00 00 0E 64 01 00 E2 07 0C 00 17 00 11 00 02 00 00 00 0F 64 01 00 E2 07 0C 00 17 00 13 00 02 00 00 00 10 64 01 00 00 00 00 00 00 00 00 00 00 00 00 00 13 64 01 00 E2 07 0C 00 16 00 0B 00 03 00 00 00 15 64 01 00 E2 07 0C 00 15 00 0C 00 03 00 00 00 18 64 01 00 E2 07 0C 00 17 00 09 00 03 00 00 00 1A 64 01 00 E2 07 0C 00 1C 00 0F 00 03 00 00 00 1F 64 01 00 E3 07 01 00 03 00 09 00 03 00 00 00 24 64 01 00 E3 07 01 00 09 00 0B 00 03 00 00 00 29 64 01 00 00 00 00 00 00 00 00 00 01 00 00 00 58 64 01 00 E2 07 0C 00 0E 00 12 00 01 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 C0 A6 8A 06 00 00 00 00 00 00 00 00 C0 A6 8A 06 C0 A6 8A 06 54 93 DB 0F E4 32 33 01 80 33 33 01 B0 3C 71 01 C0 A6 8A 06 68 93 DB 0F E4 32 33 01 80 33 33 01 B0 3C 71 01 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 "+
				"37 00 F0 03 26 0A 1C AE 00 00 E4 EC 00 00 00 0F 3C 93 DB 0F 94 93 DB 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 "+
				"37 00 F0 03 26 0A 03 27 01 00 E4 EC 00 00 00 0F 3C 93 DB 0F 94 93 DB 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 "+
				"00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00") // ????*/
				p1.RespondRawFrame("[0x0621]", "48 00 f0 03 c4 0a 4a 20 00 00 07 00 85 1c 00 00 e9 1c 00 00 4d 1d 00 00 b1 1d 00 00 15 1e 00 00 79 1e 00 00 dd 1e 00 00 15 27 00 00 00 00 5e 12 00 00 00 00 99 08 00 00 00 00 00 00 78 05 00 00 00 00 00 00 38 7f 07 00 00 00 00 00")
			}

			// 0x0621 抽蛋相關?
			// [0621][0010][03F0]0A 00 F0 03 21 06 00 00 00 00 00 00 00 00
			if f.data[6] == 0x00 {
				p1.RespondRawFrame("[0x0621][0x00]", "48 00 f0 03 c4 0a 4a 20 00 00 10 00 c5 09 00 00 29 0a 00 00 8d 0a 00 00 f1 0a 00 00 55 0b 00 00 b9 0b 00 00 1d 0c 00 00 81 0c 00 00 e5 0c 00 00 49 0d 00 00 ad 0d 00 00 11 0e 00 00 75 0e 00 00 d9 0e 00 00 3d 0f 00 00 a1 0f 00 00")
			}
			// [0621][0010][03F0]0A 00 F0 03 21 06 00 00 00 00 A1 0F 00 00
			if f.data[6] == 0xA1 {
				p1.RespondRawFrame("[0x0621][0xA1]", "48 00 f0 03 c4 0a 4a 20 00 00 10 00 05 10 00 00 69 10 00 00 cd 10 00 00 31 11 00 00 95 11 00 00 f9 11 00 00 5d 12 00 00 c1 12 00 00 25 13 00 00 89 13 00 00 ed 13 00 00 51 14 00 00 b5 14 00 00 19 15 00 00 7d 15 00 00 e1 15 00 00")
			}
			// [0621][0010][03F0]0A 00 F0 03 21 06 00 00 00 00 E1 15 00 00
			if f.data[6] == 0xE1 {
				p1.RespondRawFrame("[0x0621][0xE1]", "48 00 f0 03 c4 0a 4a 20 00 00 10 00 45 16 00 00 a9 16 00 00 0d 17 00 00 71 17 00 00 d5 17 00 00 39 18 00 00 9d 18 00 00 01 19 00 00 65 19 00 00 c9 19 00 00 2d 1a 00 00 91 1a 00 00 f5 1a 00 00 59 1b 00 00 bd 1b 00 00 21 1c 00 00")
			}

			// logout?
			// [0621][0010][03F0]0A 00 F0 03 21 06 00 00 00 00 21 1C 00 00
			if f.data[6] == 0x21 {
				p1.RespondRawFrame("[0x0621][0x21]", UNKNOWN_RESPONSE6)
			}

		case 0x0A90:
			// [90][0010][03F0]0A 00 F0 03 90 0A 00 00 00 00 00 00 00 00
			if f.data[6] == byte(0x00) {
				p1.RespondRawFrame("[0x0621][0x00]", UNKNOWN_RESPONSE7)
				p1.RespondRawFrame("[0x0621][0x00]", UNKNOWN_RESPONSE8)
				p1.RespondRawFrame("[0x0621][0x00]", "06 00 F0 03 4A 9C 85 35 00 00")
			}

		case 0x0A22:
			// (?) [22][0006][03F0]06 00 F0 03 22 07 00 00 00 00
			// [22][0006][03F0]06 00 F0 03 22 0A 00 00 00 00
			p1.RespondRawFrame("[0x0A22_DATETIME]", "0E 00 F0 03 EE 08 85 35 00 00 E3 07 01 00 11 00 10 00") // 年月日小时
			//p1.WriteRawFrame("0E 00 F0 03 EE 08 A9 19 00 00 E2 07 0C 00 01 00 11 00")

		case 0x095A: // 訓練場 >> 機格page
			// [5A][0010][03F0]0A 00 F0 03 5A 09 00 00 00 00 01 00 00 00
			Vln(4, "[0x095A]", "")
			if first {
				first = false
				p1.WriteAllPage()
			} else {
				i := int(f.data[6])
				p1.WritePage(i)
			}

		//case 0x0A71: // REQ_GET_USER_UNIT_INFO: req all slot data?
		// [0A71][0014][03F0]0E 00 F0 03 71 0A 00 00 00 00 1A 14 88 00 00 00 00 00
		//p1.WriteAllPage() // not work

		// ----
		case 0x0758: // 好友搜尋
			// [58][0024][03F0]18 00 F0 03 58 07 00 00 00 00 31 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 01
			// [58][0024][03F0]18 00 F0 03 58 07 00 00 00 00 39 39 39 00 00 00 00 00 00 00 00 00 00 00 00 00 00 03
			// [58][0024][03F0]18 00 F0 03 58 07 00 00 00 00 30 31 32 33 34 35 36 37 38 39 31 32 33 34 35 36 00 10  << MAX
			userName := make([]byte, 18, 18)
			userName[0] = f.data[23]
			copy(userName[1:], f.data[6:6+17])
			typ := fmt.Sprintf("[user][%d] % 02X", userName[0], userName)
			//p1.WriteFrame(user.GetBytes2(userName))
			buf := BuildUserInfo002Pack(userName, user.SearchExp, user.SearchID)
			p1.RespondFrame(typ, buf)

		case 0x0847:
			//p1.WriteRawFrame("0E 00 F0 03 2F 23 85 35 00 00 00 00 00 00 01 00 00 00") // 年月日小时 ?
			p1.RespondRawFrame("[0x0847]", "0E 00 F0 03 AD 07 00 00 00 00 8E 23 68 24 23 00 00 00") // 发送UDP服务器和端口 ?

		case 0x0860:
			p1.RespondRawFrame("[0x0860]", "16 00 F0 03 2C 07 85 35 00 00 E3 07 01 00 0D 00 14 00 05 00 1E 00 0B 01 00 00 ") // 年月日小时分钟秒毫秒
			//p1.WriteRawFrame("0A 00 F0 03 43 06 85 35 00 00 33 C2 EB 0B") // ?
			//p1.WriteRawFrame("16 00 F0 03 2C 07 A9 19 00 00 E2 07 0C 00 01 00 11 00 0C 00 11 00 62 00 00 00")

		case 0x080E:
			//p1.WriteRawFrame("08 00 F0 03 C8 0A 85 35 00 00 02 00")
			p1.RespondRawFrame("[0x080E]", "04 00 F0 03 DA 06 00 00")
			//p1.WriteRawFrame(UNKNOWN_RESPONSE9)
			//p1.WriteRawFrame("0E 00 F0 03 35 07 85 35 00 00 00 00 00 00 00 00 00 00")
			//p1.WriteRawFrame("0E 00 F0 03 53 09 85 35 00 00 00 00 00 00 00 00 00 00")

		case 0x0869:
			// [0869][0006][03F0]06 00 F0 03 69 08 00 00 00 00
			p1.RespondRawFrame("[0x0869]", "0E 00 F0 03 35 07 85 35 00 00 00 00 00 00 00 00 00 00")

		case 0x0625:
			// [0625][0006][03F0]06 00 F0 03 25 06 00 00 00 00
			p1.RespondRawFrame("[0x0625]", "08 00 F0 03 C8 0A 85 35 00 00 02 00")

		case 0x081A:
			p1.RespondRawFrame("[0x081A]", UNKNOWN_RESPONSE10)

		case 0x05B2:
			// 有[4]"Jack"
			p1.RespondRawFrame("[0x05B2]", UNKNOWN_RESPONSE11)

		case 0x0B1F:
			p1.RespondRawFrame("[0x0B1F]", "0C 00 F0 03 51 09 85 35 00 00 85 35 00 00 01 00")

		case 0x0585:
			if f.data[6] == byte(0x03) {
				p1.WriteRawFrame(
					"09 00 F0 03 C2 0A 85 35 00 00 00 00 01 " +
						"09 00 F0 03 28 0A 85 35 00 00 03 01 01 " +
						"11 00 F0 03 1F 06 85 35 00 00 03 00 05 05 0A 14 27 03 00 00 00 " +
						"0C 00 F0 03 51 09 85 35 00 00 69 0E 00 00 01 00 " +
						"0A 00 F0 03 1F 00 18 0B 00 00 98 29 3B 5C")
			}

		case 0x0020:
			p1.RespondRawFrame("[0x0020]", UNKNOWN_RESPONSE12)

		case 0x07C0:
			p1.RespondRawFrame("[0x07C0]",
				"08 00 F0 03 26 07 85 35 00 00 00 00 "+
					"0A 00 F0 03 8C 06 85 35 00 00 02 00 00 00")

		case 0x0A07:
			// ?
			p1.RespondRawFrame("[0x0A07]",
				"07 00 F0 03 6D 09 00 00 00 00 00 "+
					"07 00 F0 03 6D 09 00 00 00 00 01 "+
					"07 00 F0 03 6D 09 00 00 00 00 02 "+
					"07 00 F0 03 6D 09 00 00 00 00 03 "+
					"07 00 F0 03 6D 09 00 00 00 00 04")

		case 0x0A05:
			p1.RespondRawFrame("[0x0A05]",
				"08 00 F0 03 6B 06 85 35 00 00 00 00 "+
					"0A 00 F0 03 D1 05 85 35 00 00 02 00 00 00")

		case 0x0705:
			p1.RespondRawFrame("[0x0705]",
				"08 00 F0 03 6B 06 85 35 00 00 00 00 "+
					"0A 00 F0 03 D1 05 85 35 00 00 02 00 00 00")

		case 0x0744:
			if f.data[6] == byte(0xCB) { // 進房間 人員/機體列表?
				p1.RespondRawFrame("[0x0744][0xCB]", UNKNOWN_RESPONSE13)
			} else {
				p1.WriteRawFrame("06 00 F0 03 4A 9C 85 35 00 00")
			}

		case 0x9C4C:
			//UNKNOWN_COMMENT1
			p1.RespondRawFrame("[0x9C4C]", "0E 00 F0 03 4D 9C 85 35 00 00 F0 BD 80 00 0B 00 00 00")

		case 0x0AD3: // 對戰心跳包
			p1.RespondRawFrame("[0x0AD3]", "0A 00 F0 03 66 0A 85 35 00 00 85 35 00 00")

		case 0x08B7:
			p1.RespondRawFrame("[0x08B7]", "0a 00 f0 03 83 07 4a 20 00 00 08 00 00 00")

		case 0x0756:
			p1.RespondRawFrame("[0x0756]", "06 00 f0 03 22 06 4a 20 00 00")

			// --- 抽蛋
		case 0x085C:
			// [5C][0010][03F0]0A 00 F0 03 5C 08 00 00 00 00 00 00 00 00
			p1.RespondRawFrame("[0x085C]", "48 00 f0 03 28 07 4a 20 00 00 0c 00 65 00 00 00 c9 00 00 00 2d 01 00 00 91 01 00 00 f5 01 00 00 59 02 00 00 bd 02 00 00 21 03 00 00 85 03 00 00 e9 03 00 00 4d 04 00 00 b1 04 00 00 4a 00 00 00 00 00 00 00 48 00 b0 06 03 00 00 00")
			//p1.WriteRawFrame(
			//"0E 00 F0 03 EE 08 A9 19 00 00 E2 07 0C 00 01 00 11 00 " +
			//"16 00 F0 03 2C 07 A9 19 00 00 E2 07 0C 00 01 00 11 00 0C 00 11 00 62 00 00 00")

		case 0x071E, 0x05B0, 0x0817: // 抽蛋(代幣, GP, 自訂)
			// (代幣) [1E][0010][03F0]0A 00 F0 03 1E 07 00 00 00 00 62 09 00 00
			// (GP)   [B0][0010][03F0]0A 00 F0 03 B0 05 00 00 00 00 61 09 00 00
			// (自訂) [17][0042][03F0]2A 00 F0 03 17 08 85 35 00 00 03 00 AE 3A 00 00 A7 3A 00 00 9B 3A 00 00 91 3E 00 00 93 3E 00 00 95 3E 00 00 00 00 00 00 00 00 00 00 01 08
			out := eggPool.GetOne()
			bot := p1.AddNew(out.ID, out.C)
			pos := uint16(1)
			if bot != nil {
				pos = bot.Pos
			}
			Vln(4, "[egg]", out, pos)
			p1.WriteFrame(BuildEggPack(out, user.GP, pos))

			// force update grid
			p1.Flush()
			first = true

			// --- logout
		case 0x0A97:
			// [0A97][0010][03F0]0A 00 F0 03 97 0A 00 00 00 00 00 00 00 00
			p1.RespondRawFrame("[0x0A97]", "48 00 F0 03 28 07 A9 19 00 00 0C 00 65 00 00 00 C9 00 00 00 2D 01 00 00 91 01 00 00 F5 01 00 00 59 02 00 00 BD 02 00 00 21 03 00 00 85 03 00 00 E9 03 00 00 4D 04 00 00 B1 04 00 00 50 E8 B3 06 50 E8 B3 06 E5 03 00 00 02 00 52 0A")

		case 0x060C: // 設定出擊機體
			// [060C][0014][03F0]0E 00 F0 03 0C 06 00 00 00 00 [5C 27 30 01 00 00 00 00]
			// [060C][0014][03F0]0E 00 F0 03 0C 06 00 00 00 00 [19 F3 99 01 00 00 00 00] ???
			uuid := binary.LittleEndian.Uint64(f.data[6:14])
			Vf(5, "[setGO]UUID = %02X\n", uuid)
			p1.SetGoUUID(uuid)
			bot := p1.GetGo()
			if bot == nil {
				p1.GetPos(1)
			}
			buf := bot.GetBytes()
			buf = append(Raw2Byte("72 05 00 00 00 00 00 00 "), buf...)
			p1.RespondFrame("[0x060C 設定出擊機體]", buf)

		//case 0x05BC: // REQ_DELETE_USERUNIT: 刪除機體
		// [05BC][0019][03F0]13 00 F0 03 BC 05 00 00 00 00 [0F 00 DE AD 00 00 00 00] 3A 56 00 00 01 換GP
		// [05BC][0019][03F0]13 00 F0 03 BC 05 00 00 00 00 [0F 00 DE AD 00 00 00 00] 3A 56 00 00 02 換副官F

		default:
			//Vln(3, "[old]", f)
			oldFormat(p1, f)
		}
	}
}

func oldFormat(p1 *Client, f Frame) {
	cmd := uint8(f.cmd & 0xFF)
	switch cmd {
	case 0xC7:
		Vln(3, "[??C7]", f)
		// == 0x054F ?

	case 0x47:
		p1.RespondRawFrame("[??47]",
			"0E 00 F0 03 2F 23 85 35 00 00 00 00 00 00 01 00 00 00 "+
				"0E 00 F0 03 AD 07 00 00 00 00 8D 23 61 24 4C 00 00 00")

	default:
		Vln(3, "[????]", f)
		// 萬用包(時間: 年 月 日 時 分 秒 ms)
		p1.RespondRawFrame("[????_response]", "16 00 F0 03 2C 07 98 6D 00 00  E2 07  0C 00  16 00  0F 00  2A 00  22 00  EF 01 00 00")
	}
}

func main() {
	// set log
	log.SetFlags(log.Ldate | log.Ltime)
	// parse input
	flag.Parse()

	Vln(1, "[server] version =", VERSION)

	readyCh := make(chan struct{}, 1)
	go func() {
		stdin := bufio.NewReader(os.Stdin)
		var err error
		cmd := "L"
		//readData()

		for {
			err = reloadConfig(cmd)
			if err != nil {
				Vf(1, "Read Data Error: %v\n\n", err)

			} else {
				select {
				case readyCh <- struct{}{}:
				default:
				}
			}

			cmd, err = stdin.ReadString('\n')
			if err != nil {
				break
			}
			cmd = strings.Trim(cmd, "\n\r\t ")
			Vln(4, "[cmd]", cmd)
		}
	}()

	<-readyCh

	go webStart(*webAddr)
	srvStart()
}

func reloadConfig(cmd string) error {
	var err error
	switch cmd {
	case "L":
		readExtra()
		readEggPool()
		err = readData()

		//buf := grid.GetPage(1)
		//Vf(4, "[dbg][%d][%v]\n", len(buf), buf)

		if err != nil {
			Vf(1, "[config][load]Load Data Error: %v\n\n", err)
			return err
		}
		// force update
		clients.Flush()
		Vln(3, "[config][load]")

	case "R":
		// force update
		clients.Flush()
		Vln(3, "[config][flush]")

	case "S":
		err = saveData()
		Vln(3, "[config][save]", err)
	}
	return nil
}

func srvStart() {
	ln, err := net.Listen("tcp", *localAddr)
	if err != nil {
		Vln(2, "[server]Error listening:", err)
		return
	}
	defer ln.Close()

	Vf(2, "Listening: %v\n\n", *localAddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConn(conn)
	}
}

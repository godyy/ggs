package client

import (
	"encoding/json"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/chzyer/readline"
	prototypes "github.com/godyy/ggs/internal/proto/types"
)

type cmd struct {
	name          string                                      // 命令名称
	desc          string                                      // 描述
	usage         string                                      // 用法
	autoCompleter *readline.PrefixCompleter                   // 自动补全
	exec          func(c *cmd, cli *Client, args string) bool // 执行逻辑
}

var (
	cmdList []*cmd
	cmdMap  map[string]*cmd
)

func registerCmd(c ...*cmd) {
	for _, c := range c {
		if _, ok := cmdMap[c.name]; ok {
			panic("cmd " + c.name + " already registered!")
		}

		cmdList = append(cmdList, c)
		cmdMap[c.name] = c
	}
}

func cmdAllUsage() {
	log.Println("commands: name..desc..usage")
	for _, cmd := range cmdList {
		log.Printf("\t%s\t\t%s.\t\t%s", cmd.name, cmd.desc, cmd.usage)
	}
}

func cmdUsage(c *cmd) {
	log.Println("usage: " + c.usage)
}

func init() {
	cmdMap = make(map[string]*cmd)
	registerCmd(
		&cmd{
			name:          "help",
			desc:          "print commands",
			autoCompleter: readline.PcItem("help"),
			exec: func(_ *cmd, c *Client, args string) bool {
				cmdAllUsage()
				return false
			},
		},
		&cmd{
			name:          "exit",
			desc:          "exit client",
			autoCompleter: readline.PcItem("exit"),
			exec: func(_ *cmd, c *Client, args string) bool {
				return true
			},
		},
		&cmd{
			name: "sendreq",
			desc: "send request message to server",
			// usage:         "sendreq msgname[Req]" + cmdSendReqArgsSp + "msgjsonbody",
			usage:         `sendreq msgname {"key1":value1[,"key2":value2,...]}`,
			autoCompleter: readline.PcItem("sendreq"),
			exec: func(c *cmd, cli *Client, args string) bool {
				// parts := strings.Split(args, cmdSendReqArgsSp)
				// if len(parts) < 2 {
				// 	cmdUsage(c)
				// 	return false
				// }

				parts := cmdSendReqArgsRegex.FindStringSubmatch(args)
				if len(parts) != 3 {
					cmdUsage(c)
					return false
				}

				msg := parts[1]
				body := parts[2]
				log.Printf("name:%s\t json:%s", msg, body)
				if !strings.HasSuffix(msg, "Req") {
					msg = msg + "Req"
				}

				req, _, err := prototypes.C2S.CreateByName(msg)
				if err != nil {
					log.Println(err)
					return false
				}
				err = json.Unmarshal([]byte(body), req)
				if err != nil {
					log.Println(err)
					return false
				}

				resp, err := cli.sendReq(req)
				if err != nil {
					log.Println(err)
					return false
				}

				log.Printf("%s:{%+v}", reflect.TypeOf(resp).Elem().Name(), resp)
				return false
			},
		},
	)
}

var (
	cmdSendReqArgsRegex = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9_]*)\s+({.*})$`)
)

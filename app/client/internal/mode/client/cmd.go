package client

import (
	"encoding/json"
	"log"
	"reflect"
	"regexp"
	"strings"

	protoreg "github.com/godyy/ggs/internal/protocol/registry"
)

type cmdExecFunc func(c *cmd, cli *Client, args string) bool

type cmdExec func(cli *Client, args string) bool

type cmd struct {
	name  string      // 命令名称
	desc  string      // 描述
	usage string      // 用法
	exec  cmdExecFunc // 执行逻辑
}

func (c *cmd) execute(cli *Client, args string) bool {
	return c.exec(c, cli, args)
}

var (
	cmdList    []*cmd
	cmdExecMap map[string]cmdExec
	cmdNameAll []string
)

const cmdNameSep = "|"

func registerCmd(c ...*cmd) {
	for _, c := range c {
		names := strings.Split(c.name, cmdNameSep)
		for _, name := range names {
			if _, ok := cmdExecMap[name]; ok {
				panic("cmd " + name + " already registered!")
			}
			cmdExecMap[name] = c.execute
		}

		cmdList = append(cmdList, c)
		cmdNameAll = append(cmdNameAll, names...)
	}
}

func getCmdNameAll() []string {
	return cmdNameAll
}

func getCmdExec(name string) func(cli *Client, args string) bool {
	return cmdExecMap[name]
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
	cmdExecMap = make(map[string]cmdExec)
	registerCmd(
		&cmd{
			name: "help|h|?",
			desc: "print commands",
			exec: func(_ *cmd, c *Client, args string) bool {
				cmdAllUsage()
				return false
			},
		},
		&cmd{
			name: "exit|quit|q",
			desc: "exit client",
			exec: func(_ *cmd, c *Client, args string) bool {
				return true
			},
		},
		&cmd{
			name: "sendreq",
			desc: "send request message to server",
			// usage:         "sendreq msgname[Req]" + cmdSendReqArgsSp + "msgjsonbody",
			usage: `sendreq msgname {"key1":value1[,"key2":value2,...]}`,
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

				req, _, err := protoreg.Registry.C2S.CreateByName(msg)
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

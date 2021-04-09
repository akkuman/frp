// Copyright 2021 akkuman, akkumans@qq.com
// 混淆login参数，增加命令行启动socks5

package sub

import (
	"fmt"
	"os"

	"github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/consts"
	"github.com/spf13/cobra"
)

func init() {

	RegisterCommonFlags(modSocks5Cmd)

	modSocks5Cmd.PersistentFlags().StringVarP(&proxyName, "proxy_name", "n", "", "代理名称")
	modSocks5Cmd.PersistentFlags().IntVarP(&remotePort, "remote_port", "r", 0, "远端socks5代理服务监听端口")

	rootCmd.AddCommand(modSocks5Cmd)
}

var modSocks5Cmd = &cobra.Command{
	Use:   "socks5",
	Short: "一键创建socks5代理",
	Long: `一键创建socks5代理

	样例: frpc socks5 -p tcp -s 127.0.0.1:7000 -n 666xxx -r 61234 -t this_is_token --hb 2 --tls_enable

	说明: 其中需要关注的只有以上几个参数，必填的参数有 -r(--remote_port)，其他参数的默认值请查看help帮助文档

	其中需要额外说明的参数是
	1. -n(--proxy_name)，默认的代理名称生成规则是 god_{remote_port}，可以通过设置这个参数来进行覆盖
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		clientCfg, err := parseClientCommonCfg(CfgFileTypeCmd, nil)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// 基础配置预处理
		if clientCfg.Protocol == "ws" {
			clientCfg.Protocol = "websocket"
		} else if clientCfg.Protocol == "wss" {
			fmt.Println("不支持wss协议")
			os.Exit(1)
		}

		cfg := &config.TCPProxyConf{}
		cfg.ProxyName = proxyName
		if cfg.ProxyName == "" {
			cfg.ProxyName = fmt.Sprintf("god_%d", remotePort)
		}
		cfg.ProxyType = consts.TCPProxy
		cfg.Plugin = "socks5"
		cfg.RemotePort = remotePort
		cfg.UseCompression = true

		err = cfg.CheckForCli()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		proxyConfs := map[string]config.ProxyConf{
			cfg.ProxyName: cfg,
		}
		err = startService(clientCfg, proxyConfs, nil, "")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return nil
	},
}

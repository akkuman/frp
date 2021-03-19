// Copyright 2021 akkuman, akkumans@qq.com
// 通过websocket与服务端建立连接达到域前置隐藏

package sub

import (
	"fmt"
	"os"

	"github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/consts"
	frpnet "github.com/fatedier/frp/pkg/util/net"
	"github.com/spf13/cobra"
)

func init() {

	RegisterCommonFlags(modCDNCmd)

	modCDNCmd.PersistentFlags().StringVarP(&proxyName, "proxy_name", "n", "", "代理名称")
	modCDNCmd.PersistentFlags().StringVarP(&cdnSuffix, "cdn_suf", "f", ".cdn.dnsv1.com", "需要添加到回源域名后的 CDN 域名后缀")
	modCDNCmd.PersistentFlags().StringVarP(&cdnSourceHost, "cdn_source", "d", "", "CDN 的回源 Host 域名")
	modCDNCmd.PersistentFlags().IntVarP(&rport, "rport", "r", 61234, "远端socks5代理服务监听端口")

	rootCmd.AddCommand(modCDNCmd)
}

var modCDNCmd = &cobra.Command{
	Use:   "cdnhide",
	Short: "采用域前置方案隐藏真实ip，创建socks5代理",
	Long: `采用域前置方案隐藏真实ip，创建socks5代理

	样例: frpc cdnhide -d api.ding.com -f .cdn.dnsv1.com -p wss -n 666xxx -r 61234 -t this_is_token --hb 2

	说明: 其中需要关注的只有以上几个参数，必填的参数有 -d(--cdn_source)，其他参数的默认值请查看help帮助文档

	其中需要额外说明的参数是
	1. -n(--proxy_name)，默认的代理名称生成规则是 god_{remote_port}，可以通过设置这个参数来进行覆盖
	2. -s(--server_addr)，不用理会这个参数，当在这个运行模式下，会根据你传入的是ws(websocket)还是wss区分是80还是443端口，而主机将会采用 cdn_source + cdn_suf
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		clientCfg, err := parseClientCommonCfg(CfgFileTypeCmd, nil)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// 基础配置预处理
		if clientCfg.Protocol != "ws" && clientCfg.Protocol != "wss" && clientCfg.Protocol != "websocket" {
			fmt.Println("protocol必须为ws、wss 或 websocket")
			os.Exit(1)
		}
		if clientCfg.Protocol == "ws" {
			clientCfg.Protocol = "websocket"
		}

		clientCfg.ServerAddr = cdnSourceHost + cdnSuffix
		if clientCfg.Protocol == "wss" {
			clientCfg.ServerPort = 443
		} else {
			clientCfg.ServerPort = 80
		}
		frpnet.SetWebsocketConfig(cdnSourceHost, protocol == "wss")

		cfg := &config.TCPProxyConf{}
		cfg.ProxyName = proxyName
		if cfg.ProxyName == "" {
			cfg.ProxyName = fmt.Sprintf("god_%d", rport)
		}
		fmt.Println(rport)
		cfg.ProxyType = consts.TCPProxy
		cfg.Plugin = "socks5"
		cfg.RemotePort = rport
		cfg.UseEncryption = true
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

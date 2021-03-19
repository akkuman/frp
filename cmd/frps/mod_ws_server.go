// Copyright 2021 akkuman, akkumans@qq.com
// 修改一些websocket服务端的配置达到隐藏的效果

package main

import "github.com/akkuman/websocket"

func init() {
	// if strings.ToLower(req.Header.Get("Upgrade")) != "websocket" ||
	// 	!strings.Contains(strings.ToLower(req.Header.Get("Connection")), "upgrade") {
	// 	return http.StatusBadRequest, ErrNotWebSocket
	// }
	websocket.ErrNotWebSocket = &websocket.ProtocolError{"404 not found"}
}

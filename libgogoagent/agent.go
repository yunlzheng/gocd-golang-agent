package libgogoagent

import (
	"os"
	"runtime"
	"time"
)

var (
	uuid          = "564e9408-fb78-4856-4215-52e0-e14bb056"
	serverHost    = "localhost"
	sslPort       = "8154"
	httpPort      = "8153"
	hostname, _   = os.Hostname()
	workingDir, _ = os.Getwd()
)

func sslHostAndPort() string {
	return serverHost + ":" + sslPort
}

func httpsServerURL(path string) string {
	return "https://" + sslHostAndPort() + path
}

func httpServerURL(path string) string {
	return "http://" + serverHost + ":" + httpPort + path
}

func wsServerURL() string {
	return "wss://" + GoServerDN() + ":8154/go/agent-websocket"
}

func StartAgent() {
	ReadGoServerCACert()
	Register(map[string]string{
		"hostname":                      hostname,
		"uuid":                          uuid,
		"location":                      workingDir,
		"operatingSystem":               runtime.GOOS,
		"usablespace":                   "5000000000",
		"agentAutoRegisterKey":          "",
		"agentAutoRegisterResources":    "",
		"agentAutoRegisterEnvironments": "",
		"agentAutoRegisterHostname":     "",
		"elasticAgentId":                "",
		"elasticPluginId":               "",
	})

	send, received := StartWebsocket(wsServerURL(), httpsServerURL("/"))

	buildSession := BuildSession{
		HttpClient: GoServerRemoteClient(true),
		Send:       send}

	go func() {
		for {
			msg := MakeMessage("ping",
				"com.thoughtworks.go.server.service.AgentRuntimeInfo",
				AgentRuntimeInfo())
			send <- msg
			time.Sleep(10 * time.Second)
		}
	}()

	for {
		msg := <-received
		switch msg.Action {
		case "setCookie":
			str, _ := msg.Data["data"].(string)
			SetState("cookie", str)
		case "cmd":
			SetState("runtimeStatus", "Building")
			command, _ := msg.Data["data"].(map[string]interface{})
			err := buildSession.Process(MakeBuildCommand(command))
			SetState("runtimeStatus", "Idle")
			if err != nil {
				LogInfo("Error(%v) when processing message : %v", err, msg)
			}
		}
	}
}
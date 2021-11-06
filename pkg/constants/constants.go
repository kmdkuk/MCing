package constants

// Metadata
const (
	MetaPrefix = "mcing.kmdkuk.com/"
	Finalizer  = MetaPrefix + "finalizer"

	LabelAppInstance  = "app.kubernetes.io/instance"
	LabelAppName      = "app.kubernetes.io/name"
	LabelAppComponent = "app.kubernetes.io/component"
	LabelAppCreatedBy = "app.kubernetes.io/created-by"

	AppName            = "mcing"
	AppComponentServer = "server"
	ControllerName     = "mcing-controller"
)

// Container
const (
	MinecraftContainerName = "minecraft"
	ServerPortName         = "server-port"
	ServerPort             = int32(25565)
	RconPortName           = "rcon-port"
	RconPort               = int32(25575)
	DataVolumeName         = "minecraft-data"
	DataPath               = "/data"
	ServerPropsName        = "server.properties"
	ServerPropsPath        = DataPath + "/" + ServerPropsName
	BanIPName              = "banned-ips.json"
	BanIPPath              = DataPath + "/" + BanIPName
	BanPlayerName          = "banned-players.json"
	BanPlayerPath          = DataPath + "/" + BanPlayerName
	OpsName                = "ops.json"
	OpsPath                = DataPath + "/" + OpsName
	WhiteListName          = "whitelist.json"
	WhiteListPath          = DataPath + "/" + WhiteListName
	ConfigVolumeName       = "config"
	ConfigPath             = "/mcing-config"

	AgentContainerName = "mcing-agent"
	AgentPort          = int32(9080)
	AgentPortName      = "agent-port"
	DefaultAgentImage  = ImagePrefix + "mcing-agent:0.0.3"

	InitContainerName  = "mcing-init"
	ImagePrefix        = "ghcr.io/kmdkuk/"
	InitContainerImage = "mcing-init"
	InitCommand        = "mcing-init"

	DefaultServerImage = "itzg/minecraft-server:java8"
)

const (
	EulaEnvName = "EULA"
)

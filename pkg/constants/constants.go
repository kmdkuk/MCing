package constants

// Metadata.
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

// Container.
const (
	MinecraftContainerName = "minecraft"
	ServerPortName         = "server-port"
	ServerPort             = int32(25565)
	InternalServerPort     = int32(25566)
	RconPortName           = "rcon-port"
	RconPort               = int32(25575)
	RconDefaultPassword    = "minecraft" // TODO: dont use
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

	LazymcVolumeName       = "lazymc"
	LazymcConfigVolumeName = "lazymc-config"
	LazymcPath             = "/opt/lazymc"
	LazymcConfigName       = "lazymc.toml"
	LazymcBinName          = "lazymc"
	LazymcLicenseName      = "LICENSE"

	AgentContainerName = "mcing-agent"
	AgentPort          = int32(9080)
	AgentPortName      = "agent-port"

	InitContainerName  = "mcing-init"
	ImagePrefix        = "ghcr.io/kmdkuk/"
	InitContainerImage = "mcing-init"
	InitCommand        = "mcing-init"

	DefaultServerImage = "itzg/minecraft-server:java8"
)

const (
	// EulaEnvName is the environment variable name for EULA.
	EulaEnvName = "EULA"
	// RconPasswordEnvName is the environment variable name for RCON password.
	RconPasswordEnvName = "RCON_PASSWORD"
	// RconPasswordSecretKey is the secret key for RCON password.
	RconPasswordSecretKey = "rcon-password"
)

// server.properties.
const (
	WhitelistProps = "white-list"
	RconPortProps  = "rcon.port"
)

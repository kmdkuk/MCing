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

const (
	MinecraftContainerName = "minecraft"
	ServerPortName         = "server-port"
	ServerPort             = int32(25565)
	RconPortName           = "rcon-port"
	RconPort               = int32(25575)
	DataVolumeName         = "minecraft-data"
	DataPath               = "/data"

	DefaultServerImage = "itzg/minecraft-server:java17"
)

const (
	EulaEnvName = "EULA"
)

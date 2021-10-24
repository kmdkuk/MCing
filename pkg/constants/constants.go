package constants

const (
	MetaPrefix = "mcing.kmdkuk.com/"
	Finalizer  = MetaPrefix + "finalizer"

	DefaultServerImage  = "itzg/minecraft-server:java17"
	ServerContainerName = "minecraft"

	LabelAppInstance  = "app.kubernetes.io/instance"
	LabelAppName      = "app.kubernetes.io/name"
	LabelAppComponent = "app.kubernetes.io/component"
	LabelAppCreatedBy = "app.kubernetes.io/created-by"

	AppName            = "mcing"
	AppComponentServer = "server"
	ControllerName     = "mcing-controller"
)

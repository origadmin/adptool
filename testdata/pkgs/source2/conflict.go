package sourcepkg2

const MaxRetries = 5

var ConfigValue = "sourcepkg2-config"

type CommonStruct struct {
	ID int
}

func CommonFunction() string {
	return "sourcepkg2-common"
}
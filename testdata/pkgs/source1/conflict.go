package sourcepkg

const MaxRetries = 3

var ConfigValue = "sourcepkg-config"

type CommonStruct struct {
	Name string
}

func CommonFunction() string {
	return "sourcepkg-common"
}
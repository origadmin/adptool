package aliaspkg

import (
	_ "github.com/origadmin/adptool/testdata/sourcepkg3"
)

//go:adapter:package github.com/origadmin/adptool/testdata/sourcepkg
//go:adapter:package github.com/origadmin/adptool/testdata/duplicate/sourcepkg
//go:adapter:package github.com/origadmin/adptool/testdata/sourcepkg2
//go:adapter:package github.com/origadmin/adptool/testdata/duplicate/sourcepkg2
//go:adapter:package github.com/origadmin/adptool/testdata/sourcepkg3 custompkg
//go:adapter:package github.com/origadmin/adptool/testdata/duplicate/sourcepkg3
//go:adapter:package github.com/origadmin/adptool/testdata/source-pkg4
//go:adapter:package github.com/origadmin/adptool/testdata/duplicate/pkg1
//go:adapter:package github.com/origadmin/adptool/testdata/duplicate/pkg2
//go:adapter:package github.com/origadmin/adptool/testdata/duplicate/pkg3
//go:adapter:package log/slog

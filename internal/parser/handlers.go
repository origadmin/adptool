package parser

import (
	"strings"

	"github.com/origadmin/adptool/internal/config"
)

// --- Top-Level Handlers ---

func handleDefaultDirective(builder *ConfigBuilder, d *Directive) error {
	if len(d.SubCmds) < 1 || d.SubCmds[0] != "mode" || len(d.SubCmds) < 2 {
		return newDirectiveError(d, "invalid defaults Directive format. Expected 'default:mode:<field> <value>'")
	}
	if builder.config.Defaults == nil {
		builder.config.Defaults = &config.Defaults{}
	}
	if builder.config.Defaults.Mode == nil {
		builder.config.Defaults.Mode = &config.Mode{}
	}
	modeField := d.SubCmds[1]
	switch modeField {
	case "strategy":
		builder.config.Defaults.Mode.Strategy = d.Argument
	case "prefix":
		builder.config.Defaults.Mode.Prefix = d.Argument
	case "suffix":
		builder.config.Defaults.Mode.Suffix = d.Argument
	case "explicit":
		builder.config.Defaults.Mode.Explicit = d.Argument
	case "regex":
		builder.config.Defaults.Mode.Regex = d.Argument
	case "ignores":
		builder.config.Defaults.Mode.Ignores = d.Argument
	default:
		return newDirectiveError(d, "unknown defaults mode field '%s'", modeField)
	}
	return nil
}

func handlePropDirective(builder *ConfigBuilder, d *Directive) error {
	if len(d.SubCmds) != 0 {
		return newDirectiveError(d, "invalid prop directive format. Expected 'prop <name> <value>'")
	}
	name, value, err := parseNameValue(d.Argument)
	if err != nil {
		return newDirectiveError(d, "invalid prop directive argument: %v", err)
	}
	entry := &config.PropsEntry{Name: name, Value: value}
	builder.config.Props = append(builder.config.Props, entry)
	return nil
}

func handlePackageDirective(builder *ConfigBuilder, d *Directive) error {
	if len(d.SubCmds) == 0 {
		pkgParts := strings.SplitN(d.Argument, " ", 2)
		var pkg *config.Package
		if len(pkgParts) == 2 {
			pkg = &config.Package{Import: pkgParts[0], Alias: pkgParts[1]}
		} else {
			pkg = &config.Package{Import: d.Argument}
		}
		builder.SetPackageScope(pkg)
		return nil
	}

	if builder.currentPackage == nil {
		return newDirectiveError(d, "'package:%s' must follow a 'package' directive", d.SubCmds[0])
	}

	switch d.SubCmds[0] {
	case "alias":
		builder.currentPackage.Alias = d.Argument
	case "path":
		builder.currentPackage.Path = d.Argument
	case "prop":
		name, value, err := parseNameValue(d.Argument)
		if err != nil {
			return newDirectiveError(d, "invalid package prop argument: %v", err)
		}
		builder.currentPackage.Props = append(builder.currentPackage.Props, &config.PropsEntry{Name: name, Value: value})
	case "type":
		return handleTypeDirective(builder, d.SubCmds[1:], d.Argument, d)
	case "function", "func":
		return handleFuncDirective(builder, d.SubCmds[1:], d.Argument, d)
	case "variable", "var":
		return handleVarDirective(builder, d.SubCmds[1:], d.Argument, d)
	case "constant", "const":
		return handleConstDirective(builder, d.SubCmds[1:], d.Argument, d)
	default:
		return newDirectiveError(d, "unknown package sub-directive '%s'", d.SubCmds[0])
	}
	return nil
}

func handleTypeDirective(builder *ConfigBuilder, subCmds []string, argument string, d *Directive) error {
	if len(subCmds) == 0 {
		builder.AddTypeRule(argument)
		return nil
	}
	return builder.ApplySubDirective(subCmds, argument)
}

func handleFuncDirective(builder *ConfigBuilder, subCmds []string, argument string, d *Directive) error {
	if len(subCmds) == 0 {
		builder.AddFuncRule(argument)
		return nil
	}
	return builder.ApplySubDirective(subCmds, argument)
}

func handleVarDirective(builder *ConfigBuilder, subCmds []string, argument string, d *Directive) error {
	if len(subCmds) == 0 {
		builder.AddVarRule(argument)
		return nil
	}
	return builder.ApplySubDirective(subCmds, argument)
}

func handleConstDirective(builder *ConfigBuilder, subCmds []string, argument string, d *Directive) error {
	if len(subCmds) == 0 {
		builder.AddConstRule(argument)
		return nil
	}
	return builder.ApplySubDirective(subCmds, argument)
}

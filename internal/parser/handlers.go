package parser

import (
	"fmt"

	"github.com/origadmin/adptool/internal/config"
)

// --- Top-Level Handlers ---

// handleDefaultDirective for the default directive
// Example:
//go:adapter:default:mode:strategy replace
//go:adapter:default:mode:prefix append
//go:adapter:default:mode:suffix append
//go:adapter:default:mode:explicit merge
//go:adapter:default:mode:regex merge
//go:adapter:default:mode:ignores merge
func handleDefaultDirective(defaults *config.Defaults, directive *Directive) error {
	if defaults.Mode == nil {
		defaults.Mode = &config.Mode{}
	}
	switch directive.BaseCmd {
	case "strategy":
		defaults.Mode.Strategy = directive.Argument
	case "prefix":
		defaults.Mode.Prefix = directive.Argument
	case "suffix":
		defaults.Mode.Suffix = directive.Argument
	case "explicit":
		defaults.Mode.Explicit = directive.Argument
	case "regex":
		defaults.Mode.Regex = directive.Argument
	case "ignores":
		defaults.Mode.Ignores = directive.Argument
	default:
		return fmt.Errorf("unrecognized directive '%s' for Defaults", directive.BaseCmd)
	}
	return nil
}

// handlePropDirective for the prop directive
// Example:
//go:adapter:property GlobalVar1 globalValue1
//go:adapter:property GlobalVar2 globalValue2

func handlePropDirective(directive *Directive) ([]*config.PropsEntry, error) {
	if directive.Argument == "" {
		return nil, fmt.Errorf("props directive requires an argument (key value)")
	}
	name, value, err := parseNameValue(directive.Argument)
	if err != nil {
		return nil, newDirectiveError(directive, "invalid prop directive argument: %v", err)
	}
	entry := &config.PropsEntry{Name: name, Value: value}
	return []*config.PropsEntry{entry}, nil
}

//func handlePackageDirective(builder *ConfigBuilder, d *Directive) error {
//	if len(d.SubCmds) == 0 {
//		pkgParts := strings.SplitN(d.Argument, " ", 2)
//		var pkg *config.Package
//		if len(pkgParts) == 2 {
//			pkg = &config.Package{Import: pkgParts[0], Alias: pkgParts[1]}
//		} else {
//			pkg = &config.Package{Import: d.Argument}
//		}
//		builder.SetPackageScope(pkg)
//		return nil
//	}
//
//	if builder.currentPackage == nil {
//		return newDirectiveError(d, "'package:%s' must follow a 'package' directive", d.SubCmds[0])
//	}
//
//	switch d.SubCmds[0] {
//	case "alias":
//		builder.currentPackage.Alias = d.Argument
//	case "path":
//		builder.currentPackage.Path = d.Argument
//	case "prop":
//		name, value, err := parseNameValue(d.Argument)
//		if err != nil {
//			return newDirectiveError(d, "invalid package prop argument: %v", err)
//		}
//		builder.currentPackage.Props = append(builder.currentPackage.Props, &config.PropsEntry{Name: name, Value: value})
//	case "type":
//		return handleTypeDirective(builder, d.SubCmds[1:], d.Argument, d)
//	case "function", "func":
//		return handleFuncDirective(builder, d.SubCmds[1:], d.Argument, d)
//	case "variable", "var":
//		return handleVarDirective(builder, d.SubCmds[1:], d.Argument, d)
//	case "constant", "const":
//		return handleConstDirective(builder, d.SubCmds[1:], d.Argument, d)
//	default:
//		return newDirectiveError(d, "unknown package sub-directive '%s'", d.SubCmds[0])
//	}
//	return nil
//}
//
//func handleTypeDirective(builder *ConfigBuilder, subCmds []string, argument string, d *Directive) error {
//	if len(subCmds) == 0 {
//		builder.AddTypeRule(argument)
//		return nil
//	}
//	return builder.ApplySubDirective(subCmds, argument, d)
//}
//
//func handleFuncDirective(builder *ConfigBuilder, subCmds []string, argument string, d *Directive) error {
//	if len(subCmds) == 0 {
//		builder.AddFuncRule(argument)
//		return nil
//	}
//	return builder.ApplySubDirective(subCmds, argument, d)
//}
//
//func handleVarDirective(builder *ConfigBuilder, subCmds []string, argument string, d *Directive) error {
//	if len(subCmds) == 0 {
//		builder.AddVarRule(argument)
//		return nil
//	}
//	return builder.ApplySubDirective(subCmds, argument, d)
//}
//
//func handleConstDirective(builder *ConfigBuilder, subCmds []string, argument string, d *Directive) error {
//	if len(subCmds) == 0 {
//		builder.AddConstRule(argument)
//		return nil
//	}
//	return builder.ApplySubDirective(subCmds, argument, d)
//}

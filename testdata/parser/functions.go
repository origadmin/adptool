package parser

// Test global function directives
//go:adapter:function MyGlobalFunc
//go:adapter:function:disabled true
//go:adapter:function:rename NewRenamedFunc
//go:adapter:function:regex ^old(.*)$=new$1
//go:adapter:function:strategy replace

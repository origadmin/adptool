package parser

// Test global variable directives
//go:adapter:variable MyGlobalVar
//go:adapter:variable:disabled true
//go:adapter:variable:rename NewRenamedVar
//go:adapter:variable:regex ^old(.*)$=new$1
//go:adapter:variable:strategy replace

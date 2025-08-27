package pkg2

// @adptool(include)

type SameType struct{}

func (s *SameType) SameMethod() {}

func SameFunction() {}

const SameConstant = 1

var SameVariable = "hello"

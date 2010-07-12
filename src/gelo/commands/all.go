package commands

import "gelo"

var All = []map[string]interface{}{
	LogicCommands, MathCommands, ListCommands, TypePredicates, IOCommands,
	StringCommands, DictCommands, PortCommands, CombinatorCommands,
	CopyCommands, ControlCommands, ErrorCommands, RegexpCommands,
	EvalCommands, ArgParserCommands, VariableCommands, Values,
}

var Values = map[string]interface{}{
	"true":  gelo.True,
	"false": gelo.False,
}

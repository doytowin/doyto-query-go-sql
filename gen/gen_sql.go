package gen

import (
	"bytes"
	"github.com/doytowin/goooqo/rdb"
	"go/ast"
	"reflect"
	"strings"
)

type SqlGenerator struct {
	*generator
}

func init() {
	sqlOpMap := make(map[string]operator)
	sqlOpMap["Eq"] = operator{name: "Eq", sign: "="}
	sqlOpMap["Ne"] = operator{name: "Ne", sign: "<>"}
	sqlOpMap["Not"] = operator{name: "Not", sign: "!="}
	sqlOpMap["Gt"] = operator{name: "Gt", sign: ">"}
	sqlOpMap["Ge"] = operator{name: "Ge", sign: ">="}
	sqlOpMap["Lt"] = operator{name: "Lt", sign: "<"}
	sqlOpMap["Le"] = operator{name: "Le", sign: "<="}
	sqlOpMap["In"] = operator{name: "In", sign: "IN", format: "conditions = append(conditions, \"%s%s\"+strings.Repeat(\"?\", len(*q.%s)))"}
	sqlOpMap["NotIn"] = operator{name: "NotIn", sign: "NOT IN", format: "conditions = append(conditions, \"%s%s\" + strings.Repeat(\"?\", len(*q.%s)))"}
	sqlOpMap["Null"] = operator{name: "Null", sign: "IS NULL", format: "conditions = append(conditions, \"%s %s\")"}
	sqlOpMap["NotNull"] = operator{name: "NotNull", sign: "IS NOT NULL", format: "conditions = append(conditions, \"%s %s\")"}
	sqlOpMap["Like"] = operator{name: "Like", sign: "LIKE", format: "conditions = append(conditions, \"%s %s ?\")"}
	opMap["sql"] = sqlOpMap
}

func NewSqlGenerator() *SqlGenerator {
	return &SqlGenerator{&generator{
		Buffer:     bytes.NewBuffer(make([]byte, 0, 1024)),
		key:        "sql",
		imports:    []string{`"github.com/doytowin/goooqo/rdb"`, `"strings"`},
		bodyFormat: "conditions = append(conditions, \"%s %s ?\")",
		ifFormat:   "if q.%s%s {",
	}}
}

func (g *SqlGenerator) appendBuildMethod(ts *ast.TypeSpec) {
	g.writeInstruction("func (q %s) BuildConditions() ([]string, []any) {", ts.Name)
	g.appendFuncBody(ts)
	g.writeInstruction("}")
}

func (g *SqlGenerator) appendFuncBody(ts *ast.TypeSpec) {
	g.intent = "\t"
	g.writeInstruction("conditions := make([]string, 0, 4)")
	g.writeInstruction("args := make([]any, 0, 4)")
	for _, field := range ts.Type.(*ast.StructType).Fields.List {
		if field.Names != nil {
			g.appendCondition(field, field.Names[0].Name)
		}
	}
	g.writeInstruction("return conditions, args")
	g.intent = ""
}

func (g *SqlGenerator) appendStruct(stp *ast.StructType) {
	for _, field := range stp.Fields.List {
		if field.Names != nil {
			g.appendCondition(field, field.Names[0].Name)
		}
	}
}

func (g *SqlGenerator) appendCondition(field *ast.Field, fieldName string) {
	column, op := g.suffixMatch(fieldName)

	stp := toStructPointer(field)
	if stp != nil && field.Tag != nil {
		g.appendIfStartNil(fieldName)
		tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
		if subqueryTag, ok := tag.Lookup("subquery"); ok {
			fpSubquery := rdb.BuildSubquery(subqueryTag, fieldName)
			g.appendIfBody("whereClause, args1 := rdb.BuildWhereClause(q.%s)", fieldName)
			g.appendIfBody("condition := \"" + fpSubquery.Subquery() + "\" + whereClause + \")\"")
			g.appendIfBody("conditions = append(conditions, condition)")
			g.appendIfBody("args = append(args, args1...)")
		}
	} else if strings.Contains(op.sign, "NULL") {
		g.appendIfStart(fieldName, "")
		g.appendIfBody(op.format, column, op.sign)
	} else if strings.Contains(op.sign, "IN") {
		g.appendIfStartNil(fieldName)
		g.appendIfBody(op.format, column, op.sign, fieldName)
		g.appendIfBody("args = append(args, q.%s)", fieldName)
	} else {
		g.appendIfStartNil(fieldName)
		g.appendIfBody(op.format, column, op.sign)
		g.appendIfBody("args = append(args, q.%s)", fieldName)
	}
	g.appendIfEnd()
}

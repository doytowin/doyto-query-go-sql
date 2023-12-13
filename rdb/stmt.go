package rdb

import (
	. "github.com/doytowin/goquery/core"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

var whereId = " WHERE id = ?"

type EntityMetadata[E any] struct {
	TableName       string
	ColStr          string
	fieldsWithoutId []string
	createStr       string
	placeholders    string
	updateStr       string
}

func readId(entity any) any {
	rv := reflect.ValueOf(entity)
	value := rv.FieldByName("Id")
	readValue := ReadValue(value)
	return readValue
}

func (em *EntityMetadata[E]) buildArgs(entity E) []any {
	var args []any

	rv := reflect.ValueOf(entity)
	for _, col := range em.fieldsWithoutId {
		value := rv.FieldByName(col)
		args = append(args, ReadValue(value))
	}
	return args
}

func (em *EntityMetadata[E]) buildSelect(query GoQuery) (string, []any) {
	whereClause, args := BuildWhereClause(query)
	s := "SELECT " + em.ColStr + " FROM " + em.TableName + whereClause
	pageQuery := query.GetPageQuery()
	if pageQuery.NeedPaging() {
		s += pageQuery.BuildPageClause()
	}
	logrus.Debug("SQL: ", s)
	logrus.Debug("ARG: ", args)
	return s, args
}

func (em *EntityMetadata[E]) buildSelectById() string {
	return "SELECT " + em.ColStr + " FROM " + em.TableName + whereId
}

func (em *EntityMetadata[E]) buildCount(query GoQuery) (string, []any) {
	whereClause, args := BuildWhereClause(query)
	s := "SELECT count(0) FROM " + em.TableName + whereClause

	logrus.Debug("SQL: ", s)
	return s, args
}

func (em *EntityMetadata[E]) buildDeleteById() string {
	return "DELETE FROM " + em.TableName + whereId
}

func (em *EntityMetadata[E]) buildDelete(query any) (string, []any) {
	whereClause, args := BuildWhereClause(query)
	s := "DELETE FROM " + em.TableName + whereClause
	logrus.Debug("SQL: " + s)
	return s, args
}

func (em *EntityMetadata[E]) buildCreate(entity E) (string, []any) {
	return em.createStr, em.buildArgs(entity)
}

func (em *EntityMetadata[E]) buildCreateMulti(entities []E) (string, []any) {
	var args []any
	for _, entity := range entities {
		args = append(args, em.buildArgs(entity)...)
	}
	createStr := em.createStr
	for i := 1; i < len(entities); i++ {
		createStr += ", " + em.placeholders
	}
	return createStr, args
}

func (em *EntityMetadata[E]) buildUpdate(entity E) (string, []any) {
	args := em.buildArgs(entity)
	args = append(args, readId(entity))
	return em.updateStr, args
}

func (em *EntityMetadata[E]) buildPatch(entity E) (string, []any) {
	var args []any
	sqlStr := "UPDATE " + em.TableName + " SET "

	rv := reflect.ValueOf(entity)
	for _, col := range em.fieldsWithoutId {
		value := rv.FieldByName(col)
		v := ReadValue(value)
		if v != nil {
			sqlStr += UnCapitalize(col) + " = ?, "
			args = append(args, v)
		}
	}
	return sqlStr[0 : len(sqlStr)-2], args
}

func (em *EntityMetadata[E]) buildPatchById(entity E) (string, []any) {
	sqlStr, args := em.buildPatch(entity)
	sqlStr = sqlStr + whereId
	args = append(args, readId(entity))
	logrus.Info("PATCH SQL: ", sqlStr)
	return sqlStr, args
}

func (em *EntityMetadata[E]) buildPatchByQuery(entity E, query GoQuery) ([]any, string) {
	patchClause, argsE := em.buildPatch(entity)
	whereClause, argsQ := BuildWhereClause(query)

	args := append(argsE, argsQ...)
	sqlStr := patchClause + whereClause

	logrus.Debug("PATCH SQL: ", sqlStr)
	return args, sqlStr
}

func buildEntityMetadata[E any](entity any) EntityMetadata[E] {
	refType := reflect.TypeOf(entity)
	columns := make([]string, refType.NumField())
	var columnsWithoutId []string
	var fieldsWithoutId []string
	for i := 0; i < refType.NumField(); i++ {
		field := refType.Field(i)
		columns[i] = UnCapitalize(field.Name)
		if field.Name != "Id" {
			fieldsWithoutId = append(fieldsWithoutId, field.Name)
			columnsWithoutId = append(columnsWithoutId, UnCapitalize(field.Name))
		}
	}
	var tableName string
	v, ok := entity.(Entity)
	if ok {
		tableName = v.GetTableName()
	} else {
		tableName = strings.TrimSuffix(refType.Name(), "Entity")
	}

	placeholders := "(?"
	for i := 1; i < len(columnsWithoutId); i++ {
		placeholders += ", ?"
	}
	placeholders += ")"
	createStr := "INSERT INTO " + tableName +
		" (" + strings.Join(columnsWithoutId, ", ") + ") " +
		"VALUES " + placeholders
	logrus.Debug("CREATE SQL: ", createStr)

	set := make([]string, len(columnsWithoutId))
	for i, col := range columnsWithoutId {
		set[i] = col + " = ?"
	}
	updateStr := "UPDATE " + tableName + " SET " + strings.Join(set, ", ") + whereId
	logrus.Debug("UPDATE SQL: ", updateStr)

	return EntityMetadata[E]{
		TableName:       tableName,
		ColStr:          strings.Join(columns, ", "),
		fieldsWithoutId: fieldsWithoutId,
		createStr:       createStr,
		placeholders:    placeholders,
		updateStr:       updateStr,
	}
}
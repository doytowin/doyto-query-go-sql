package rdb

import (
	"database/sql"
	. "github.com/doytowin/goquery/core"
	log "github.com/sirupsen/logrus"
	"reflect"
)

type Connection interface {
	Prepare(query string) (*sql.Stmt, error)
}

type RelationalDataAccess[C Connection, E any] struct {
	em     EntityMetadata[E]
	create func() E
}

func BuildRelationalDataAccess[E Entity](createEntity func() E) DataAccess[Connection, E] {
	return &RelationalDataAccess[Connection, E]{
		em:     buildEntityMetadata[E](createEntity()),
		create: createEntity,
	}
}

func (da *RelationalDataAccess[C, E]) Get(conn C, id any) (*E, error) {
	sqlStr := da.em.buildSelectById()
	rows, err := da.doQuery(conn, sqlStr, []any{id})
	if len(rows) == 1 {
		return &rows[0], err
	}
	return nil, err
}

func (da *RelationalDataAccess[C, E]) Query(conn C, query GoQuery) ([]E, error) {
	sqlStr, args := da.em.buildSelect(query)
	return da.doQuery(conn, sqlStr, args)
}

func (da *RelationalDataAccess[C, E]) doQuery(conn C, sqlStr string, args []any) ([]E, error) {
	log.Debug("SQL: ", sqlStr)
	log.Debug("ARG: ", args)

	result := []E{}

	entity := da.create()
	elem := reflect.ValueOf(&entity).Elem()
	length := elem.NumField()
	pointers := make([]any, length)
	for i := range pointers {
		pointers[i] = elem.Field(i).Addr().Interface()
	}

	stmt, err := conn.Prepare(sqlStr)
	if NoError(err) {
		defer Close(stmt)
		var rows *sql.Rows
		rows, err = stmt.Query(args...)
		for NoError(err) && rows.Next() {
			err = rows.Scan(pointers...)
			if err == nil {
				result = append(result, entity)
			}
		}
		defer Close(rows)
	}

	return result, err
}

func (da *RelationalDataAccess[C, E]) Count(conn C, query GoQuery) (int64, error) {
	var cnt int64
	sqlStr, args := da.em.buildCount(query)
	stmt, err := conn.Prepare(sqlStr)
	if NoError(err) {
		defer Close(stmt)
		row := stmt.QueryRow(args...)
		err = row.Scan(&cnt)
	}
	return cnt, err
}

func (da *RelationalDataAccess[C, E]) Page(conn C, query GoQuery) (PageList[E], error) {
	var cnt int64
	data, err := da.Query(conn, query)
	if NoError(err) {
		cnt, err = da.Count(conn, query)
	}
	return PageList[E]{List: data, Total: cnt}, err
}

func (da *RelationalDataAccess[C, E]) Delete(conn C, id any) (int64, error) {
	sqlStr := da.em.buildDeleteById()
	return parse(da.doUpdate(conn, sqlStr, []any{id}))
}

func (da *RelationalDataAccess[C, E]) DeleteByQuery(conn C, query any) (int64, error) {
	sqlStr, args := da.em.buildDelete(query)
	return parse(da.doUpdate(conn, sqlStr, args))
}

func (da *RelationalDataAccess[C, E]) doUpdate(conn C, sqlStr string, args []any) (sql.Result, error) {
	stmt, err := conn.Prepare(sqlStr)
	if NoError(err) {
		defer Close(stmt)
		return stmt.Exec(args...)
	}
	return nil, err
}

func (da *RelationalDataAccess[C, E]) Create(conn C, entity *E) (int64, error) {
	sqlStr, args := da.em.buildCreate(*entity)
	result, err := da.doUpdate(conn, sqlStr, args)
	var id int64
	if NoError(err) {
		id, err = result.LastInsertId()
		if NoError(err) {
			elem := reflect.ValueOf(entity).Elem()
			elem.FieldByName("Id").SetInt(id)
		}
	}
	return id, err
}

func (da *RelationalDataAccess[C, E]) CreateMulti(conn C, entities []E) (int64, error) {
	if len(entities) == 0 {
		return 0, nil
	}
	sqlStr, args := da.em.buildCreateMulti(entities)
	return parse(da.doUpdate(conn, sqlStr, args))
}

func (da *RelationalDataAccess[C, E]) Update(conn C, entity E) (int64, error) {
	sqlStr, args := da.em.buildUpdate(entity)
	return parse(da.doUpdate(conn, sqlStr, args))
}

func (da *RelationalDataAccess[C, E]) Patch(conn C, entity E) (int64, error) {
	sqlStr, args := da.em.buildPatchById(entity)
	return parse(da.doUpdate(conn, sqlStr, args))
}

func (da *RelationalDataAccess[C, E]) PatchByQuery(conn C, entity E, query GoQuery) (int64, error) {
	args, sqlStr := da.em.buildPatchByQuery(entity, query)
	return parse(da.doUpdate(conn, sqlStr, args))
}

func parse(result sql.Result, err error) (int64, error) {
	if NoError(err) {
		return result.RowsAffected()
	}
	return 0, err
}
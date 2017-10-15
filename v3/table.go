package v3

import (
	"bytes"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/parser"
)

type columnIndex uint

type column struct {
	index  bitmapIndex
	name   string
	tables []string
}

func (c column) hasColumn(tableName, colName string) bool {
	if colName != c.name {
		return false
	}
	if tableName == "" {
		return true
	}
	return c.hasTable(tableName)
}

func (c column) hasTable(tableName string) bool {
	for _, t := range c.tables {
		if t == tableName {
			return true
		}
	}
	return false
}

func (c column) resolvedName(tableName string) *parser.ColumnItem {
	if tableName == "" {
		if len(c.tables) > 0 {
			tableName = c.tables[0]
		}
	}
	return &parser.ColumnItem{
		TableName: parser.TableName{
			TableName:               parser.Name(tableName),
			DBNameOriginallyOmitted: true,
		},
		ColumnName: parser.Name(c.name),
	}
}

func (c column) newVariableExpr(tableName string) *expr {
	e := &expr{
		op:   variableOp,
		body: c.resolvedName(tableName),
	}
	e.inputVars.set(c.index)
	e.updateProperties()
	return e
}

func findColumn(cols []column, name string) column {
	for _, col := range cols {
		if col.name == name {
			return col
		}
	}
	return column{}
}

// TODO(peter): adds keys, unique keys, and foreign keys. track NOT NULL and
// column value constraints.
type table struct {
	name    string
	columns []column
}

func (t *table) String() string {
	var buf bytes.Buffer
	for i, col := range t.columns {
		if i > 0 {
			buf.WriteString(" ")
		}
		if tables := col.tables; len(tables) > 1 {
			buf.WriteString("{")
			for j, table := range tables {
				if j > 0 {
					buf.WriteString(",")
				}
				buf.WriteString(table)
			}
			buf.WriteString("}")
		} else if len(tables) == 1 {
			buf.WriteString(tables[0])
		} else {
			buf.WriteString(t.name)
		}
		buf.WriteString(".")
		buf.WriteString(col.name)
		buf.WriteString(":")
		fmt.Fprintf(&buf, "%d", col.index)
	}
	return buf.String()
}

func concatTable(left, right *table) *table {
	t := &table{
		columns: make([]column, len(left.columns)+len(right.columns)),
	}
	copy(t.columns[:], left.columns)
	copy(t.columns[len(left.columns):], right.columns)
	return t
}

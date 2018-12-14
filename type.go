package dbstruct

import (
	"reflect"
	"strings"
	"time"
)

var (
	Bit       = "BIT"
	TinyInt   = "TINYINT"
	SmallInt  = "SMALLINT"
	MediumInt = "MEDIUMINT"
	Int       = "INT"
	Integer   = "INTEGER"
	BigInt    = "BIGINT"

	Enum = "ENUM"
	Set  = "SET"

	Char             = "CHAR"
	Varchar          = "VARCHAR"
	NVarchar         = "NVARCHAR"
	TinyText         = "TINYTEXT"
	Text             = "TEXT"
	NText            = "NTEXT"
	Clob             = "CLOB"
	MediumText       = "MEDIUMTEXT"
	LongText         = "LONGTEXT"
	Uuid             = "UUID"
	UniqueIdentifier = "UNIQUEIDENTIFIER"
	SysName          = "SYSNAME"

	Date       = "DATE"
	DateTime   = "DATETIME"
	Time       = "TIME"
	TimeStamp  = "TIMESTAMP"
	TimeStampz = "TIMESTAMPZ"

	Decimal = "DECIMAL"
	Numeric = "NUMERIC"

	Real   = "REAL"
	Float  = "FLOAT"
	Double = "DOUBLE"

	Binary     = "BINARY"
	VarBinary  = "VARBINARY"
	TinyBlob   = "TINYBLOB"
	Blob       = "BLOB"
	MediumBlob = "MEDIUMBLOB"
	LongBlob   = "LONGBLOB"
	Bytea      = "BYTEA"

	Bool    = "BOOL"
	Boolean = "BOOLEAN"

	Serial    = "SERIAL"
	BigSerial = "BIGSERIAL"

	Json  = "JSON"
	Jsonb = "JSONB"
)

// default sql type change to go types
func SQLType2Type(dbField DbField) reflect.Type {

	strTypes := strings.Split(dbField.Type, " ")

	dbType := strTypes[0]

	if i := strings.Index(dbType, "("); i > 0 {
		dbType = dbType[0:i]
	}

	var ret reflect.Type

	name := strings.ToUpper(dbType)
	switch name {
	case Bit, TinyInt, SmallInt, MediumInt, Int, Integer, Serial:
		ret = reflect.TypeOf(1)
	case BigInt, BigSerial:
		ret = reflect.TypeOf(int64(1))
	case Float, Real:
		ret = reflect.TypeOf(float32(1))
	case Double:
		ret = reflect.TypeOf(float64(1))
	case Char, Varchar, NVarchar, TinyText, Text, NText, MediumText, LongText, Enum, Set, Uuid, Clob, SysName:
		ret = reflect.TypeOf("")
	case TinyBlob, Blob, LongBlob, Bytea, Binary, MediumBlob, VarBinary, UniqueIdentifier:
		ret = reflect.TypeOf([]byte{})
	case Bool:
		ret = reflect.TypeOf(true)
	case DateTime, Date, Time, TimeStamp, TimeStampz:
		ret = reflect.TypeOf(time.Time{})
	case Decimal, Numeric:
		ret = reflect.TypeOf("")
	default:
		ret = reflect.TypeOf("")
	}

	if dbField.Null == "YES" {
		ret = reflect.PtrTo(ret)
	}

	return ret
}

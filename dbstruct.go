package dbstruct

import (
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx"
)

type NameMapFunc func(string) string
type TaggerFunc func(dbName, fieldName string) reflect.StructTag

type Options struct {
	NameMap NameMapFunc
	Tagger  TaggerFunc
	DSN     string
	Driver  string
}

type Option func(*Options)

func DSN(dsn string) Option {
	return func(o *Options) {
		o.DSN = dsn
	}
}

func Driver(driver string) Option {
	return func(o *Options) {
		o.Driver = driver
	}
}

func Tagger(tagger TaggerFunc) Option {
	return func(o *Options) {
		o.Tagger = tagger
	}
}

func NameMapper(nameMap NameMapFunc) Option {
	return func(o *Options) {
		o.NameMap = nameMap
	}
}

type DbField struct {
	Field   string      `db:"Field"`
	Type    string      `db:"Type"`
	Null    string      `db:"Null"`
	Key     string      `db:"Key"`
	Default interface{} `db:"Default"`
	Extra   string      `db:"Extra"`

	Name      string
	StructTag reflect.StructTag
}

func (p *DbField) GoType() reflect.Type {
	return SQLType2Type(*p)
}

type DbTable struct {
	Name   string
	Fields []DbField

	typ reflect.Type
}

func (p *DbTable) init() {

	var fields []reflect.StructField
	for i := 0; i < len(p.Fields); i++ {
		field := reflect.StructField{
			Name: p.Fields[i].Name,
			Type: p.Fields[i].GoType(),
			Tag:  p.Fields[i].StructTag,
		}
		fields = append(fields, field)
	}

	structTyp := reflect.StructOf(fields)

	p.typ = structTyp
}

func (p *DbTable) NewStruct() (v interface{}, err error) {

	val := reflect.New(p.typ)

	if !val.IsValid() {
		err = fmt.Errorf("create struct of %s failure", p.Name)
		return
	}

	v = val.Interface()

	return
}

func (p *DbTable) NewStructSlice() (v interface{}, err error) {

	val := reflect.MakeSlice(reflect.SliceOf(p.typ), 0, 0)

	if !val.IsValid() {
		err = fmt.Errorf("create struct slice of %s failure", p.Name)
		return
	}

	v = val.Interface()

	return
}

type DBStruct struct {
	Options Options
}

func New(opts ...Option) (dbStruct *DBStruct, err error) {

	options := Options{}

	for _, o := range opts {
		o(&options)
	}

	dbStruct = &DBStruct{
		Options: options,
	}

	return
}

func (p *DBStruct) Describe(tableName string) (tb DbTable, err error) {

	db, err := sqlx.Connect(p.Options.Driver, p.Options.DSN)

	if err != nil {
		return
	}

	defer db.Close()

	fields := []DbField{}

	err = db.Select(&fields, "DESCRIBE "+tableName)
	if err != nil {
		return
	}

	for i := 0; i < len(fields); i++ {
		fields[i].Name = fields[i].Field
		if p.Options.NameMap != nil {
			fields[i].Name = p.Options.NameMap(fields[i].Field)
		}

		if p.Options.Tagger != nil {
			fields[i].StructTag = p.Options.Tagger(tableName, fields[i].Field)
		}
	}

	tb = DbTable{
		Name:   tableName,
		Fields: fields,
	}

	tb.init()

	return
}

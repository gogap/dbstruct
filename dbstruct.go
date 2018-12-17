package dbstruct

import (
	"fmt"
	"github.com/google/uuid"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
)

type NameMapperFunc func(string) string
type TaggerFunc func(dbName, fieldName string) reflect.StructTag
type TypeMapperFunc func(dbName, fieldName, dbFieldType string) reflect.Type

type Options struct {
	NameMap        NameMapperFunc
	Tagger         TaggerFunc
	TypeMapper     TypeMapperFunc
	DSN            string
	Driver         string
	CreateTableDSN string
}

type Option func(*Options)

func DataSource(dsn, driver string) Option {
	return func(o *Options) {
		o.DSN = driver
		o.Driver = dsn
	}
}

func CreateTabelDSN(dsn string) Option {
	return func(o *Options) {
		o.CreateTableDSN = dsn
	}
}

func Tagger(tagger TaggerFunc) Option {
	return func(o *Options) {
		o.Tagger = tagger
	}
}

func TypeMapper(typeMapper TypeMapperFunc) Option {
	return func(o *Options) {
		o.TypeMapper = typeMapper
	}
}

func NameMapper(nameMap NameMapperFunc) Option {
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
	GoType    reflect.Type
	StructTag reflect.StructTag
}

type DbTable struct {
	Name   string
	Fields []DbField

	typ reflect.Type
}

func (p *DbTable) FieldByName(name string) (field DbField, exist bool) {

	for i := 0; i < len(p.Fields); i++ {
		if p.Fields[i].Name == name {
			field = p.Fields[i]
			exist = true
			return
		}
	}
	return
}

func (p *DbTable) UpdateField(name string, field DbField) (err error) {

	for i := 0; i < len(p.Fields); i++ {
		if p.Fields[i].Name == name {
			p.Fields[i] = field
			p.rebuild()
		}
	}

	err = fmt.Errorf("field of %s not exist", name)

	return
}

func (p *DbTable) rebuild() {

	var fields []reflect.StructField
	for i := 0; i < len(p.Fields); i++ {
		field := reflect.StructField{
			Name: p.Fields[i].Name,
			Type: p.Fields[i].GoType,
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

	tableName = strings.TrimSpace(tableName)

	if len(tableName) == 0 {
		err = fmt.Errorf("the describe table name is empty")
		return
	}

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

		if p.Options.TypeMapper != nil {
			fields[i].GoType = p.Options.TypeMapper(tableName, fields[i].Field, fields[i].Type)
		} else {
			fields[i].GoType = SQLType2Type(fields[i])
		}
	}

	tb = DbTable{
		Name:   tableName,
		Fields: fields,
	}

	tb.rebuild()

	return
}

func (p *DBStruct) DescribeQuery(query string) (tb DbTable, err error) {

	query = strings.TrimSpace(query)

	if len(query) == 0 {
		err = fmt.Errorf("the describe query is empty")
		return
	}

	if len(p.Options.CreateTableDSN) == 0 {
		err = fmt.Errorf("the CreateTableDSN option must be set")
		return
	}

	db, err := sqlx.Connect(p.Options.Driver, p.Options.CreateTableDSN)

	if err != nil {
		return
	}

	defer db.Close()

	query = strings.TrimSuffix(query, ";")

	if strings.Index(query, ";") >= 0 {
		err = fmt.Errorf("could not have more than one sql")
		return
	}

	limitQuery := query

	upperQuery := strings.ToUpper(query)
	if strings.Contains(upperQuery, "LIMIT") {
		limitQuery = updateLimit(query)
	} else {
		limitQuery += " LIMIT 0 "
	}

	tableName := "temp_" + strings.Replace(uuid.New().String(), "-", "", -1)

	createTableSQL := fmt.Sprintf("CREATE TABLE `%s` AS %s", tableName, limitQuery)

	_, err = db.Exec(createTableSQL)

	if err != nil {
		return
	}

	defer func() {
		deleteTableSQL := fmt.Sprintf("DROP TABLE `%s`", tableName)
		db.Exec(deleteTableSQL)
	}()

	tb, err = p.Describe(tableName)

	return
}

func updateLimit(query string) string {

	fields := strings.Fields(query)

	for i := 0; i < len(fields); i++ {
		if strings.ToUpper(fields[i]) == "LIMIT" {
			if isUnClosedWithSpace(fields[i+1]) {
				fields[i+1] = ""
				fields[i+2] = newLimitArgs(fields[i+2])
				i += 2
			} else {
				fields[i+1] = newLimitArgs(fields[i+1])
				i++
			}
		}
	}

	return strings.Join(fields, " ")
}

func isUnClosedWithSpace(args string) bool {
	return args[len(args)-1] == ','
}

func newLimitArgs(args string) string {

	i := strings.Index(args, ")")
	if i == -1 {
		return "0"
	}

	return "0" + args[i:]
}

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	"github.com/ssrathi/go-attr"
	"gorm.io/gorm"
)

type (
	table struct {
		name    string
		entry   interface{}
		propMap propMapping
		joins   map[string]string
	}
	tableResponse struct {
		Data tableResponseData  `json:"data"`
		Meta *tableResponseMeta `json:"meta,omitempty"`
	}
	property struct {
		Table string
		Name  string
	}
	tableResponseData []map[string]interface{}
	tableResponseMeta struct {
		Total          int64       `json:"total"`
		Offset         int         `json:"offset"`
		Limit          int         `json:"limit"`
		Count          int         `json:"count"`
		AvailableProps propSlice   `json:"available_props"`
		IncludedProps  propSlice   `json:"included_props"`
		Mapping        propMapping `json:"mapping"`
	}
	propSlice   []property
	propMapping map[property]string
)

func parseProperty(s string, table string) property {
	if s == "" {
		return property{}
	}
	l := strings.SplitN(s, ".", 2)
	if len(l) == 2 {
		return property{Table: l[0], Name: l[1]}
	}
	return property{Table: table, Name: l[0]}
}

func (t propMapping) MarshalJSON() ([]byte, error) {
	m := make(map[string]string)
	for prop, s := range t {
		m[prop.String()] = s
	}
	return json.Marshal(m)
}

func (t propSlice) MarshalJSON() ([]byte, error) {
	l := make([]string, len(t))
	for i, prop := range t {
		l[i] = prop.String()
	}
	return json.Marshal(l)
}

func (t property) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(t.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (t property) String() string {
	if t.Table == "" {
		return t.Name
	}
	return t.Table + "." + t.Name
}

func newTable(name string) *table {
	t := table{name: name}
	t.joins = make(map[string]string)
	return &t
}

func (t table) Name() string {
	return t.name
}

func (t *table) SetEntry(e interface{}) *table {
	t.entry = e
	t.makePropMap(e)
	return t
}

func (t *table) SetJoin(table, sql string) *table {
	t.joins[table] = sql
	return t
}

func (t table) parseProperty(s string) property {
	return parseProperty(s, t.name)
}

func (t table) withJoins(tx *gorm.DB, props propSlice) *gorm.DB {
	joined := make(map[string]interface{})
	for _, prop := range props {
		if _, ok := joined[prop.Table]; ok {
			// already joined
			continue
		}
		if join, ok := t.joins[prop.Table]; ok {
			joined[prop.Table] = nil
			tx = tx.Joins(join)
		}
	}
	return tx
}

func (t table) MakeResponse(r *http.Request, tx *gorm.DB, data interface{}) (*tableResponse, error) {
	var total int64

	// props selection
	props := t.queryProps(r)
	tx = t.withJoins(tx, props)

	// paging
	offset, limit := page(r)
	tx = tx.Offset(offset).Limit(limit)

	// grouping
	for _, prop := range t.queryGroups(r) {
		tx = tx.Group(prop.String())
	}

	result := tx.Count(&total).Find(data)
	if result.Error != nil {
		return nil, result.Error
	}

	td := &tableResponse{}
	td.Data = t.remap(data, props)
	if queryMeta(r) {
		td.Meta = &tableResponseMeta{
			Total:          total,
			Offset:         offset,
			Limit:          limit,
			Count:          len(td.Data),
			AvailableProps: t.props(),
			IncludedProps:  props,
		}
	}
	return td, nil
}

func (t *table) makePropMap(i interface{}) {
	t.propMap = make(propMapping)
	fieldNames, err := attr.Names(i)
	if err != nil {
		return
	}
	for _, fieldName := range fieldNames {
		if propName, err := attr.GetTag(i, fieldName, "json"); err == nil {
			t.propMap[t.parseProperty(propName)] = fieldName
		}
	}
}

func (t table) props() propSlice {
	props := make(propSlice, len(t.propMap))
	i := 0
	for prop, _ := range t.propMap {
		props[i] = prop
		i = i + 1
	}
	return props
}

func (t table) linePropValue(line interface{}, prop property) (interface{}, bool) {
	if fieldName, ok := t.propMap[prop]; ok {
		i, _ := attr.GetValue(line, fieldName)
		return i, true
	}
	return nil, false
}

func (t table) lineMap(line interface{}, props propSlice) map[string]interface{} {
	if len(props) == 0 {
		props = t.props()
	}
	rm := make(map[string]interface{})
	for _, p := range props {
		if v, ok := t.linePropValue(line, p); ok {
			rm[p.String()] = v
		}
	}
	return rm
}

func (t table) remap(data interface{}, props propSlice) []map[string]interface{} {
	switch reflect.TypeOf(data).Kind() {
	case reflect.Ptr:
		s := reflect.ValueOf(data).Elem()
		l := make([]map[string]interface{}, s.Len())
		for i := 0; i < s.Len(); i++ {
			l[i] = t.lineMap(s.Index(i).Interface(), props)
		}
		return l
	case reflect.Slice:
		s := reflect.ValueOf(data)
		l := make([]map[string]interface{}, s.Len())
		for i := 0; i < s.Len(); i++ {
			l[i] = t.lineMap(s.Index(i).Interface(), props)
		}
		return l
	default:
		return make([]map[string]interface{}, 0)
	}
}

func queryMeta(r *http.Request) bool {
	switch r.URL.Query().Get("meta") {
	case "f", "F", "false", "0":
		return false
	default:
		return true
	}
}

func (t table) queryProps(r *http.Request) propSlice {
	s := r.URL.Query().Get("props")
	l := strings.Split(s, ",")
	props := make(propSlice, 0)
	for _, s := range l {
		if s == "" {
			continue
		}
		props = append(props, t.parseProperty(s))
	}
	return props
}

func (t table) queryGroups(r *http.Request) propSlice {
	s := r.URL.Query().Get("groupby")
	l := strings.Split(s, ",")
	props := make(propSlice, 0)
	for _, s := range l {
		if s == "" {
			continue
		}
		props = append(props, t.parseProperty(s))
	}
	return props
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		Remap string
		Desc  bool
	}
	tableResponseData []map[string]interface{}
	tableResponseMeta struct {
		Total          int64     `json:"total"`
		Offset         int       `json:"offset"`
		Limit          int       `json:"limit"`
		Count          int       `json:"count"`
		AvailableProps propSlice `json:"available_props"`
		IncludedProps  propSlice `json:"included_props"`
	}
	propSlice   []property
	propMapping map[property]string
)

func parseProperty(s string, table string) property {
	prop := property{}

	// negation prefix "~"
	if strings.HasPrefix(s, "~") {
		prop.Desc = true
		s = strings.TrimLeft(s, "~")
	}
	if s == "" {
		return prop
	}

	// remapping
	l := strings.SplitN(s, ":", 2)
	if len(l) == 2 {
		s = l[0]
		prop.Remap = l[1]
	}

	// table.name split
	l = strings.SplitN(s, ".", 2)
	if len(l) == 2 {
		prop.Table = l[0]
		prop.Name = l[1]
	} else {
		prop.Table = table
		prop.Name = l[0]
	}
	return prop
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

func (t property) SQLWithOrder() string {
	s := t.SQL()
	if t.Desc {
		s += " DESC"
	}
	return s
}

func (t property) SQL() string {
	if t.Table == "" {
		return fmt.Sprintf("`%s`", t.Name)
	}
	return fmt.Sprintf("`%s`.`%s`", t.Table, t.Name)
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

func (t table) Entry() interface{} {
	return t.entry
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

func (t table) DBMigrate() error {
	return t.DBTable().AutoMigrate(t.entry)
}

func (t table) DBTable() *gorm.DB {
	return db.Table(t.name)
}

func (t table) withJoins(tx *gorm.DB, props propSlice) *gorm.DB {
	joined := make(map[string]interface{})
	selects := make([]string, len(props))
	for i, prop := range props {
		var as string
		if prop.Remap != "" {
			as = prop.Remap
		} else {
			as = prop.String()
		}
		selects[i] = fmt.Sprintf("%s as `%s`", prop.SQL(), as)

		if _, ok := joined[prop.Table]; ok {
			// already joined
			continue
		}
		if join, ok := t.joins[prop.Table]; ok {
			joined[prop.Table] = nil
			tx = tx.Joins(join)
		}
	}
	return tx.Select(strings.Join(selects, ","))
}

func (t table) MakeResponse(r *http.Request, tx *gorm.DB, data2 interface{}) (*tableResponse, error) {
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

	// ordering
	qOrders := t.queryOrders(r)
	if len(qOrders) > 0 {
		sqlOrders := make([]string, len(qOrders))
		for i, prop := range qOrders {
			sqlOrders[i] = prop.SQLWithOrder()
		}
		tx = tx.Order(strings.Join(sqlOrders, ","))
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, err
	}

	data := make([]map[string]interface{}, 0)
	if err := tx.Find(&data).Error; err != nil {
		return nil, err
	}

	td := &tableResponse{}
	td.Data = data
	//td.Data = t.remap(data, props)
	if queryMeta(r) {
		td.Meta = &tableResponseMeta{
			Total:          total,
			Offset:         offset,
			Limit:          limit,
			Count:          len(td.Data),
			AvailableProps: availProps(props),
			IncludedProps:  props,
		}
	}
	return td, nil
}

func availProps(props []property) []property {
	done := make(map[string]interface{})
	ap := make([]property, 0)
	for _, prop := range props {
		if _, ok := done[prop.Table]; ok {
			// table props already added
			continue
		}
		if t, ok := tables[prop.Table]; !ok {
			// unknown table
			continue
		} else {
			ap = append(ap, t.props()...)
		}
	}
	return ap
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
	return t.parsePropSlice(s)
}

func (t table) queryGroups(r *http.Request) propSlice {
	s := r.URL.Query().Get("groupby")
	return t.parsePropSlice(s)
}

func (t table) queryOrders(r *http.Request) propSlice {
	s := r.URL.Query().Get("orderby")
	return t.parsePropSlice(s)
}

func (t table) parsePropSlice(s string) propSlice {
	props := make(propSlice, 0)
	for _, s := range strings.Split(s, ",") {
		if s == "" {
			continue
		}
		props = append(props, t.parseProperty(s))
	}
	return props
}

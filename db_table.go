package main

import (
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
		propMap map[string]string
	}
	tableResponse struct {
		Data tableResponseData  `json:"data"`
		Meta *tableResponseMeta `json:"meta,omitempty"`
	}
	tableResponseData []map[string]interface{}
	tableResponseMeta struct {
		Total          int64             `json:"total"`
		Offset         int               `json:"offset"`
		Limit          int               `json:"limit"`
		Count          int               `json:"count"`
		AvailableProps []string          `json:"available_props"`
		IncludedProps  []string          `json:"included_props"`
		Mapping        map[string]string `json:"mapping"`
	}
)

func newTable() *table {
	return &table{}
}
func (t table) Name() string {
	return t.name
}
func (t *table) SetName(name string) *table {
	t.name = name
	return t
}
func (t *table) SetEntry(e interface{}) *table {
	t.entry = e
	t.propMap = makePropMap(e)
	return t
}

func (t table) MakeResponse(r *http.Request, tx *gorm.DB, data interface{}) (*tableResponse, error) {
	var total int64
	props := queryProps(r)
	offset, limit := page(r)
	result := tx.Offset(offset).Limit(limit).Count(&total).Find(data)
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

func makePropMap(i interface{}) map[string]string {
	m := make(map[string]string)
	names, err := attr.Names(i)
	if err != nil {
		return m
	}
	for _, name := range names {
		if prop, err := attr.GetTag(i, name, "json"); err == nil {
			m[prop] = name
		}
	}
	return m
}

func (t table) props() []string {
	props := make([]string, len(t.propMap))
	i := 0
	for k, _ := range t.propMap {
		props[i] = k
		i = i + 1
	}
	return props
}

func (t table) linePropValue(line interface{}, s string) (interface{}, bool) {
	if fieldName, ok := t.propMap[s]; ok {
		i, _ := attr.GetValue(line, fieldName)
		return i, true
	}
	return nil, false
}

func (t table) lineMap(line interface{}, props []string) map[string]interface{} {
	if len(props) == 0 {
		props = t.props()
	}
	rm := make(map[string]interface{})
	for _, p := range props {
		if v, ok := t.linePropValue(line, p); ok {
			rm[p] = v
		}
	}
	return rm
}

func (t table) remap(data interface{}, props []string) []map[string]interface{} {
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

func queryProps(r *http.Request) []string {
	s := r.URL.Query().Get("props")
	return strings.Split(s, ",")
}

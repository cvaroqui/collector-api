package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/shaj13/go-guardian/v2/auth"
	"github.com/ssrathi/go-attr"
	"gorm.io/gorm"
)

type (
	tableRoute struct {
		From, To string
		Via      []string
	}
	tableJoin struct {
		From, To string
		Cols     [][]string
	}
	table struct {
		name    string
		entry   interface{}
		propMap propMapping
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
	propSlice    []property
	propMapping  map[property]string
	joinedTables map[string]interface{}
	request      struct {
		tx     *gorm.DB
		table  table
		joined joinedTables
		on     joinedTables
	}
)

var (
	reFilter   = regexp.MustCompile(`([a-zA-Z_.]+)\s*(=|>| |~|>=|<=)\s*(.*)`)
	tableJoins = []tableJoin{
		{From: "tags", To: "node_tags", Cols: [][]string{{"tag_id", "tag_id"}}},
		{From: "tags", To: "svc_tags", Cols: [][]string{{"tag_id", "tag_id"}}},
		{From: "nodes", To: "node_tags", Cols: [][]string{{"node_id", "node_id"}}},
		{From: "nodes", To: "svcmon", Cols: [][]string{{"node_id", "node_id"}}},
		{From: "nodes", To: "apps", Cols: [][]string{{"app", "app"}}},
		{From: "services", To: "apps", Cols: [][]string{{"svc_app", "app"}}},
		{From: "services", To: "svcmon", Cols: [][]string{{"svc_id", "svc_id"}}},
		{From: "services", To: "svc_tags", Cols: [][]string{{"svc_id", "svc_id"}}},
		{From: "svcmon", To: "svc_tags", Cols: [][]string{{"svc_id", "svc_id"}}},
		{From: "svcmon", To: "resmon", Cols: [][]string{{"node_id", "node_id"}, {"svc_id", "svc_id"}}},
		{From: "auth_user", To: "auth_membership", Cols: [][]string{{"user_id", "user_id"}}},
		{From: "auth_membership", To: "apps_publications", Cols: [][]string{{"group_id", "group_id"}}},
		{From: "auth_membership", To: "apps_responsibles", Cols: [][]string{{"group_id", "group_id"}}},
		{From: "apps", To: "apps_publications", Cols: [][]string{{"id", "app_id"}}},
		{From: "apps", To: "apps_responsibles", Cols: [][]string{{"id", "app_id"}}},
	}
	tableRoutes = []tableRoute{
		{From: "tags", To: "nodes", Via: []string{"node_tags"}},
		{From: "tags", To: "services", Via: []string{"svc_tags"}},
		{From: "nodes", To: "apps_publications", Via: []string{"apps"}},
		{From: "nodes", To: "auth_membership", Via: []string{"apps", "apps_publications"}},
		{From: "node_tags", To: "apps", Via: []string{"nodes", "apps", "apps_publications"}},
		{From: "node_tags", To: "apps_publications", Via: []string{"nodes", "apps"}},
		{From: "services", To: "nodes", Via: []string{"svcmon"}},
		{From: "services", To: "apps_publications", Via: []string{"apps"}},
		{From: "services", To: "apps_responsibles", Via: []string{"apps"}},
		{From: "services", To: "auth_membership", Via: []string{"apps", "apps_publications"}},
		{From: "svc_tags", To: "apps", Via: []string{"services", "apps", "apps_publications"}},
		{From: "svc_tags", To: "apps_publications", Via: []string{"services"}},
	}
)

func (t table) Request() *request {
	req := request{
		table:  t,
		tx:     t.Table(),
		joined: make(joinedTables),
		on:     make(joinedTables),
	}
	return &req
}

func (t *request) Where(query interface{}, args ...interface{}) {
	t.tx = t.tx.Where(query, args...)
}

func (t *request) Not(query interface{}, args ...interface{}) {
	t.tx = t.tx.Not(query, args...)
}

func findJoin(from, to string) (tableJoin, bool) {
	for _, j := range tableJoins {
		if from == j.From && to == j.To {
			return j, true
		}
		if to == j.From && from == j.To {
			return j, true
		}
	}
	return tableJoin{}, false
}

func (t joinedTables) Add(j tableJoin) {
	s := j.Key()
	t[s] = nil
}
func (t joinedTables) Has(j tableJoin) bool {
	s := j.Key()
	_, ok := t[s]
	return ok
}

func genKey(a, b string) string {
	if a < b {
		return fmt.Sprintf("%s:%s", a, b)
	} else {
		return fmt.Sprintf("%s:%s", b, a)
	}
}

func (t tableJoin) Key() string {
	return genKey(t.From, t.To)
}
func (t tableJoin) String() string {
	return t.SQL(t.To)
}
func (t tableJoin) SQL(s string) string {
	cols := make([]string, len(t.Cols))
	for i, col := range t.Cols {
		cols[i] = fmt.Sprintf("`%s`.`%s`=`%s`.`%s`", t.From, col[0], t.To, col[1])
	}
	return fmt.Sprintf("LEFT JOIN `%s` ON %s", s, strings.Join(cols, " AND "))
}

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

func (t table) parseProperty(s string) property {
	return parseProperty(s, t.name)
}

func (t table) AutoMigrate() error {
	return t.Table().AutoMigrate(t.entry)
}

func (t table) Table() *gorm.DB {
	return db.Table(t.name)
}

func (t *request) withFilters(filters []string) {
	for _, s := range filters {
		l := reFilter.FindStringSubmatch(s)
		if l == nil {
			continue
		}
		if len(l) != 4 {
			continue
		}
		var (
			op, value string
			neg       bool
		)
		switch l[2] {
		case "~", " ":
			op = "LIKE"
		default:
			op = l[2]
		}
		value = l[3]
		if strings.HasPrefix(value, "!") {
			neg = true
			value = strings.TrimLeft(value, "!")
		}
		if value == "empty" {
			where := fmt.Sprintf("%s IS NULL or %s = ?", l[1], l[1])
			if neg {
				t.tx = t.tx.Not(where, "")
			} else {
				t.tx = t.tx.Where(where, "")
			}
		} else {
			if strings.HasPrefix(value, "(") && op == " " {
				op = "IN"
			}
			where := l[1] + " " + op + " ?"
			if neg {
				t.tx = t.tx.Not(where, value)
			} else {
				t.tx = t.tx.Where(where, value)
			}
		}
	}
}

func (t *request) withJoins(props propSlice) {
	if len(props) == 0 {
		props = t.table.props()
	}
	selects := make([]string, len(props))
	for i, prop := range props {
		var as string
		if prop.Remap != "" {
			as = prop.Remap
		} else {
			as = prop.String()
		}
		t.AutoJoin(prop.Table)
		selects[i] = fmt.Sprintf("%s as `%s`", prop.SQL(), as)
	}
	t.tx.Select(strings.Join(selects, ","))
}

func (t *request) withACL(user auth.Info) {
	if _, err := strconv.Atoi(user.GetID()); err != nil {
		// node
		t.AutoJoin("apps")
		t.AutoJoin("nodes")
		t.Where("apps.app = nodes.app")
		t.Where("nodes.node_id = ?", user.GetID())
	} else {
		t.AutoJoin("apps_publications")
		t.AutoJoin("auth_membership")
		t.Where("apps_publications.group_id = auth_membership.group_id")
		t.Where("auth_membership.user_id = ?", user.GetID())
		// user
	}
}

// getHops returns
//  []string{"svc_tags", "tags"} as the "svc to tags" route
//  []string{"tags"} as the "svc_tags to tags" route
func getHops(from, to string) []string {
	for _, r := range tableRoutes {
		routeKey := genKey(r.From, r.To)
		if routeKey == genKey(from, to) {
			return append(r.Via, to)
		}
		if routeKey == genKey(to, from) {
			// return the reversed hops
			s := append([]string{to}, r.Via...)
			for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
				s[i], s[j] = s[j], s[i]
			}
			return s
		}
	}
	return []string{to}
}

func (t *request) AutoJoin(table string) {
	if table == t.table.name {
		// self join, noop
		return
	}
	here := t.table.name
	for _, there := range getHops(t.table.name, table) {
		j, ok := findJoin(here, there)
		if !ok {
			log.Printf("missing autojoin from %s to %s", here, there)
			return
		}
		if t.joined.Has(j) {
			// already joined
		} else {
			sql := j.SQL(there)
			t.joined.Add(j)
			t.tx = t.tx.Joins(sql)
		}
		here = there
	}
	return
}

func (t *request) withGroups(groups propSlice) {
	for _, prop := range groups {
		t.tx = t.tx.Group(prop.String())
	}
}

func (t *request) withOrders(orders propSlice) {
	if len(orders) == 0 {
		return
	}

	sqlOrders := make([]string, len(orders))
	for i, prop := range orders {
		sqlOrders[i] = prop.SQLWithOrder()
	}
	t.tx = t.tx.Order(strings.Join(sqlOrders, ","))
}

func (t *request) MakeResponse(r *http.Request) (*tableResponse, error) {
	var total int64

	user := auth.User(r)
	t.withACL(user)

	// props selection
	props := t.table.queryProps(r)
	t.withJoins(props)

	// filters
	filters := t.table.queryFilters(r)
	t.withFilters(filters)

	// grouping
	groups := t.table.queryGroups(r)
	t.withGroups(groups)

	// ordering
	orders := t.table.queryOrders(r)
	t.withOrders(orders)

	if err := t.tx.Count(&total).Error; err != nil {
		return nil, err
	}

	// paging
	offset, limit := page(r)
	t.tx = t.tx.Offset(offset).Limit(limit)

	data := make([]map[string]interface{}, 0)
	if err := t.tx.Find(&data).Error; err != nil {
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

func availProps(props propSlice) propSlice {
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
	return propSlice(ap)
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

func (t table) queryFilters(r *http.Request) []string {
	if l, ok := r.URL.Query()["filters"]; ok {
		return l
	}
	return []string{}
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

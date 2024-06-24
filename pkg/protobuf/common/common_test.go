package common

import (
	"testing"
)

func TestSortSql(t *testing.T) {
	datas := []struct {
		sort string
		sql  string
	}{
		{sort: "created_at", sql: "ORDER BY created_at DESC"},
		{sort: "+created_at", sql: "ORDER BY created_at ASC"},
		{sort: "-created_at", sql: "ORDER BY created_at DESC"},
		{sort: "-created_at,updated_at", sql: "ORDER BY created_at DESC,updated_at DESC"},
		{sort: "-created_at,+updated_at", sql: "ORDER BY created_at DESC,updated_at ASC"},
		{sort: "+created_at,+updated_at", sql: "ORDER BY created_at ASC,updated_at ASC"},
		{sort: "+created_at ,+updated_at", sql: "ORDER BY created_at ASC,updated_at ASC"},
		{sort: "+created_at , +updated_at", sql: "ORDER BY created_at ASC,updated_at ASC"},
	}
	for _, data := range datas {
		req := &ReqWithPage{Sort: data.sort}
		if sql := req.orderSql(); sql != data.sql {
			t.Errorf("sort %s get sql fail, want: %s, get: %s", data.sort, data.sql, sql)
			t.FailNow()
		}
	}
}

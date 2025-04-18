package requestpb

import (
	"strings"

	"gorm.io/gorm"
)

const (
	sortDelimiter = ","
	// 降序
	sortDesc    = "-"
	sortDescStr = "DESC"
	// 升序
	sortAsc    = "+"
	sortAscStr = "ASC"
	// 默认排序
	defaultSortStr = sortDescStr
)

// IsValid 请求参数是否有效
// 校验page，size，sort参数
// 如果page >0 则-1
// 如果size为0，则设置为默认size
// func (r *ReqWithPage) IsValid(defaultSize int32, supportSortFields []string) error {
// 	if defaultSize == 0 {
// 		return status.Error(codes.InvalidArgument, "defaultSize not setted")
// 	}
// 	if r.Page > 0 {
// 		r.Page -= 1
// 	}
// 	if r.Size == 0 {
// 		r.Size = defaultSize
// 	}
// 	if r.Sort != "" {
// 		for _, sortField := range strings.Split(r.Sort, sortDelimiter) {
// 			supported := false
// 			sf := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(sortField), sortDesc), sortAsc)
// 			for _, ssf := range supportSortFields {
// 				if ssf == sf {
// 					supported = true
// 					break
// 				}
// 			}
// 			if !supported {
// 				return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid sort field %s", sortField))
// 			}
// 		}
// 	} else if len(supportSortFields) != 0 {
// 		r.Sort = supportSortFields[0]
// 	}
// 	return nil
// }

// func (r *ReqWithId) IsValid(fullMethod string) error {
// 	if r.Id == 0 {
// 		return status.Error(codes.InvalidArgument, "id is must")
// 	}
// 	return nil
// }

// db.Scopes(in.GormScope())
func (r *ReqWithPage) GormScope(defaultSort string) func(*gorm.DB) *gorm.DB {
	return ReqWithPageGormScope(r.Page, r.Size, r.Sort, defaultSort)
}

// ReqWithPageGormScope gorm分页查询
// size 小于0不分页
func ReqWithPageGormScope(page, size int32, sort, defaultSort string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if size > 0 {
			if page > 0 {
				page -= 1
			}
			db.Offset(int(page * size)).
				Limit(int(size))
		}
		if sort != "" {
			db.Order(gormOrderStr(sort))
		} else if defaultSort != "" {
			db.Order(gormOrderStr(defaultSort))
		}
		return db
	}
}

func gormOrderStr(sort string) string {
	sql := ""
	for index, sortField := range strings.Split(sort, sortDelimiter) {
		if index != 0 {
			sql += ","
		}
		sf := strings.TrimSpace(sortField)
		sfName := strings.TrimPrefix(strings.TrimPrefix(sf, sortDesc), sortAsc)
		if strings.HasPrefix(sf, sortDesc) {
			sql += sfName + " " + sortDescStr
		} else if strings.HasPrefix(sf, sortAsc) {
			sql += sfName + " " + sortAscStr
		} else {
			sql += sfName + " " + defaultSortStr
		}

	}
	return sql
}

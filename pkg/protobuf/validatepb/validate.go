package validatepb

import (
	"fmt"
	"strings"

	"github.com/asjard/asjard/core/status"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
)

const (
	sortDelimiter = ","
	// 降序
	sortDesc = "-"
	// 升序
	sortAsc = "+"
)

// Validater 参数校验需要实现的方法
type Validater interface {
	// 是否为有效的参数
	IsValid(parentFieldName, fullMethod string) error
}

var DefaultValidator = validator.New()

func init() {
	if err := DefaultValidator.RegisterValidation("sortFields", isSortField); err != nil {
		panic(err)
	}
}

func isSortField(fl validator.FieldLevel) bool {
	return isSortValid(fl.Field().String(), strings.Split(fl.Param(), " ")) == nil
}

func isSortValid(sort string, supportSortFields []string) error {
	if sort == "" {
		return nil
	}
	for _, sortField := range strings.Split(sort, sortDelimiter) {
		supported := false
		sf := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(sortField), sortDesc), sortAsc)
		for _, ssf := range supportSortFields {
			if ssf == sf {
				supported = true
				break
			}
		}
		if !supported {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid sort field %s", sortField))
		}
	}
	return nil
}

func ValidateFieldName(parentFieldName, fieldName string) string {
	if parentFieldName == "" {
		return fieldName
	}
	return parentFieldName + "." + fieldName
}

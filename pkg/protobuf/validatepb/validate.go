package validatepb

import (
	"fmt"
	reflect "reflect"
	"strconv"
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
	if err := DefaultValidator.RegisterValidation("country_code", isCountryCode); err != nil {
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

var (
	countryCodeValidtors = map[string]validator.Func{
		"iso3166_1_alpha2":           isIso3166Alpha2,
		"iso3166_1_alpha2_eu":        isIso3166Alpha2EU,
		"iso3166_1_alpha3":           isIso3166Alpha3,
		"iso3166_1_alpha3_eu":        isIso3166Alpha3EU,
		"iso3166_1_alpha_numeric":    isIso3166AlphaNumeric,
		"iso3166_1_alpha_numeric_eu": isIso3166AlphaNumericEU,
	}
)

func isCountryCode(fl validator.FieldLevel) bool {
	fn, ok := countryCodeValidtors[fl.Param()]
	if !ok {
		return false
	}
	return fn(fl)
}

// isIso3166Alpha2 is the validation function for validating if the current field's value is a valid iso3166-1 alpha-2 country code.
func isIso3166Alpha2(fl validator.FieldLevel) bool {
	_, ok := iso3166_1_alpha2[fl.Field().String()]
	return ok
}

// isIso3166Alpha2EU is the validation function for validating if the current field's value is a valid iso3166-1 alpha-2 European Union country code.
func isIso3166Alpha2EU(fl validator.FieldLevel) bool {
	_, ok := iso3166_1_alpha2_eu[fl.Field().String()]
	return ok
}

// isIso3166Alpha3 is the validation function for validating if the current field's value is a valid iso3166-1 alpha-3 country code.
func isIso3166Alpha3(fl validator.FieldLevel) bool {
	_, ok := iso3166_1_alpha3[fl.Field().String()]
	return ok
}

// isIso3166Alpha3EU is the validation function for validating if the current field's value is a valid iso3166-1 alpha-3 European Union country code.
func isIso3166Alpha3EU(fl validator.FieldLevel) bool {
	_, ok := iso3166_1_alpha3_eu[fl.Field().String()]
	return ok
}

// isIso3166AlphaNumeric is the validation function for validating if the current field's value is a valid iso3166-1 alpha-numeric country code.
func isIso3166AlphaNumeric(fl validator.FieldLevel) bool {
	field := fl.Field()

	var code int
	switch field.Kind() {
	case reflect.String:
		i, err := strconv.Atoi(field.String())
		if err != nil {
			return false
		}
		code = i % 1000
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		code = int(field.Int() % 1000)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		code = int(field.Uint() % 1000)
	default:
		panic(fmt.Sprintf("Bad field type %s", field.Type()))
	}

	_, ok := iso3166_1_alpha_numeric[code]
	return ok
}

// isIso3166AlphaNumericEU is the validation function for validating if the current field's value is a valid iso3166-1 alpha-numeric European Union country code.
func isIso3166AlphaNumericEU(fl validator.FieldLevel) bool {
	field := fl.Field()

	var code int
	switch field.Kind() {
	case reflect.String:
		i, err := strconv.Atoi(field.String())
		if err != nil {
			return false
		}
		code = i % 1000
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		code = int(field.Int() % 1000)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		code = int(field.Uint() % 1000)
	default:
		panic(fmt.Sprintf("Bad field type %s", field.Type()))
	}

	_, ok := iso3166_1_alpha_numeric_eu[code]
	return ok
}

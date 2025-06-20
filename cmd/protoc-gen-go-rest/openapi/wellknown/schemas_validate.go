package wellknown

import (
	"strconv"
	"strings"
	"sync"

	v3 "github.com/google/gnostic/openapiv3"
)

// Func accepts a FieldLevel interface for all validation needs. The return
// value should be true when validation succeeds.
type Func func(sm *v3.Schema, rv string)

var (
	// bakedInAliases is a default mapping of a single validation tag that
	// defines a common or complex set of validation(s) to simplify
	// adding validation to structs.
	bakedInAliases = map[string][]string{
		"iscolor":         []string{"hexcolor", "rgb", "rgba", "hsl", "hsla"},
		"country_code":    []string{"iso3166_1_alpha2", "iso3166_1_alpha3", "iso3166_1_alpha_numeric"},
		"eu_country_code": []string{"iso3166_1_alpha2_eu", "iso3166_1_alpha3_eu", "iso3166_1_alpha_numeric_eu"},
	}

	// bakedInValidators is the default map of ValidationFunc
	// you can add, remove or even replace items to suite your needs,
	// or even disregard and use your own map if so desired.
	bakedInValidators = map[string]Func{
		"color":                         iscolor,
		"country_code":                  isCountryCode,
		"eu_country_code":               euCountryCode,
		"required":                      hasValue,
		"required_if":                   requiredIf,
		"required_unless":               requiredUnless,
		"skip_unless":                   skipUnless,
		"required_with":                 requiredWith,
		"required_with_all":             requiredWithAll,
		"required_without":              requiredWithout,
		"required_without_all":          requiredWithoutAll,
		"excluded_if":                   excludedIf,
		"excluded_unless":               excludedUnless,
		"excluded_with":                 excludedWith,
		"excluded_with_all":             excludedWithAll,
		"excluded_without":              excludedWithout,
		"excluded_without_all":          excludedWithoutAll,
		"isdefault":                     isDefault,
		"len":                           hasLengthOf,
		"min":                           hasMinOf,
		"max":                           hasMaxOf,
		"eq":                            isEq,
		"eq_ignore_case":                isEqIgnoreCase,
		"ne":                            isNe,
		"ne_ignore_case":                isNeIgnoreCase,
		"lt":                            isLt,
		"lte":                           isLte,
		"gt":                            isGt,
		"gte":                           isGte,
		"eqfield":                       isEqField,
		"eqcsfield":                     isEqCrossStructField,
		"necsfield":                     isNeCrossStructField,
		"gtcsfield":                     isGtCrossStructField,
		"gtecsfield":                    isGteCrossStructField,
		"ltcsfield":                     isLtCrossStructField,
		"ltecsfield":                    isLteCrossStructField,
		"nefield":                       isNeField,
		"gtefield":                      isGteField,
		"gtfield":                       isGtField,
		"ltefield":                      isLteField,
		"ltfield":                       isLtField,
		"fieldcontains":                 fieldContains,
		"fieldexcludes":                 fieldExcludes,
		"alpha":                         isAlpha,
		"alphanum":                      isAlphanum,
		"alphaunicode":                  isAlphaUnicode,
		"alphanumunicode":               isAlphanumUnicode,
		"boolean":                       isBoolean,
		"numeric":                       isNumeric,
		"number":                        isNumber,
		"hexadecimal":                   isHexadecimal,
		"hexcolor":                      isHEXColor,
		"rgb":                           isRGB,
		"rgba":                          isRGBA,
		"hsl":                           isHSL,
		"hsla":                          isHSLA,
		"e164":                          isE164,
		"email":                         isEmail,
		"url":                           isURL,
		"http_url":                      isHttpURL,
		"uri":                           isURI,
		"urn_rfc2141":                   isUrnRFC2141, // RFC 2141
		"file":                          isFile,
		"filepath":                      isFilePath,
		"base32":                        isBase32,
		"base64":                        isBase64,
		"base64url":                     isBase64URL,
		"base64rawurl":                  isBase64RawURL,
		"contains":                      contains,
		"containsany":                   containsAny,
		"containsrune":                  containsRune,
		"excludes":                      excludes,
		"excludesall":                   excludesAll,
		"excludesrune":                  excludesRune,
		"startswith":                    startsWith,
		"endswith":                      endsWith,
		"startsnotwith":                 startsNotWith,
		"endsnotwith":                   endsNotWith,
		"image":                         isImage,
		"isbn":                          isISBN,
		"isbn10":                        isISBN10,
		"isbn13":                        isISBN13,
		"issn":                          isISSN,
		"eth_addr":                      isEthereumAddress,
		"eth_addr_checksum":             isEthereumAddressChecksum,
		"btc_addr":                      isBitcoinAddress,
		"btc_addr_bech32":               isBitcoinBech32Address,
		"uuid":                          isUUID,
		"uuid3":                         isUUID3,
		"uuid4":                         isUUID4,
		"uuid5":                         isUUID5,
		"uuid_rfc4122":                  isUUIDRFC4122,
		"uuid3_rfc4122":                 isUUID3RFC4122,
		"uuid4_rfc4122":                 isUUID4RFC4122,
		"uuid5_rfc4122":                 isUUID5RFC4122,
		"ulid":                          isULID,
		"md4":                           isMD4,
		"md5":                           isMD5,
		"sha256":                        isSHA256,
		"sha384":                        isSHA384,
		"sha512":                        isSHA512,
		"ripemd128":                     isRIPEMD128,
		"ripemd160":                     isRIPEMD160,
		"tiger128":                      isTIGER128,
		"tiger160":                      isTIGER160,
		"tiger192":                      isTIGER192,
		"ascii":                         isASCII,
		"printascii":                    isPrintableASCII,
		"multibyte":                     hasMultiByteCharacter,
		"datauri":                       isDataURI,
		"latitude":                      isLatitude,
		"longitude":                     isLongitude,
		"ssn":                           isSSN,
		"ipv4":                          isIPv4,
		"ipv6":                          isIPv6,
		"ip":                            isIP,
		"cidrv4":                        isCIDRv4,
		"cidrv6":                        isCIDRv6,
		"cidr":                          isCIDR,
		"tcp4_addr":                     isTCP4AddrResolvable,
		"tcp6_addr":                     isTCP6AddrResolvable,
		"tcp_addr":                      isTCPAddrResolvable,
		"udp4_addr":                     isUDP4AddrResolvable,
		"udp6_addr":                     isUDP6AddrResolvable,
		"udp_addr":                      isUDPAddrResolvable,
		"ip4_addr":                      isIP4AddrResolvable,
		"ip6_addr":                      isIP6AddrResolvable,
		"ip_addr":                       isIPAddrResolvable,
		"unix_addr":                     isUnixAddrResolvable,
		"mac":                           isMAC,
		"hostname":                      isHostnameRFC952,  // RFC 952
		"hostname_rfc1123":              isHostnameRFC1123, // RFC 1123
		"fqdn":                          isFQDN,
		"unique":                        isUnique,
		"oneof":                         isOneOf,
		"oneofci":                       isOneOfCI,
		"html":                          isHTML,
		"html_encoded":                  isHTMLEncoded,
		"url_encoded":                   isURLEncoded,
		"dir":                           isDir,
		"dirpath":                       isDirPath,
		"json":                          isJSON,
		"jwt":                           isJWT,
		"hostname_port":                 isHostnamePort,
		"port":                          isPort,
		"lowercase":                     isLowercase,
		"uppercase":                     isUppercase,
		"datetime":                      isDatetime,
		"timezone":                      isTimeZone,
		"iso3166_1_alpha2":              isIso3166Alpha2,
		"iso3166_1_alpha2_eu":           isIso3166Alpha2EU,
		"iso3166_1_alpha3":              isIso3166Alpha3,
		"iso3166_1_alpha3_eu":           isIso3166Alpha3EU,
		"iso3166_1_alpha_numeric":       isIso3166AlphaNumeric,
		"iso3166_1_alpha_numeric_eu":    isIso3166AlphaNumericEU,
		"iso3166_2":                     isIso31662,
		"iso4217":                       isIso4217,
		"iso4217_numeric":               isIso4217Numeric,
		"bcp47_language_tag":            isBCP47LanguageTag,
		"postcode_iso3166_alpha2":       isPostcodeByIso3166Alpha2,
		"postcode_iso3166_alpha2_field": isPostcodeByIso3166Alpha2Field,
		"bic":                           isIsoBicFormat,
		"semver":                        isSemverFormat,
		"dns_rfc1035_label":             isDnsRFC1035LabelFormat,
		"credit_card":                   isCreditCard,
		"cve":                           isCveFormat,
		"luhn_checksum":                 hasLuhnChecksum,
		"mongodb":                       isMongoDBObjectId,
		"mongodb_connection_string":     isMongoDBConnectionString,
		"cron":                          isCron,
		"spicedb":                       isSpiceDB,
		"ein":                           isEIN,
		"validateFn":                    isValidateFn,
	}
)

var (
	oneofValsCache       = map[string][]string{}
	oneofValsCacheRWLock = sync.RWMutex{}
)

func parseOneOfParam2(s string) []string {
	oneofValsCacheRWLock.RLock()
	vals, ok := oneofValsCache[s]
	oneofValsCacheRWLock.RUnlock()
	if !ok {
		oneofValsCacheRWLock.Lock()
		vals = splitParamsRegex().FindAllString(s, -1)
		for i := 0; i < len(vals); i++ {
			vals[i] = strings.ReplaceAll(vals[i], "'", "")
		}
		oneofValsCache[s] = vals
		oneofValsCacheRWLock.Unlock()
	}
	return vals
}

func schemaWithValidate(sm *v3.Schema, rules map[string]string) *v3.Schema {
	for rk, rv := range rules {
		schemaWithRule(sm, rk, rv)
	}
	return sm
}

func schemaWithRule(sm *v3.Schema, rk, rv string) {
	if fn, ok := bakedInValidators[rk]; ok {
		fn(sm, rv)
	}
}

func iscolor(sm *v3.Schema, rv string) {
	if vals, ok := bakedInAliases["iscolor"]; ok {
		sm.Enum = make([]*v3.Any, 0, len(vals))
		for _, v := range vals {
			sm.Enum = append(sm.Enum, &v3.Any{
				Yaml: v,
			})
		}
	}
}

func isCountryCode(sm *v3.Schema, rv string) {
	if vals, ok := bakedInAliases["country_code"]; ok {
		sm.Enum = make([]*v3.Any, 0, len(vals))
		for _, v := range vals {
			sm.Enum = append(sm.Enum, &v3.Any{
				Yaml: v,
			})
		}
	}
}

func euCountryCode(sm *v3.Schema, rv string) {
	if vals, ok := bakedInAliases["eu_country_code"]; ok {
		sm.Enum = make([]*v3.Any, 0, len(vals))
		for _, v := range vals {
			sm.Enum = append(sm.Enum, &v3.Any{
				Yaml: v,
			})
		}
	}
}

func isURLEncoded(sm *v3.Schema, rv string) {
	sm.Pattern = uRLEncodedRegexString
}

func isHTMLEncoded(sm *v3.Schema, rv string) {
	sm.Pattern = hTMLEncodedRegexString
}

func isHTML(sm *v3.Schema, rv string) {
	sm.Pattern = hTMLRegexString
}

func isOneOf(sm *v3.Schema, rv string) {
	vals := parseOneOfParam2(rv)
	sm.Enum = make([]*v3.Any, 0, len(vals))
	for _, v := range vals {
		sm.Enum = append(sm.Enum, &v3.Any{
			Yaml: v,
		})
	}
}

// isOneOfCI is the validation function for validating if the current field's value is one of the provided string values (case insensitive).
func isOneOfCI(sm *v3.Schema, rv string) {
	vals := parseOneOfParam2(rv)
	sm.Enum = make([]*v3.Any, 0, len(vals))
	for _, v := range vals {
		sm.Enum = append(sm.Enum, &v3.Any{
			Yaml: v,
		})
	}
}

// isUnique is the validation function for validating if each array|slice|map value is unique
func isUnique(sm *v3.Schema, rv string) {
	sm.UniqueItems = true
}

// isMAC is the validation function for validating if the field's value is a valid MAC address.
func isMAC(sm *v3.Schema, rv string) {
	sm.Format = "MAC"
}

// isCIDRv4 is the validation function for validating if the field's value is a valid v4 CIDR address.
func isCIDRv4(sm *v3.Schema, rv string) {
	sm.Format = "CIDRv4"
}

// isCIDRv6 is the validation function for validating if the field's value is a valid v6 CIDR address.
func isCIDRv6(sm *v3.Schema, rv string) {
	sm.Format = "CIDRv6"
}

// isCIDR is the validation function for validating if the field's value is a valid v4 or v6 CIDR address.
func isCIDR(sm *v3.Schema, rv string) {
	sm.Format = "CIDR"
}

// isIPv4 is the validation function for validating if a value is a valid v4 IP address.
func isIPv4(sm *v3.Schema, rv string) {
	sm.Format = "IPv4"
}

// isIPv6 is the validation function for validating if the field's value is a valid v6 IP address.
func isIPv6(sm *v3.Schema, rv string) {
	sm.Format = "IPv6"
}

// isIP is the validation function for validating if the field's value is a valid v4 or v6 IP address.
func isIP(sm *v3.Schema, rv string) {
	sm.Format = "IP"
}

// isSSN is the validation function for validating if the field's value is a valid SSN.
func isSSN(sm *v3.Schema, rv string) {
	sm.Pattern = sSNRegexString
}

// isLongitude is the validation function for validating if the field's value is a valid longitude coordinate.
func isLongitude(sm *v3.Schema, rv string) {
	sm.Pattern = longitudeRegexString
}

// isLatitude is the validation function for validating if the field's value is a valid latitude coordinate.
func isLatitude(sm *v3.Schema, rv string) {
	sm.Pattern = latitudeRegexString
}

// isDataURI is the validation function for validating if the field's value is a valid data URI.
func isDataURI(sm *v3.Schema, rv string) {
	sm.Format = "DataURI"
}

// hasMultiByteCharacter is the validation function for validating if the field's value has a multi byte character.
func hasMultiByteCharacter(sm *v3.Schema, rv string) {
	sm.Pattern = multibyteRegexString
}

// isPrintableASCII is the validation function for validating if the field's value is a valid printable ASCII character.
func isPrintableASCII(sm *v3.Schema, rv string) {
	sm.Pattern = printableASCIIRegexString
}

// isASCII is the validation function for validating if the field's value is a valid ASCII character.
func isASCII(sm *v3.Schema, rv string) {
	sm.Pattern = aSCIIRegexString
}

// isUUID5 is the validation function for validating if the field's value is a valid v5 UUID.
func isUUID5(sm *v3.Schema, rv string) {
	sm.Pattern = uUID5RegexString
}

// isUUID4 is the validation function for validating if the field's value is a valid v4 UUID.
func isUUID4(sm *v3.Schema, rv string) {
	sm.Pattern = uUID4RegexString
}

// isUUID3 is the validation function for validating if the field's value is a valid v3 UUID.
func isUUID3(sm *v3.Schema, rv string) {
	sm.Pattern = uUID3RegexString
}

// isUUID is the validation function for validating if the field's value is a valid UUID of any version.
func isUUID(sm *v3.Schema, rv string) {
	sm.Pattern = uUIDRegexString
}

// isUUID5RFC4122 is the validation function for validating if the field's value is a valid RFC4122 v5 UUID.
func isUUID5RFC4122(sm *v3.Schema, rv string) {
	sm.Pattern = uUID5RegexString
}

// isUUID4RFC4122 is the validation function for validating if the field's value is a valid RFC4122 v4 UUID.
func isUUID4RFC4122(sm *v3.Schema, rv string) {
	sm.Pattern = uUID4RFC4122RegexString
}

// isUUID3RFC4122 is the validation function for validating if the field's value is a valid RFC4122 v3 UUID.
func isUUID3RFC4122(sm *v3.Schema, rv string) {
	sm.Pattern = uUID3RFC4122RegexString
}

// isUUIDRFC4122 is the validation function for validating if the field's value is a valid RFC4122 UUID of any version.
func isUUIDRFC4122(sm *v3.Schema, rv string) {
	sm.Pattern = uUIDRFC4122RegexString
}

// isULID is the validation function for validating if the field's value is a valid ULID.
func isULID(sm *v3.Schema, rv string) {
	sm.Pattern = uLIDRegexString
}

// isMD4 is the validation function for validating if the field's value is a valid MD4.
func isMD4(sm *v3.Schema, rv string) {
	sm.Pattern = md4RegexString
}

// isMD5 is the validation function for validating if the field's value is a valid MD5.
func isMD5(sm *v3.Schema, rv string) {
	sm.Pattern = md5RegexString
}

// isSHA256 is the validation function for validating if the field's value is a valid SHA256.
func isSHA256(sm *v3.Schema, rv string) {
	sm.Pattern = sha256RegexString
}

// isSHA384 is the validation function for validating if the field's value is a valid SHA384.
func isSHA384(sm *v3.Schema, rv string) {
	sm.Pattern = sha384RegexString
}

// isSHA512 is the validation function for validating if the field's value is a valid SHA512.
func isSHA512(sm *v3.Schema, rv string) {
	sm.Pattern = sha512RegexString
}

// isRIPEMD128 is the validation function for validating if the field's value is a valid PIPEMD128.
func isRIPEMD128(sm *v3.Schema, rv string) {
	sm.Pattern = ripemd128RegexString
}

// isRIPEMD160 is the validation function for validating if the field's value is a valid PIPEMD160.
func isRIPEMD160(sm *v3.Schema, rv string) {
	sm.Pattern = ripemd160RegexString
}

// isTIGER128 is the validation function for validating if the field's value is a valid TIGER128.
func isTIGER128(sm *v3.Schema, rv string) {
	sm.Pattern = tiger128RegexString
}

// isTIGER160 is the validation function for validating if the field's value is a valid TIGER160.
func isTIGER160(sm *v3.Schema, rv string) {
	sm.Pattern = tiger160RegexString
}

// isTIGER192 is the validation function for validating if the field's value is a valid isTIGER192.
func isTIGER192(sm *v3.Schema, rv string) {
	sm.Pattern = tiger192RegexString
}

// isISBN is the validation function for validating if the field's value is a valid v10 or v13 ISBN.
func isISBN(sm *v3.Schema, rv string) {
	sm.Pattern = iSBN13RegexString
}

// isISBN13 is the validation function for validating if the field's value is a valid v13 ISBN.
func isISBN13(sm *v3.Schema, rv string) {
	sm.Pattern = iSBN13RegexString
}

// isISBN10 is the validation function for validating if the field's value is a valid v10 ISBN.
func isISBN10(sm *v3.Schema, rv string) {
	sm.Pattern = iSBN10RegexString
}

// isISSN is the validation function for validating if the field's value is a valid ISSN.
func isISSN(sm *v3.Schema, rv string) {
	sm.Pattern = iSSNRegexString
}

// isEthereumAddress is the validation function for validating if the field's value is a valid Ethereum address.
func isEthereumAddress(sm *v3.Schema, rv string) {
	sm.Pattern = ethAddressRegexString
}

// isEthereumAddressChecksum is the validation function for validating if the field's value is a valid checksummed Ethereum address.
func isEthereumAddressChecksum(sm *v3.Schema, rv string) {
	sm.Pattern = ethAddressRegexString
}

// isBitcoinAddress is the validation function for validating if the field's value is a valid btc address
func isBitcoinAddress(sm *v3.Schema, rv string) {
	sm.Pattern = btcAddressRegexString
}

// isBitcoinBech32Address is the validation function for validating if the field's value is a valid bech32 btc address
func isBitcoinBech32Address(sm *v3.Schema, rv string) {
	sm.Pattern = btcAddressLowerRegexStringBech32
}

// excludesRune is the validation function for validating that the field's value does not contain the rune specified within the param.
func excludesRune(sm *v3.Schema, rv string) {
}

// excludesAll is the validation function for validating that the field's value does not contain any of the characters specified within the param.
func excludesAll(sm *v3.Schema, rv string) {
}

// excludes is the validation function for validating that the field's value does not contain the text specified within the param.
func excludes(sm *v3.Schema, rv string) {
}

// containsRune is the validation function for validating that the field's value contains the rune specified within the param.
func containsRune(sm *v3.Schema, rv string) {
}

// containsAny is the validation function for validating that the field's value contains any of the characters specified within the param.
func containsAny(sm *v3.Schema, rv string) {
}

// contains is the validation function for validating that the field's value contains the text specified within the param.
func contains(sm *v3.Schema, rv string) {
}

// startsWith is the validation function for validating that the field's value starts with the text specified within the param.
func startsWith(sm *v3.Schema, rv string) {
	sm.Pattern = "^" + rv + ".*"
}

// endsWith is the validation function for validating that the field's value ends with the text specified within the param.
func endsWith(sm *v3.Schema, rv string) {
	sm.Pattern = ".*" + rv + "$"
}

// startsNotWith is the validation function for validating that the field's value does not start with the text specified within the param.
func startsNotWith(sm *v3.Schema, rv string) {
	sm.Pattern = "^(?!" + rv + ").*"
}

// endsNotWith is the validation function for validating that the field's value does not end with the text specified within the param.
func endsNotWith(sm *v3.Schema, rv string) {
	sm.Pattern = ".*(?<!" + rv + ")$"
}

// fieldContains is the validation function for validating if the current field's value contains the field specified by the param's value.
func fieldContains(sm *v3.Schema, rv string) {
	sm.Pattern = ".*" + rv + ".*"
}

// fieldExcludes is the validation function for validating if the current field's value excludes the field specified by the param's value.
func fieldExcludes(sm *v3.Schema, rv string) {
}

// isNeField is the validation function for validating if the current field's value is not equal to the field specified by the param's value.
func isNeField(sm *v3.Schema, rv string) {

}

// isNe is the validation function for validating that the field's value does not equal the provided param value.
func isNe(sm *v3.Schema, rv string) {
}

// isNeIgnoreCase is the validation function for validating that the field's string value does not equal the
// provided param value. The comparison is case-insensitive
func isNeIgnoreCase(sm *v3.Schema, rv string) {
}

// isLteCrossStructField is the validation function for validating if the current field's value is less than or equal to the field, within a separate struct, specified by the param's value.
func isLteCrossStructField(sm *v3.Schema, rv string) {
}

// isLtCrossStructField is the validation function for validating if the current field's value is less than the field, within a separate struct, specified by the param's value.
// NOTE: This is exposed for use within your own custom functions and not intended to be called directly.
func isLtCrossStructField(sm *v3.Schema, rv string) {

}

// isGteCrossStructField is the validation function for validating if the current field's value is greater than or equal to the field, within a separate struct, specified by the param's value.
func isGteCrossStructField(sm *v3.Schema, rv string) {

}

// isGtCrossStructField is the validation function for validating if the current field's value is greater than the field, within a separate struct, specified by the param's value.
func isGtCrossStructField(sm *v3.Schema, rv string) {

}

// isNeCrossStructField is the validation function for validating that the current field's value is not equal to the field, within a separate struct, specified by the param's value.
func isNeCrossStructField(sm *v3.Schema, rv string) {

}

// isEqCrossStructField is the validation function for validating that the current field's value is equal to the field, within a separate struct, specified by the param's value.
func isEqCrossStructField(sm *v3.Schema, rv string) {

}

// isEqField is the validation function for validating if the current field's value is equal to the field specified by the param's value.
func isEqField(sm *v3.Schema, rv string) {

}

// isEq is the validation function for validating if the current field's value is equal to the param's value.
func isEq(sm *v3.Schema, rv string) {

}

// isEqIgnoreCase is the validation function for validating if the current field's string value is
// equal to the param's value.
// The comparison is case-insensitive.
func isEqIgnoreCase(sm *v3.Schema, rv string) {

}

// isPostcodeByIso3166Alpha2 validates by value which is country code in iso 3166 alpha 2
// example: `postcode_iso3166_alpha2=US`
func isPostcodeByIso3166Alpha2(sm *v3.Schema, rv string) {

}

// isPostcodeByIso3166Alpha2Field validates by field which represents for a value of country code in iso 3166 alpha 2
// example: `postcode_iso3166_alpha2_field=CountryCode`
func isPostcodeByIso3166Alpha2Field(sm *v3.Schema, rv string) {

}

// isBase32 is the validation function for validating if the current field's value is a valid base 32.
func isBase32(sm *v3.Schema, rv string) {
	sm.Pattern = base32RegexString
}

// isBase64 is the validation function for validating if the current field's value is a valid base 64.
func isBase64(sm *v3.Schema, rv string) {
	sm.Pattern = base64RegexString
}

// isBase64URL is the validation function for validating if the current field's value is a valid base64 URL safe string.
func isBase64URL(sm *v3.Schema, rv string) {
	sm.Pattern = base64URLRegexString
}

// isBase64RawURL is the validation function for validating if the current field's value is a valid base64 URL safe string without '=' padding.
func isBase64RawURL(sm *v3.Schema, rv string) {
	sm.Pattern = base64RawURLRegexString
}

// isURI is the validation function for validating if the current field's value is a valid URI.
func isURI(sm *v3.Schema, rv string) {
	sm.Format = "uri"
}

// isURL is the validation function for validating if the current field's value is a valid URL.
func isURL(sm *v3.Schema, rv string) {
	sm.Format = "url"
}

// isHttpURL is the validation function for validating if the current field's value is a valid HTTP(s) URL.
func isHttpURL(sm *v3.Schema, rv string) {
	sm.Format = "httpURL"
}

// isUrnRFC2141 is the validation function for validating if the current field's value is a valid URN as per RFC 2141.
func isUrnRFC2141(sm *v3.Schema, rv string) {
}

// isFile is the validation function for validating if the current field's value is a valid existing file path.
func isFile(sm *v3.Schema, rv string) {
	sm.Format = "file"
}

// isImage is the validation function for validating if the current field's value contains the path to a valid image file
func isImage(sm *v3.Schema, rv string) {
	sm.Format = "image"
}

// isFilePath is the validation function for validating if the current field's value is a valid file path.
func isFilePath(sm *v3.Schema, rv string) {
	sm.Format = "filepath"
}

// isE164 is the validation function for validating if the current field's value is a valid e.164 formatted phone number.
func isE164(sm *v3.Schema, rv string) {
	sm.Pattern = e164RegexString
}

// isEmail is the validation function for validating if the current field's value is a valid email address.
func isEmail(sm *v3.Schema, rv string) {
	sm.Format = "email"
}

// isHSLA is the validation function for validating if the current field's value is a valid HSLA color.
func isHSLA(sm *v3.Schema, rv string) {
	sm.Pattern = hslaRegexString
}

// isHSL is the validation function for validating if the current field's value is a valid HSL color.
func isHSL(sm *v3.Schema, rv string) {
	sm.Pattern = hslRegexString
}

// isRGBA is the validation function for validating if the current field's value is a valid RGBA color.
func isRGBA(sm *v3.Schema, rv string) {
	sm.Pattern = rgbaRegexString
}

// isRGB is the validation function for validating if the current field's value is a valid RGB color.
func isRGB(sm *v3.Schema, rv string) {
	sm.Pattern = rgbRegexString
}

// isHEXColor is the validation function for validating if the current field's value is a valid HEX color.
func isHEXColor(sm *v3.Schema, rv string) {
	sm.Pattern = hexColorRegexString
}

// isHexadecimal is the validation function for validating if the current field's value is a valid hexadecimal.
func isHexadecimal(sm *v3.Schema, rv string) {
	sm.Pattern = hexadecimalRegexString
}

// isNumber is the validation function for validating if the current field's value is a valid number.
func isNumber(sm *v3.Schema, rv string) {
}

// isNumeric is the validation function for validating if the current field's value is a valid numeric value.
func isNumeric(sm *v3.Schema, rv string) {

}

// isAlphanum is the validation function for validating if the current field's value is a valid alphanumeric value.
func isAlphanum(sm *v3.Schema, rv string) {
	sm.Pattern = alphaNumericRegexString
}

// isAlpha is the validation function for validating if the current field's value is a valid alpha value.
func isAlpha(sm *v3.Schema, rv string) {
	sm.Pattern = alphaRegexString
}

// isAlphanumUnicode is the validation function for validating if the current field's value is a valid alphanumeric unicode value.
func isAlphanumUnicode(sm *v3.Schema, rv string) {
	sm.Pattern = alphaUnicodeRegexString
}

// isAlphaUnicode is the validation function for validating if the current field's value is a valid alpha unicode value.
func isAlphaUnicode(sm *v3.Schema, rv string) {
	sm.Pattern = alphaUnicodeRegexString
}

// isBoolean is the validation function for validating if the current field's value is a valid boolean value or can be safely converted to a boolean value.
func isBoolean(sm *v3.Schema, rv string) {

}

// isDefault is the opposite of required aka hasValue
func isDefault(sm *v3.Schema, rv string) {
}

// hasValue is the validation function for validating if the current field's value is not the default static value.
func hasValue(sm *v3.Schema, rv string) {

}

// hasNotZeroValue is the validation function for validating if the current field's value is not the zero value for its type.
func hasNotZeroValue(sm *v3.Schema, rv string) {

}

// requiredIf is the validation function
// The field under validation must be present and not empty only if all the other specified fields are equal to the value following with the specified field.
func requiredIf(sm *v3.Schema, rv string) {

}

// excludedIf is the validation function
// The field under validation must not be present or is empty only if all the other specified fields are equal to the value following with the specified field.
func excludedIf(sm *v3.Schema, rv string) {

}

// requiredUnless is the validation function
// The field under validation must be present and not empty only unless all the other specified fields are equal to the value following with the specified field.
func requiredUnless(sm *v3.Schema, rv string) {

}

// skipUnless is the validation function
// The field under validation must be present and not empty only unless all the other specified fields are equal to the value following with the specified field.
func skipUnless(sm *v3.Schema, rv string) {

}

// excludedUnless is the validation function
// The field under validation must not be present or is empty unless all the other specified fields are equal to the value following with the specified field.
func excludedUnless(sm *v3.Schema, rv string) {

}

// excludedWith is the validation function
// The field under validation must not be present or is empty if any of the other specified fields are present.
func excludedWith(sm *v3.Schema, rv string) {

}

// requiredWith is the validation function
// The field under validation must be present and not empty only if any of the other specified fields are present.
func requiredWith(sm *v3.Schema, rv string) {

}

// excludedWithAll is the validation function
// The field under validation must not be present or is empty if all of the other specified fields are present.
func excludedWithAll(sm *v3.Schema, rv string) {

}

// requiredWithAll is the validation function
// The field under validation must be present and not empty only if all of the other specified fields are present.
func requiredWithAll(sm *v3.Schema, rv string) {

}

// excludedWithout is the validation function
// The field under validation must not be present or is empty when any of the other specified fields are not present.
func excludedWithout(sm *v3.Schema, rv string) {

}

// requiredWithout is the validation function
// The field under validation must be present and not empty only when any of the other specified fields are not present.
func requiredWithout(sm *v3.Schema, rv string) {

}

// excludedWithoutAll is the validation function
// The field under validation must not be present or is empty when all of the other specified fields are not present.
func excludedWithoutAll(sm *v3.Schema, rv string) {

}

// requiredWithoutAll is the validation function
// The field under validation must be present and not empty only when all of the other specified fields are not present.
func requiredWithoutAll(sm *v3.Schema, rv string) {

}

// isGteField is the validation function for validating if the current field's value is greater than or equal to the field specified by the param's value.
func isGteField(sm *v3.Schema, rv string) {
}

// isGtField is the validation function for validating if the current field's value is greater than the field specified by the param's value.
func isGtField(sm *v3.Schema, rv string) {

}

// isGte is the validation function for validating if the current field's value is greater than or equal to the param's value.
func isGte(sm *v3.Schema, rv string) {

}

// isGt is the validation function for validating if the current field's value is greater than the param's value.
func isGt(sm *v3.Schema, rv string) {

}

// hasLengthOf is the validation function for validating if the current field's value is equal to the param's value.
func hasLengthOf(sm *v3.Schema, rv string) {

}

// hasMinOf is the validation function for validating if the current field's value is greater than or equal to the param's value.
func hasMinOf(sm *v3.Schema, rv string) {
	if sm.Type == "string" {
		v, _ := strconv.ParseInt(rv, 10, 64)
		sm.MinLength = v
	} else if sm.Type == "array" {
		v, _ := strconv.ParseInt(rv, 10, 64)
		sm.MinItems = v
	} else {
		v, _ := strconv.ParseFloat(rv, 64)
		sm.Minimum = v
	}
}

// isLteField is the validation function for validating if the current field's value is less than or equal to the field specified by the param's value.
func isLteField(sm *v3.Schema, rv string) {

}

// isLtField is the validation function for validating if the current field's value is less than the field specified by the param's value.
func isLtField(sm *v3.Schema, rv string) {

}

// isLte is the validation function for validating if the current field's value is less than or equal to the param's value.
func isLte(sm *v3.Schema, rv string) {

}

// isLt is the validation function for validating if the current field's value is less than the param's value.
func isLt(sm *v3.Schema, rv string) {

}

// hasMaxOf is the validation function for validating if the current field's value is less than or equal to the param's value.
func hasMaxOf(sm *v3.Schema, rv string) {
	if sm.Type == "string" {
		v, _ := strconv.ParseInt(rv, 10, 64)
		sm.MaxLength = v
	} else if sm.Type == "array" {
		v, _ := strconv.ParseInt(rv, 10, 64)
		sm.MaxItems = v
	} else {
		v, _ := strconv.ParseFloat(rv, 64)
		sm.Maximum = v
	}
}

// isTCP4AddrResolvable is the validation function for validating if the field's value is a resolvable tcp4 address.
func isTCP4AddrResolvable(sm *v3.Schema, rv string) {
	sm.Format = "TCP4"
}

// isTCP6AddrResolvable is the validation function for validating if the field's value is a resolvable tcp6 address.
func isTCP6AddrResolvable(sm *v3.Schema, rv string) {
	sm.Format = "TCP6"
}

// isTCPAddrResolvable is the validation function for validating if the field's value is a resolvable tcp address.
func isTCPAddrResolvable(sm *v3.Schema, rv string) {
	sm.Format = "TCP"
}

// isUDP4AddrResolvable is the validation function for validating if the field's value is a resolvable udp4 address.
func isUDP4AddrResolvable(sm *v3.Schema, rv string) {
	sm.Format = "UDP4"
}

// isUDP6AddrResolvable is the validation function for validating if the field's value is a resolvable udp6 address.
func isUDP6AddrResolvable(sm *v3.Schema, rv string) {
	sm.Format = "UDP6"
}

// isUDPAddrResolvable is the validation function for validating if the field's value is a resolvable udp address.
func isUDPAddrResolvable(sm *v3.Schema, rv string) {
	sm.Format = "UDP"
}

// isIP4AddrResolvable is the validation function for validating if the field's value is a resolvable ip4 address.
func isIP4AddrResolvable(sm *v3.Schema, rv string) {
	sm.Format = "IP4"
}

// isIP6AddrResolvable is the validation function for validating if the field's value is a resolvable ip6 address.
func isIP6AddrResolvable(sm *v3.Schema, rv string) {
	sm.Format = "IP6"
}

// isIPAddrResolvable is the validation function for validating if the field's value is a resolvable ip address.
func isIPAddrResolvable(sm *v3.Schema, rv string) {
	sm.Format = "IP"
}

// isUnixAddrResolvable is the validation function for validating if the field's value is a resolvable unix address.
func isUnixAddrResolvable(sm *v3.Schema, rv string) {
	sm.Format = "UNIX"
}

func isHostnameRFC952(sm *v3.Schema, rv string) {
	sm.Pattern = hostnameRegexStringRFC952
}

func isHostnameRFC1123(sm *v3.Schema, rv string) {
	sm.Pattern = hostnameRegexStringRFC1123
}

func isFQDN(sm *v3.Schema, rv string) {
	sm.Pattern = fqdnRegexStringRFC1123
}

// isDir is the validation function for validating if the current field's value is a valid existing directory.
func isDir(sm *v3.Schema, rv string) {

}

// isDirPath is the validation function for validating if the current field's value is a valid directory.
func isDirPath(sm *v3.Schema, rv string) {

}

// isJSON is the validation function for validating if the current field's value is a valid json string.
func isJSON(sm *v3.Schema, rv string) {
	sm.Format = "JSON"
}

// isJWT is the validation function for validating if the current field's value is a valid JWT string.
func isJWT(sm *v3.Schema, rv string) {
	sm.Pattern = jWTRegexString
}

// isHostnamePort validates a <dns>:<port> combination for fields typically used for socket address.
func isHostnamePort(sm *v3.Schema, rv string) {
	sm.Pattern = hostnameRegexStringRFC1123
}

// IsPort validates if the current field's value represents a valid port
func isPort(sm *v3.Schema, rv string) {
	sm.Pattern = "port"
}

// isLowercase is the validation function for validating if the current field's value is a lowercase string.
func isLowercase(sm *v3.Schema, rv string) {
	sm.Format = "lowercase"
}

// isUppercase is the validation function for validating if the current field's value is an uppercase string.
func isUppercase(sm *v3.Schema, rv string) {
	sm.Format = "uppercase"
}

// isDatetime is the validation function for validating if the current field's value is a valid datetime string.
func isDatetime(sm *v3.Schema, rv string) {
	sm.Format = "date-time"
}

// isTimeZone is the validation function for validating if the current field's value is a valid time zone string.
func isTimeZone(sm *v3.Schema, rv string) {
	sm.Format = "timezone"
}

// isIso3166Alpha2 is the validation function for validating if the current field's value is a valid iso3166-1 alpha-2 country code.
func isIso3166Alpha2(sm *v3.Schema, rv string) {
	sm.Format = "iso3166_1_alpha2"
}

// isIso3166Alpha2EU is the validation function for validating if the current field's value is a valid iso3166-1 alpha-2 European Union country code.
func isIso3166Alpha2EU(sm *v3.Schema, rv string) {
	sm.Format = "iso3166_1_alpha2_eu"
}

// isIso3166Alpha3 is the validation function for validating if the current field's value is a valid iso3166-1 alpha-3 country code.
func isIso3166Alpha3(sm *v3.Schema, rv string) {
	sm.Format = "iso3166_1_alpha3"
}

// isIso3166Alpha3EU is the validation function for validating if the current field's value is a valid iso3166-1 alpha-3 European Union country code.
func isIso3166Alpha3EU(sm *v3.Schema, rv string) {
	sm.Format = "iso3166_1_alpha3_eu"
}

// isIso3166AlphaNumeric is the validation function for validating if the current field's value is a valid iso3166-1 alpha-numeric country code.
func isIso3166AlphaNumeric(sm *v3.Schema, rv string) {
	sm.Format = "iso3166_1_alpha_numeric"
}

// isIso3166AlphaNumericEU is the validation function for validating if the current field's value is a valid iso3166-1 alpha-numeric European Union country code.
func isIso3166AlphaNumericEU(sm *v3.Schema, rv string) {
	sm.Format = "iso3166_1_alpha_numeric_eu"
}

// isIso31662 is the validation function for validating if the current field's value is a valid iso3166-2 code.
func isIso31662(sm *v3.Schema, rv string) {
	sm.Format = "iso3166_2"
}

// isIso4217 is the validation function for validating if the current field's value is a valid iso4217 currency code.
func isIso4217(sm *v3.Schema, rv string) {
	sm.Format = "iso4217"
}

// isIso4217Numeric is the validation function for validating if the current field's value is a valid iso4217 numeric currency code.
func isIso4217Numeric(sm *v3.Schema, rv string) {
	sm.Format = "iso4217_numeric"
}

// isBCP47LanguageTag is the validation function for validating if the current field's value is a valid BCP 47 language tag, as parsed by language.Parse
func isBCP47LanguageTag(sm *v3.Schema, rv string) {
	sm.Format = "bcp47language"
}

// isIsoBicFormat is the validation function for validating if the current field's value is a valid Business Identifier Code (SWIFT code), defined in ISO 9362
func isIsoBicFormat(sm *v3.Schema, rv string) {
	sm.Pattern = bicRegexString
}

// isSemverFormat is the validation function for validating if the current field's value is a valid semver version, defined in Semantic Versioning 2.0.0
func isSemverFormat(sm *v3.Schema, rv string) {
	sm.Pattern = semverRegexString
}

// isCveFormat is the validation function for validating if the current field's value is a valid cve id, defined in CVE mitre org
func isCveFormat(sm *v3.Schema, rv string) {
	sm.Pattern = cveRegexString
}

// isDnsRFC1035LabelFormat is the validation function
// for validating if the current field's value is
// a valid dns RFC 1035 label, defined in RFC 1035.
func isDnsRFC1035LabelFormat(sm *v3.Schema, rv string) {
	sm.Pattern = dnsRegexStringRFC1035Label
}

// isMongoDBObjectId is the validation function for validating if the current field's value is valid MongoDB ObjectID
func isMongoDBObjectId(sm *v3.Schema, rv string) {
	sm.Pattern = mongodbIdRegexString
}

// isMongoDBConnectionString is the validation function for validating if the current field's value is valid MongoDB Connection String
func isMongoDBConnectionString(sm *v3.Schema, rv string) {
	sm.Pattern = mongodbConnStringRegexString
}

// isSpiceDB is the validation function for validating if the current field's value is valid for use with Authzed SpiceDB in the indicated way
func isSpiceDB(sm *v3.Schema, rv string) {

}

// isCreditCard is the validation function for validating if the current field's value is a valid credit card number
func isCreditCard(sm *v3.Schema, rv string) {

}

// hasLuhnChecksum is the validation for validating if the current field's value has a valid Luhn checksum
func hasLuhnChecksum(sm *v3.Schema, rv string) {

}

// isCron is the validation function for validating if the current field's value is a valid cron expression
func isCron(sm *v3.Schema, rv string) {
	sm.Pattern = cronRegexString
}

// isEIN is the validation function for validating if the current field's value is a valid U.S. Employer Identification Number (EIN)
func isEIN(sm *v3.Schema, rv string) {
	sm.Pattern = einRegexString
}

func isValidateFn(sm *v3.Schema, rv string) {

}

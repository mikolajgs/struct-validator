package structvalidator

import (
	"fmt"
	"html"
	"reflect"
	"strconv"
	"strings"
)

// Optional configuration for validation:
// * RestrictFields defines what struct fields should be generated
// * OverwriteFieldTags can be used to overwrite tags for specific fields
// * OverwriteTagName sets tag used to define validation (default is "validation")
// * ValidateWhenSuffix will validate certain fields based on their name, eg. "PrimaryEmail" field will need to be a valid email
// * OverwriteFieldValues is to use overwrite values for fields, so these values are validated not the ones in struct
// * TextareaMinLenght is a length at which textarea should be used instead of input, automatically
// * IDPrefix - if added, an element will contain an 'id' attribute in form of prefix + field name
// * NamePrefix - use this to put a prefix in the 'name' attribute
type HTMLOptions struct {
	RestrictFields       map[string]bool
	OverwriteFieldTags   map[string]map[string]string
	OverwriteTagName     string
	ValidateWhenSuffix   bool
	IDPrefix             string
	NamePrefix           string
}

// GenerateHTMLInput takes a struct and generates HTML inputs for each of the fields, eg. <input> or <textarea>
func GenerateHTML(obj interface{}, options *HTMLOptions) (map[string]string) {
	v := reflect.ValueOf(obj)
	i := reflect.Indirect(v)
	s := i.Type()

	tagName := "validation"
	if options != nil && options.OverwriteTagName != "" {
		tagName = options.OverwriteTagName
	}
	
	fields := map[string]string{}

	for j := 0; j < s.NumField(); j++ {
		field := s.Field(j)
		fieldKind := field.Type.Kind()

		// check if only specified field should be checked
		if options != nil && len(options.RestrictFields) > 0 && !options.RestrictFields[field.Name] {
			continue
		}

		// generate only ints, string and bool
		if !isInt(fieldKind) && !isString(fieldKind) && !isBool(fieldKind) {
			continue
		}

		// 'id' attribute
		fieldIDAttr := ""
		if options.IDPrefix != "" {
			fieldIDAttr = fmt.Sprintf(` id="%s%s"`, options.IDPrefix, field.Name)
		}
		fieldNameAttr := fmt.Sprintf(` name="%s%s"`, options.NamePrefix, field.Name)

		if isBool(fieldKind) {
			fields[field.Name] = fmt.Sprintf(`<input type="checkbox"%s%s>`, fieldNameAttr, fieldIDAttr)
			continue
		}

		// get tag values
		tagVal := field.Tag.Get(tagName)
		tagRegexpVal := field.Tag.Get(tagName + "_regexp")
		if options != nil && len(options.OverwriteFieldTags) > 0 {
			if len(options.OverwriteFieldTags[field.Name]) > 0 {
				if options.OverwriteFieldTags[field.Name][tagName] != "" {
					tagVal = options.OverwriteFieldTags[field.Name][tagName]
				}
				if options.OverwriteFieldTags[field.Name][tagName+"_regexp"] != "" {
					tagRegexpVal = options.OverwriteFieldTags[field.Name][tagName+"_regexp"]
				}
			}
		}

		patternAttr := ""
		if tagRegexpVal != "" {
			patternAttr = fmt.Sprintf(` pattern="%s"`, html.EscapeString(tagRegexpVal))
		}
		validationAttrs, isEmail, isTextarea := getHTMLAttributesFromTag(tagVal)

		if options != nil && options.ValidateWhenSuffix {
			if strings.HasSuffix(field.Name, "Email") {
				isEmail = true
			}
			// Price not supported here yet
		}

		if isInt(fieldKind) {
			fields[field.Name] = fmt.Sprintf(`<input type="text"%s%s%s/>`, fieldNameAttr, fieldIDAttr, validationAttrs)
			continue
		}

		if isString(fieldKind) {
			if isTextarea {
				fields[field.Name] = fmt.Sprintf(`<textarea %s%s%s%s></textarea>`, fieldNameAttr, fieldIDAttr, validationAttrs, patternAttr)
				continue
			}
			fieldTypeAttr := ` type="text"`
			if isEmail {
				fieldTypeAttr = ` type="email"`
			}
			fields[field.Name] = fmt.Sprintf(`<input%s%s%s%s%s/>`, fieldTypeAttr, fieldNameAttr, fieldIDAttr, validationAttrs, patternAttr)
			continue
		}
	}

	return fields
}

func getHTMLAttributesFromTag(tag string) (string, bool, bool) {
	attrs := ""
	email := false
	textarea := false

	opts := strings.SplitN(tag, " ", -1)
	for _, opt := range opts {
		if opt == "req" {
			attrs = attrs + " required"
		}
		if opt == "email" {
			email = true
			continue
		}
		if opt == "uitextarea" {
			textarea = true
		}
		for _, valOpt := range []string{"lenmin", "lenmax", "valmin", "valmax", "regexp"} {
			if strings.HasPrefix(opt, valOpt+":") {
				val := strings.Replace(opt, valOpt+":", "", 1)
				if valOpt == "regexp" {
					attrs = attrs + fmt.Sprintf(` pattern="%s"`, html.EscapeString(val))
					continue
				}

				i, err := strconv.Atoi(val)
				if err != nil {
					continue
				}
				switch valOpt {
				case "lenmin":
					attrs = attrs + fmt.Sprintf(` minlength="%d"`, i)
				case "lenmax":
					attrs = attrs + fmt.Sprintf(` maxlength="%d"`, i)
				case "valmin":
					attrs = attrs + fmt.Sprintf(` min="%d"`, i)
				case "valmax":
					attrs = attrs + fmt.Sprintf(` max="%d"`, i)
				}
			}
		}
	}

	return attrs, email, textarea
}

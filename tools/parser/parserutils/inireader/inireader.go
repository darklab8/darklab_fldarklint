/*
Okay we need to create syntax. To augment currently possible stuff
*/
package inireader

import (
	"darktool/tools/utils"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type INIFile struct {
	File     utils.File
	Comments []string

	Sections []*Section

	// denormalization
	SectionMap map[string][]*Section
}

/*
[BaseGood] // this is Type
abc = 123 // this is Param going into list and hashmap
*/
type Section struct {
	Type   string
	Params []*Param
	// denormialization of Param list due to being more comfortable
	ParamMap map[string][]*Param
}

// abc = qwe, 1, 2, 3, 4
// abc is key
// qwe is first value
// qwe, 1, 2, 3, 4 are values
// ;abc = qwe, 1, 2, 3 is Comment
type Param struct {
	Key       string
	Values    []UniValue
	IsComment bool     // if commented out
	First     UniValue // denormalization due to very often being needed
}

type UniValue interface {
	AsString() string
}
type ValueString string
type ValueNumber struct {
	Value     float64
	Precision int
}

func (v ValueString) AsString() string {
	return string(v)
}

func (v ValueString) ToLowerValue() ValueString {
	return ValueString(strings.ToLower(string(v)))
}

func (v ValueNumber) AsString() string {
	return strconv.FormatFloat(float64(v.Value), 'f', v.Precision, 64)
}

func UniParse(input string) UniValue {

	numberMatch := regexNumber.FindAllString(input, -1)
	if len(numberMatch) > 0 {
		parsed_number, err := strconv.ParseFloat(input, 64)
		utils.CheckFatal(err, "failed to read number, input=", input)

		var precision int
		if len(numberMatch) == 1 {
			precision = 0
		} else {
			precision = len(numberMatch[1])
		}

		return ValueNumber{Value: parsed_number, Precision: precision}
	}

	v := ValueString(input)
	return v
}

var regexNumber *regexp.Regexp
var regexComment *regexp.Regexp
var regexSection *regexp.Regexp
var regexParam *regexp.Regexp

func init() {
	initRegexExpression(&regexNumber, `^[0-9\-]+(?:\.)?([0-9\-]*)`)
	initRegexExpression(&regexComment, `;(.*)`)
	initRegexExpression(&regexSection, `^\[.*\]`)
	// param or commented out param
	initRegexExpression(&regexParam, `(;?%?)([a-zA-Z_]+)\s=\s([a-zA-Z_, 0-9-]+)`)
}

func INIFileRead(fileref utils.File) INIFile {
	log.Debug("started reading INIFileRead for", fileref.Filepath)
	config := INIFile{File: fileref}

	log.Debug("opening file", fileref.Filepath)
	file := fileref.OpenToReadF()
	log.Debug("defer file close", fileref.Filepath)
	defer file.Close()

	log.Debug("reading lines")
	lines := file.ReadLines()

	log.Debug("setting current section")
	var cur_section *Section
	cur_section = &Section{}
	for _, line := range lines {

		comment_match := regexComment.FindStringSubmatch(line)
		section_match := regexSection.FindStringSubmatch(line)
		param_match := regexParam.FindStringSubmatch(line)

		if len(comment_match) > 0 {
			config.Comments = append(config.Comments, comment_match[1])
		} else if len(section_match) > 0 {

			config.Sections = append(config.Sections, cur_section)
			cur_section.Type = section_match[0]

			// Denormalization adding to hashmap
			key := section_match[0]
			if config.SectionMap == nil {
				config.SectionMap = make(map[string][]*Section)
			}
			if _, ok := config.SectionMap[key]; !ok {
				config.SectionMap[key] = make([]*Section, 0)
			}
			config.SectionMap[key] = append(config.SectionMap[key], cur_section)

			cur_section = &Section{} // create new
		} else if len(param_match) > 0 {
			isComment := len(param_match[1]) > 0
			key := param_match[2]
			splitted_values := strings.Split(param_match[3], ", ")
			first_value := UniParse(splitted_values[0])
			var values []UniValue
			for _, value := range splitted_values {
				values = append(values, UniParse(value))
			}

			// TODO add reading commented param
			param := Param{Key: key, First: first_value, Values: values, IsComment: isComment}
			cur_section.Params = append(cur_section.Params, &param)

			// Denormalization, adding to hashmap
			if cur_section.ParamMap == nil {
				cur_section.ParamMap = make(map[string][]*Param)
			}
			if _, ok := cur_section.ParamMap[key]; !ok {
				cur_section.ParamMap[key] = make([]*Param, 0)
			}
			cur_section.ParamMap[key] = append(cur_section.ParamMap[key], &param)
		}

	}

	return config
}

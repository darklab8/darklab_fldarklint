/*
parse universe.ini
*/
package universe

import (
	"darktool/tools/parser/parserutils/inireader"
	"darktool/tools/utils"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// terrain_tiny = nonmineable_asteroid90
// terrain_sml = nonmineable_asteroid60
// terrain_mdm = nonmineable_asteroid90
// terrain_lrg = nonmineable_asteroid60
// terrain_dyna_01 = mineable1_asteroid10
// terrain_dyna_02 = mineable1_asteroid10

var KEY_BASE_TERRAINS = [...]string{"terrain_tiny", "terrain_sml", "terrain_mdm", "terrain_lrg", "terrain_dyna_01", "terrain_dyna_02"}

const (
	FILENAME      = "universe.ini"
	KEY_BASE_TAG  = "[Base]"
	KEY_NICKNAME  = "nickname"
	KEY_STRIDNAME = "strid_name"
	KEY_SYSTEM    = "system"
	KEY_FILE      = "file"

	KEY_BASE_BGCS = "BGCS_base_run_by"

	KEY_SYSTEM_TAG           = "[system]"
	KEY_SYSTEM_MSG_ID_PREFIX = "msg_id_prefix"
	KEY_SYSTEM_VISIT         = "visit"
	KEY_SYSTEM_IDS_INFO      = "ids_info"
	KEY_SYSTEM_NAVMAPSCALE   = "NavMapScale"

	KEY_TIME_TAG     = "[Time]"
	KEY_TIME_SECONDS = "seconds_per_day"
)

type Time struct {
	seconds_per_day int
}

// Linux friendly filepath, that can be returned to Windows way from linux
type Path string

func PathCreate(input string) Path {
	input = strings.ReplaceAll(input, `\`, `/`)
	input = strings.ToLower(input)
	return Path(input)
}

func (p Path) LinuxPath() string {
	return string(p)
}

func (p Path) WindowsPath() string {
	return strings.ReplaceAll(string(p), `/`, `\`)
}

type Base struct {
	Nickname  string
	System    string
	StridName int
	// Danger. filepath in Windows File path system
	// Hopefully filepath will read it
	File             Path
	BGCS_base_run_by string

	Terrains map[string]string
}

type BaseNickname string

type SystemNickname string

type System struct {
	Nickname      string
	Pos           [2]int
	Msg_id_prefix string
	Visit         int
	Strid_name    int
	Ids_info      int
	File          Path
	NavMapScale   inireader.ValueNumber
}

type Config struct {
	Bases    []*Base
	BasesMap map[BaseNickname]*Base //key is

	Systems   []*System
	SystemMap map[SystemNickname]*System //key is

	Time Time
}

func (frelconfig *Config) AddBase(base_to_add *Base) {
	frelconfig.Bases = append(frelconfig.Bases, base_to_add)
	frelconfig.BasesMap[BaseNickname(base_to_add.Nickname)] = base_to_add
}

func (frelconfig *Config) AddSystem(system_to_add *System) {
	frelconfig.Systems = append(frelconfig.Systems, system_to_add)
	frelconfig.SystemMap[SystemNickname(system_to_add.Nickname)] = system_to_add
}

func (frelconfig *Config) Read(input_file *utils.File) (*Config, inireader.INIFile) {
	if frelconfig.BasesMap == nil {
		frelconfig.BasesMap = make(map[BaseNickname]*Base)
	}

	if frelconfig.Bases == nil {
		frelconfig.Bases = make([]*Base, 0)
	}

	iniconfig := inireader.INIFile.Read(inireader.INIFile{}, input_file)

	bases, ok := iniconfig.SectionMap[KEY_BASE_TAG]
	if !ok {
		log.Trace("failed to access iniconfig.SectionMap", KEY_BASE_TAG)
	}
	for _, base := range bases {
		base_to_add := Base{}

		check_nickname := base.ParamMap[KEY_NICKNAME][0].First.(inireader.ValueString).AsString()
		if !utils.IsLower(check_nickname) {
			log.Warn("nickname: ", check_nickname, "in file universe.txt is not in lower case. Autofixing")
		}
		base_to_add.Nickname = strings.ToLower(base.ParamMap[KEY_NICKNAME][0].First.AsString())
		strid, err := strconv.Atoi(base.ParamMap[KEY_STRIDNAME][0].First.AsString())
		if err != nil {
			log.Fatal("failed to parse strid in universe.ini for base=", base_to_add)
		}
		base_to_add.StridName = strid

		base_to_add.System = strings.ToLower(base.ParamMap[KEY_SYSTEM][0].First.AsString())
		base_to_add.File = PathCreate(base.ParamMap[KEY_FILE][0].First.AsString())

		if len(base.ParamMap[KEY_BASE_BGCS]) > 0 {
			base_to_add.BGCS_base_run_by = base.ParamMap[KEY_BASE_BGCS][0].First.AsString()
		}

		if base_to_add.Terrains == nil {
			base_to_add.Terrains = make(map[string]string)
		}
		for _, terrain_key := range KEY_BASE_TERRAINS {
			terrain_param, ok := base.ParamMap[terrain_key]
			if ok {
				base_to_add.Terrains[terrain_key] = terrain_param[0].First.AsString()
			}
		}

		frelconfig.AddBase(&base_to_add)
	}

	// Systems
	if frelconfig.SystemMap == nil {
		frelconfig.SystemMap = make(map[SystemNickname]*System)
	}

	if frelconfig.Systems == nil {
		frelconfig.Systems = make([]*System, 0)
	}
	systems, ok := iniconfig.SectionMap[KEY_SYSTEM_TAG]
	if !ok {
		log.Trace("failed to access iniconfig.SectionMap", KEY_SYSTEM_TAG)
	}
	for _, system := range systems {
		system_to_add := System{}

		system_to_add.Nickname = strings.ToLower(system.ParamMap[KEY_NICKNAME][0].First.AsString())

		if len(system.ParamMap[KEY_FILE]) > 0 {
			system_to_add.File = PathCreate(system.ParamMap[KEY_FILE][0].First.AsString())
		}

		if len(system.ParamMap[KEY_SYSTEM_MSG_ID_PREFIX]) > 0 {
			system_to_add.Msg_id_prefix = strings.ToLower(system.ParamMap[KEY_SYSTEM_MSG_ID_PREFIX][0].First.AsString())
		}

		if len(system.ParamMap[KEY_SYSTEM_VISIT]) > 0 {
			visits, err := strconv.Atoi(strings.ToLower(system.ParamMap[KEY_SYSTEM_VISIT][0].First.AsString()))
			if err == nil {
				system_to_add.Visit = visits
			}
		}

		if len(system.ParamMap[KEY_STRIDNAME]) > 0 {
			visits, err := strconv.Atoi(strings.ToLower(system.ParamMap[KEY_STRIDNAME][0].First.AsString()))
			if err == nil {
				system_to_add.Visit = visits
			}
		}

		if len(system.ParamMap[KEY_SYSTEM_IDS_INFO]) > 0 {
			visits, err := strconv.Atoi(strings.ToLower(system.ParamMap[KEY_SYSTEM_IDS_INFO][0].First.AsString()))
			if err == nil {
				system_to_add.Visit = visits
			}
		}

		if len(system.ParamMap[KEY_SYSTEM_NAVMAPSCALE]) > 0 {
			system_to_add.NavMapScale = system.ParamMap[KEY_SYSTEM_IDS_INFO][0].First.(inireader.ValueNumber)
		}

		frelconfig.AddSystem(&system_to_add)
	}

	// TIME
	seconds_per_day, err := strconv.Atoi(iniconfig.SectionMap[KEY_TIME_TAG][0].ParamMap[KEY_TIME_SECONDS][0].First.AsString())
	if err != nil {
		log.Fatal("unable to parse time in universe.ini")
	}
	frelconfig.Time = Time{seconds_per_day: seconds_per_day}

	return frelconfig, iniconfig
}

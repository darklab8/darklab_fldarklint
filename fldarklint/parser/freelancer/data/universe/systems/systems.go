package systems

import (
	"darklint/fldarklint/parser/freelancer/data/universe"
	"darklint/fldarklint/parser/parserutils/filefind"
	"darklint/fldarklint/parser/parserutils/filefind/file"
	"darklint/fldarklint/parser/parserutils/inireader"
	"darklint/fldarklint/parser/parserutils/semantic"
	"strings"

	"github.com/darklab8/darklab_goutils/goutils/utils/utils_types"
)

const (
	KEY_OBJECT   = "[Object]"
	KEY_NICKNAME = "nickname"
	KEY_BASE     = "base"
)

type Base struct {
	semantic.Model
	Nickname *semantic.String
	Base     *semantic.String // base.nickname in universe.ini
	DockWith *semantic.String
}
type System struct {
	semantic.ConfigModel
	Nickname    string
	Bases       []*Base
	BasesByNick map[string]*Base
	BasesByBase map[string]*Base
}

type Config struct {
	SystemsMap map[string]*System
	Systems    []*System
}

func (frelconfig *Config) Read(universe_config *universe.Config, filesystem filefind.Filesystem) *Config {

	var system_files map[string]*file.File = make(map[string]*file.File)
	for _, base := range universe_config.Bases {
		filename := universe_config.SystemMap[universe.SystemNickname(base.System.Get())].File.FileName()
		path := filesystem.GetFile(utils_types.FilePath(strings.ToLower(filename)))
		system_files[base.System.Get()] = file.NewFile(path.GetFilepath())
	}

	var system_iniconfigs map[string]inireader.INIFile = make(map[string]inireader.INIFile)
	for system_key, file := range system_files {
		system := inireader.INIFile{}
		system_iniconfigs[system_key] = inireader.INIFile.Read(system, file)
	}

	frelconfig.SystemsMap = make(map[string]*System)
	frelconfig.Systems = make([]*System, 0)
	for system_key, sysiniconf := range system_iniconfigs {
		system_to_add := System{}
		system_to_add.Init(sysiniconf.Sections, sysiniconf.Comments, sysiniconf.File.GetFilepath())

		system_to_add.Nickname = system_key
		system_to_add.BasesByNick = make(map[string]*Base)
		system_to_add.BasesByBase = make(map[string]*Base)
		system_to_add.Bases = make([]*Base, 0)
		frelconfig.SystemsMap[system_key] = &system_to_add
		frelconfig.Systems = append(frelconfig.Systems, &system_to_add)

		objects, ok := sysiniconf.SectionMap[KEY_OBJECT]
		if ok {
			for _, obj := range objects {

				// check if it is base object
				_, ok := obj.ParamMap[KEY_BASE]
				if ok {
					base_to_add := &Base{}
					base_to_add.Map(obj)

					base_to_add.Nickname = (&semantic.String{}).Map(obj, KEY_NICKNAME, semantic.TypeVisible, inireader.REQUIRED_p)
					base_to_add.Base = (&semantic.String{}).Map(obj, KEY_BASE, semantic.TypeVisible, inireader.REQUIRED_p)
					base_to_add.DockWith = (&semantic.String{}).Map(obj, "dock_with", semantic.TypeVisible, inireader.OPTIONAL_p)

					base_to_add.Nickname.Set(strings.ToLower(base_to_add.Nickname.Get()))
					base_to_add.Base.Set(strings.ToLower(base_to_add.Base.Get()))
					if base_to_add.DockWith.Get() != "" {
						base_to_add.DockWith.Set(strings.ToLower(base_to_add.DockWith.Get()))
					}

					system_to_add.BasesByBase[base_to_add.Base.Get()] = base_to_add
					system_to_add.BasesByNick[base_to_add.Nickname.Get()] = base_to_add
					system_to_add.Bases = append(system_to_add.Bases, base_to_add)
				}
			}
		}

	}

	return frelconfig
}

func (frelconfig *Config) Write() []*file.File {
	var files []*file.File = make([]*file.File, 0)
	for _, system := range frelconfig.Systems {
		inifile := system.Render()
		files = append(files, inifile.Write(inifile.File))
	}
	return files
}

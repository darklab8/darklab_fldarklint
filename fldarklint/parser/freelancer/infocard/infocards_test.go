package infocard

import (
	"darklint/fldarklint/parser/parserutils/filefind/file"
	"testing"

	"github.com/darklab8/darklab_goutils/goutils/utils"
	"github.com/darklab8/darklab_goutils/goutils/utils/utils_filepath"

	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {
	test_directory := utils.GetCurrrentTestFolder()
	fileref := file.NewFile(utils_filepath.Join(test_directory, FILENAME))
	config := Config{}
	config.Read(fileref)

	assert.Greater(t, len(config.Records), 0)
}

func TestInterfaceStruct(t *testing.T) {
	var record Record

	record = Infocard{}

	record = Name{}
	assert.True(t, record != nil)
}

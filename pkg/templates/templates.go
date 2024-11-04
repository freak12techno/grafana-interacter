package templates

import (
	"bytes"
	"html/template"
	"io/fs"
	"main/pkg/types/render"
	"main/pkg/utils"
	"time"
)

type TemplateManager struct {
	Timezone   *time.Location
	Templates  map[string]*template.Template
	Filesystem fs.FS
}

func NewTemplateManager(timezone *time.Location, filesystem fs.FS) *TemplateManager {
	return &TemplateManager{
		Templates:  make(map[string]*template.Template, 0),
		Timezone:   timezone,
		Filesystem: filesystem,
	}
}

func (manager *TemplateManager) GetTemplate(name string) (*template.Template, error) {
	if t, ok := manager.Templates[name]; ok {
		return t, nil
	}

	filename := name + ".html"

	t, err := template.New(filename).Funcs(template.FuncMap{
		"GetEmojiByStatus":        utils.GetEmojiByStatus,
		"GetEmojiBySilenceStatus": utils.GetEmojiBySilenceStatus,
		"StrToFloat64":            utils.StrToFloat64,
		"FormatDuration":          utils.FormatDuration,
		"FormatDate":              utils.FormatDate(manager.Timezone),
	}).ParseFS(manager.Filesystem, filename)
	if err != nil {
		return nil, err
	}

	manager.Templates[name] = t
	return t, nil
}

func (manager *TemplateManager) Render(name string, data render.RenderStruct) (string, error) {
	t, err := manager.GetTemplate(name)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	err = t.Execute(&buffer, data)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

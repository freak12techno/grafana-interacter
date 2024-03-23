package templates

import (
	"bytes"
	"html/template"
	"main/pkg/types/render"
	"main/pkg/utils"
	"time"

	templatesList "main/templates"
)

type TemplateManager struct {
	Timezone  *time.Location
	Templates map[string]*template.Template
}

func NewTemplateManager(timezone *time.Location) *TemplateManager {
	return &TemplateManager{
		Templates: make(map[string]*template.Template, 0),
		Timezone:  timezone,
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
	}).ParseFS(templatesList.Templates, filename)
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

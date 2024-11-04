package templates

import (
	"main/assets"
	"main/pkg/types/render"
	templatesList "main/templates"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTemplateRenderNotFound(t *testing.T) {
	t.Parallel()

	timezone, err := time.LoadLocation("Etc/GMT")
	require.NoError(t, err)

	manager := NewTemplateManager(timezone, templatesList.Templates)
	result, err := manager.Render("not-found", render.RenderStruct{})

	require.Error(t, err)
	require.Empty(t, result)
}

func TestTemplateRenderFailedToRender(t *testing.T) {
	t.Parallel()

	timezone, err := time.LoadLocation("Etc/GMT")
	require.NoError(t, err)

	manager := NewTemplateManager(timezone, assets.EmbedFS)
	result, err := manager.Render("template-invalid", render.RenderStruct{})

	require.Error(t, err)
	require.Empty(t, result)
}

func TestTemplateRenderOk(t *testing.T) {
	t.Parallel()

	timezone, err := time.LoadLocation("Etc/GMT")
	require.NoError(t, err)

	manager := NewTemplateManager(timezone, templatesList.Templates)
	result, err := manager.Render("help", render.RenderStruct{Data: "v1.2.3"})
	require.NoError(t, err)
	require.NotEmpty(t, result)

	result2, err2 := manager.Render("help", render.RenderStruct{Data: "v1.2.3"})
	require.NoError(t, err2)
	require.NotEmpty(t, result2)
}

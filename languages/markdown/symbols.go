package markdown

import (
	"fmt"
	"strings"

	"github.com/roveo/topo-mcp/languages"
)

// Heading represents a Markdown heading (# to ######)
type Heading struct {
	name  string          // The heading text
	level int             // 1-6 for # to ######
	loc   languages.Range // Range includes everything under this heading
}

func (h *Heading) Name() string              { return h.name }
func (h *Heading) Kind() string              { return fmt.Sprintf("h%d", h.level) }
func (h *Heading) Location() languages.Range { return h.loc }
func (h *Heading) String() string {
	return fmt.Sprintf("%s %s", strings.Repeat("#", h.level), h.name)
}

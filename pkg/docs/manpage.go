// ABOUTME: Man page generation for llmspell CLI commands.
// ABOUTME: Generates UNIX manual pages in troff format.

package docs

import (
	"fmt"
	"strings"
	"time"
)

// ManPage represents a manual page
type ManPage struct {
	Name        string
	Section     int
	Date        string
	Version     string
	Title       string
	Description string
	Synopsis    string
	Options     []Option
	Commands    []Command
	Examples    []Example
	Files       []string
	SeeAlso     []string
	Authors     []string
	Bugs        string
}

// Option represents a command-line option
type Option struct {
	Short       string
	Long        string
	Arg         string
	Description string
	Default     string
}

// Command represents a subcommand
type Command struct {
	Name        string
	Description string
	Options     []Option
	Examples    []Example
}

// Example represents a usage example
type Example struct {
	Command     string
	Description string
}

// NewManPage creates a new man page
func NewManPage(name string, section int, version string) *ManPage {
	return &ManPage{
		Name:    name,
		Section: section,
		Date:    time.Now().Format("January 2006"),
		Version: version,
		Title:   strings.ToUpper(name),
	}
}

// Generate generates the man page in troff format
func (m *ManPage) Generate() string {
	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf(".TH %s %d \"%s\" \"%s\" \"%s Manual\"\n",
		m.Title, m.Section, m.Date, m.Version, m.Name))

	// Name section
	b.WriteString(".SH NAME\n")
	b.WriteString(fmt.Sprintf("%s \\- %s\n", m.Name, m.Description))

	// Synopsis section
	if m.Synopsis != "" {
		b.WriteString(".SH SYNOPSIS\n")
		b.WriteString(".B " + m.Name + "\n")
		b.WriteString(m.Synopsis + "\n")
	}

	// Description section
	b.WriteString(".SH DESCRIPTION\n")
	b.WriteString(m.formatDescription(m.Description) + "\n")

	// Options section
	if len(m.Options) > 0 {
		b.WriteString(".SH OPTIONS\n")
		for _, opt := range m.Options {
			m.writeOption(&b, opt)
		}
	}

	// Commands section
	if len(m.Commands) > 0 {
		b.WriteString(".SH COMMANDS\n")
		for _, cmd := range m.Commands {
			b.WriteString(".SS " + cmd.Name + "\n")
			b.WriteString(cmd.Description + "\n")
			if len(cmd.Options) > 0 {
				b.WriteString(".PP\nOptions:\n")
				for _, opt := range cmd.Options {
					m.writeOption(&b, opt)
				}
			}
			if len(cmd.Examples) > 0 {
				b.WriteString(".PP\nExamples:\n")
				for _, ex := range cmd.Examples {
					b.WriteString(".PP\n.RS\n")
					b.WriteString(".nf\n")
					b.WriteString(ex.Command + "\n")
					b.WriteString(".fi\n")
					b.WriteString(".RE\n")
					if ex.Description != "" {
						b.WriteString(".PP\n")
						b.WriteString(ex.Description + "\n")
					}
				}
			}
		}
	}

	// Examples section
	if len(m.Examples) > 0 {
		b.WriteString(".SH EXAMPLES\n")
		for _, ex := range m.Examples {
			b.WriteString(".PP\n")
			if ex.Description != "" {
				b.WriteString(ex.Description + "\n")
				b.WriteString(".PP\n")
			}
			b.WriteString(".RS\n")
			b.WriteString(".nf\n")
			b.WriteString(ex.Command + "\n")
			b.WriteString(".fi\n")
			b.WriteString(".RE\n")
		}
	}

	// Files section
	if len(m.Files) > 0 {
		b.WriteString(".SH FILES\n")
		for _, file := range m.Files {
			b.WriteString(".TP\n")
			b.WriteString(".I " + file + "\n")
		}
	}

	// See also section
	if len(m.SeeAlso) > 0 {
		b.WriteString(".SH SEE ALSO\n")
		refs := make([]string, len(m.SeeAlso))
		for i, ref := range m.SeeAlso {
			refs[i] = ".BR " + ref
		}
		b.WriteString(strings.Join(refs, ",\n") + "\n")
	}

	// Authors section
	if len(m.Authors) > 0 {
		b.WriteString(".SH AUTHORS\n")
		b.WriteString(strings.Join(m.Authors, ", ") + "\n")
	}

	// Bugs section
	if m.Bugs != "" {
		b.WriteString(".SH BUGS\n")
		b.WriteString(m.Bugs + "\n")
	}

	return b.String()
}

// writeOption writes a single option to the man page
func (m *ManPage) writeOption(b *strings.Builder, opt Option) {
	b.WriteString(".TP\n")

	var optStr []string
	if opt.Short != "" {
		optStr = append(optStr, "\\fB\\-"+opt.Short+"\\fR")
	}
	if opt.Long != "" {
		escapedLong := strings.ReplaceAll(opt.Long, "-", "\\-")
		if opt.Arg != "" {
			escapedArg := strings.ReplaceAll(opt.Arg, "-", "\\-")
			optStr = append(optStr, "\\fB\\-\\-"+escapedLong+"\\fR=\\fI"+escapedArg+"\\fR")
		} else {
			optStr = append(optStr, "\\fB\\-\\-"+escapedLong+"\\fR")
		}
	}
	b.WriteString(strings.Join(optStr, ", ") + "\n")

	b.WriteString(opt.Description)
	if opt.Default != "" {
		b.WriteString(" (default: " + opt.Default + ")")
	}
	b.WriteString("\n")
}

// formatDescription formats description text for man pages
func (m *ManPage) formatDescription(desc string) string {
	// Replace multiple newlines with .PP for paragraph breaks
	paragraphs := strings.Split(desc, "\n\n")
	formatted := make([]string, len(paragraphs))
	for i, p := range paragraphs {
		formatted[i] = strings.TrimSpace(p)
	}
	return strings.Join(formatted, "\n.PP\n")
}

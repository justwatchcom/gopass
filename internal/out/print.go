package out

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
)

var (
	// Stdout is exported for tests
	Stdout io.Writer = os.Stdout
	// Stderr is exported for tests
	Stderr io.Writer = os.Stderr
)

var outputFilter = func(in string) string {
	return in
}

func init() {
	if sv := os.Getenv("GOPASS_NO_SYMBOLS"); sv != "" {
		outputFilter = func(in string) string {
			return strings.TrimFunc(in, func(r rune) bool {
				return unicode.IsSymbol(r)
			})
		}
	}
}

// Secret is a string wrapper for strings containing secrets. These won't be
// logged as long a GOPASS_DEBUG_LOG_SECRETS is not set
type Secret string

// SafeStr always return "(elided)"
func (s Secret) SafeStr() string {
	return "(elided)"
}

func newline(ctx context.Context) string {
	if HasNewline(ctx) {
		return "\n"
	}
	return ""
}

// Print prints the given string
func Print(ctx context.Context, arg interface{}) {
	Printf(ctx, "%s", arg)
}

// Printf formats and prints the given string
func Printf(ctx context.Context, format string, args ...interface{}) {
	if ctxutil.IsHidden(ctx) {
		return
	}
	debug.LogN(1, format, args...)
	fmt.Fprintf(Stdout, outputFilter(Prefix(ctx)+format+newline(ctx)), args...)
}

// Notice prints the string with an exclamation mark
func Notice(ctx context.Context, arg interface{}) {
	Noticef(ctx, "%s", arg)
}

// Noticef prints the string with an exclamation mark in front
func Noticef(ctx context.Context, format string, args ...interface{}) {
	if ctxutil.IsHidden(ctx) {
		return
	}
	debug.LogN(1, "NOTICE: "+format, args...)
	fmt.Fprintf(Stdout, outputFilter(Prefix(ctx)+"⚠ "+format+newline(ctx)), args...)
}

// Error prints the string with a red cross in front
func Error(ctx context.Context, arg interface{}) {
	Errorf(ctx, "%s", arg)
}

// Errorf prints the string in red to stderr
func Errorf(ctx context.Context, format string, args ...interface{}) {
	if ctxutil.IsHidden(ctx) {
		return
	}
	debug.LogN(1, "ERROR: "+format, args...)
	fmt.Fprint(Stderr, outputFilter(color.RedString(Prefix(ctx)+"❌ "+format+newline(ctx), args...)))
}

// OK prints the string with a green checkmark in front
func OK(ctx context.Context, arg interface{}) {
	OKf(ctx, "%s", arg)
}

// OKf prints the string in with an OK checkmark in front
func OKf(ctx context.Context, format string, args ...interface{}) {
	if ctxutil.IsHidden(ctx) {
		return
	}
	debug.LogN(1, "OK: "+format, args...)
	fmt.Fprintf(Stdout, outputFilter(Prefix(ctx)+"✅ "+format+newline(ctx)), args...)
}

// Warning prints the string with a warning sign in front
func Warning(ctx context.Context, arg interface{}) {
	Warningf(ctx, "%s", arg)
}

// Warningf prints the string in yellow to stderr and prepends a warning sign
func Warningf(ctx context.Context, format string, args ...interface{}) {
	if ctxutil.IsHidden(ctx) {
		return
	}
	debug.LogN(1, "WARNING: "+format, args...)
	fmt.Fprint(Stderr, outputFilter(color.YellowString(Prefix(ctx)+"⚠ "+format+newline(ctx), args...)))
}

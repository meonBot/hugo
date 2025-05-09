// Copyright 2024 The Hugo Authors. All rights reserved.
// Some functions in this file (see comments) is based on the Go source code,
// copyright The Go Authors and  governed by a BSD-style license.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package loggers

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"

	"github.com/bep/logg"
)

// newNoAnsiEscapeHandler creates a new noAnsiEscapeHandler
func newNoAnsiEscapeHandler(outWriter, errWriter io.Writer, noLevelPrefix bool, predicate func(*logg.Entry) bool) *noAnsiEscapeHandler {
	if predicate == nil {
		predicate = func(e *logg.Entry) bool { return true }
	}
	return &noAnsiEscapeHandler{
		noLevelPrefix: noLevelPrefix,
		outWriter:     outWriter,
		errWriter:     errWriter,
		predicate:     predicate,
	}
}

type noAnsiEscapeHandler struct {
	mu            sync.Mutex
	outWriter     io.Writer
	errWriter     io.Writer
	predicate     func(*logg.Entry) bool
	noLevelPrefix bool
}

func (h *noAnsiEscapeHandler) HandleLog(e *logg.Entry) error {
	if !h.predicate(e) {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	var w io.Writer
	if e.Level > logg.LevelInfo {
		w = h.errWriter
	} else {
		w = h.outWriter
	}

	var prefix string
	for _, field := range e.Fields {
		if field.Name == FieldNameCmd {
			prefix = fmt.Sprint(field.Value)
			break
		}
	}

	if prefix != "" {
		prefix = prefix + ": "
	}

	msg := stripANSI(e.Message)

	if h.noLevelPrefix {
		fmt.Fprintf(w, "%s%s", prefix, msg)
	} else {
		fmt.Fprintf(w, "%s %s%s", levelString[e.Level], prefix, msg)
	}

	for _, field := range e.Fields {
		if strings.HasPrefix(field.Name, reservedFieldNamePrefix) {
			continue
		}
		fmt.Fprintf(w, " %s %v", field.Name, field.Value)

	}
	fmt.Fprintln(w)

	return nil
}

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes ANSI escape codes from s.
func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

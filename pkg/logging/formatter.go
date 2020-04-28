// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/Sirupsen/logrus"
)

type withSourceWrapper struct {
	wrapped logrus.Formatter
}

// WithSource allows to wrap a logrus.Formatter, to specify standard formatting options,
// whilst adding the souce of the log, i.e. the caller.
func WithSource(wrapped logrus.Formatter) logrus.Formatter {
	return &withSourceWrapper{
		wrapped: wrapped,
	}
}

func (w *withSourceWrapper) Format(entry *logrus.Entry) ([]byte, error) {
	caller := getCaller()
	if caller == "" {
		caller = "[UNKNOWN]"
	}
	// We need to create a new entry with the updated Data, and then copy the Data over from
	// that into the original entry. Modifying Data directly could result in a race condition
	// in the map if the entry is used mutably concurrently.
	entry.Data = entry.WithField("source", caller).Data
	return w.wrapped.Format(entry)
}

func getCaller() string {
	programCounters := make([]uintptr, 8)
	// There are at least 5 stack frames before the actual log caller, so skip them.
	found := runtime.Callers(5, programCounters)
	frames := runtime.CallersFrames(programCounters[0:found])
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		fnName := frame.Function
		if strings.HasPrefix(fnName, "google3/third_party/golang/logrus") ||
			strings.HasPrefix(fnName, "google3/third_party/simulated_hospital/pkg/logging/logging.") {
			continue
		}
		// We're now out of all logging code so this is our caller.
		file := frame.File
		index := strings.LastIndex(file, "/go/src/")
		if index != -1 {
			file = file[index+8:]
		}
		index = strings.LastIndex(fnName, "/")
		if index != -1 {
			fnName = fnName[index+1:]
		}
		return fmt.Sprintf("%v(%v:%v)", fnName, file, frame.Line)
	}
	return ""
}

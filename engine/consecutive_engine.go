// Copyright 2022 Google LLC
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

package engine

import (
	"os"

	"github.com/google/yamlfmt"
)

// Engine that will process each file one by one consecutively.
type ConsecutiveEngine struct {
	LineSepCharacter string
	Formatter        yamlfmt.Formatter
	Quiet            bool
}

func (e *ConsecutiveEngine) FormatContent(content []byte) ([]byte, error) {
	return e.Formatter.Format(content)
}

func (e *ConsecutiveEngine) Format(paths []string) error {
	formatDiffs, formatErrs := e.formatAll(paths)
	if len(formatErrs) > 0 {
		return formatErrs
	}
	return formatDiffs.ApplyAll()
}

func (e *ConsecutiveEngine) Lint(paths []string) (*yamlfmt.EngineOutput, error) {
	formatDiffs, formatErrs := e.formatAll(paths)
	if len(formatErrs) > 0 {
		return nil, formatErrs
	}
	if formatDiffs.ChangedCount() == 0 {
		return nil, nil
	}

	message := "The following formatting differences were found:"
	if e.Quiet {
		message = "The following files had formatting differences:"
	}

	return &yamlfmt.EngineOutput{
		Message: message,
		Files:   formatDiffs,
		Quiet:   e.Quiet,
	}, nil
}

func (e *ConsecutiveEngine) DryRun(paths []string) (*yamlfmt.EngineOutput, error) {
	formatDiffs, formatErrs := e.formatAll(paths)
	if len(formatErrs) > 0 {
		return nil, formatErrs
	}

	output := &yamlfmt.EngineOutput{
		Files: formatDiffs,
		Quiet: e.Quiet,
	}
	if len(formatDiffs) > 0 && e.Quiet {
		output.Message = "The following files will have formatting changes:"
	}
	return output, nil
}

func (e *ConsecutiveEngine) formatAll(paths []string) (yamlfmt.FileDiffs, FormatErrors) {
	formatDiffs := yamlfmt.FileDiffs{}
	formatErrs := FormatErrors{}
	for _, path := range paths {
		fileDiff, err := e.formatFileContent(path)
		if err != nil {
			formatErrs = append(formatErrs, wrapFormatError(path, err))
			continue
		}
		formatDiffs.Add(fileDiff)
	}
	return formatDiffs, formatErrs
}

func (e *ConsecutiveEngine) formatFileContent(path string) (*yamlfmt.FileDiff, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	formatted, err := e.FormatContent(content)
	if err != nil {
		return nil, err
	}
	return &yamlfmt.FileDiff{
		Path: path,
		Diff: &yamlfmt.FormatDiff{
			Original:  string(content),
			Formatted: string(formatted),
			LineSep:   e.LineSepCharacter,
		},
	}, nil
}

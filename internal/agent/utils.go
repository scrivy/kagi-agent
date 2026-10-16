package agent

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/scrivy/kagi-agent/internal/config"
)

const (
	kagiDownloads string = "kagi-downloads"
)

func (a *Agent) inputDoneFunc(key tcell.Key) {
	switch key {
	case tcell.KeyEnter:
		a.input <- a.ui.input.GetText()
		a.ui.input.SetText("")
	default:
		a.logger.Warnf("unhandled inputDoneFunc: %#v", key)
	}
}

func (a *Agent) loadConfiguration(ctx context.Context) error {
	var err error
	a.cfg, err = config.Load()

	if err != nil {
		// config doesn't exist
		if errors.Is(err, fs.ErrNotExist) {
			a.output(`

Not authenticated.

To continue you'll need to enter your session link to authenticate with Kagi.

Go to Kagi settings, Account, and copy paste the session link down below.`)

			re := regexp.MustCompile(`token=([^&]+)`)
			/* TODO
			a.update(func() {
				a.ui.app.EnablePaste(true)
				a.ui.app.EnableMouse(false)
			})
			*/

			var token string
			for {
				link, err := a.readInput(ctx, "Enter session link: ", "")
				if err != nil {
					return fmt.Errorf("a.readInput(): %w", err)
				}

				match := re.FindStringSubmatch(link)
				if len(match) == 0 {
					a.output("\n\ntoken not found in session link, enter another?\n")
					continue
				}

				token = match[1]

				break
			}

			/* TODO
			a.update(func() {
				a.ui.app.EnablePaste(false)
				a.ui.app.EnableMouse(true)
			})
			*/

			a.cfg, err = config.New(token)
			if err != nil {
				return fmt.Errorf("config.New(): %w", err)
			}

			err = a.cfg.Save()
			if err != nil {
				return fmt.Errorf("cfg.Save(): %w", err)
			}
		} else {
			return fmt.Errorf("config.Load(): %w", err)
		}
	}

	a.cache, err = config.LoadCache()
	if err != nil {
		return fmt.Errorf("config.LoadCache(): %w", err)
	}

	return nil
}

func (a *Agent) selectModel(ctx context.Context) (string, error) {
	available := a.state.init.Models.Models
	if len(available) == 0 {
		return "", fmt.Errorf("no models available")
	}

	selection := make(chan string, 1)

	a.update(func() {
		list := tview.NewList().ShowSecondaryText(false)
		list.SetBorder(true).SetTitle(" select a model ")
		list.SetBorderPadding(0, 0, 1, 1)

		for i, model := range available {
			label := model.DisplayName
			if label == "" {
				label = model.ID
			}

			switch {
			case model.Retired:
				label = fmt.Sprintf("%s (retired)", label)
			case model.Deprecated:
				label = fmt.Sprintf("%s (deprecated)", label)
			}

			list.AddItem(label, model.ID, 0, func() {
				a.ui.pages.RemovePage("select-model")
				a.ui.app.SetFocus(a.ui.input)
				selection <- model.ID
			})

			if model.ID == a.cache.LastModel {
				list.SetCurrentItem(i)
			}
		}

		// center the list within the output area
		modal := tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(
				tview.NewFlex().
					SetDirection(tview.FlexRow).
					AddItem(nil, 0, 1, false).
					AddItem(list, 18, 1, true).
					AddItem(nil, 0, 1, false),
				50, 2, true).
			AddItem(nil, 0, 1, false)

		a.ui.pages.AddPage("select-model", modal, true, true)
		a.ui.app.SetFocus(list)
	})

	var model string

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case model = <-selection:
	}

	return model, nil
}

func (a *Agent) checkReasoningEffort() {
	if a.state.model.Parameters.ReasoningEffort == nil {
		a.cache.LastReasoningEffort = ""
		return
	}

	options := a.state.model.Parameters.ReasoningEffort.Options
	if !slices.Contains(options, a.cache.LastReasoningEffort) {
		a.cache.LastReasoningEffort = ""
	}
}

func (a *Agent) selectReasoningEffort(ctx context.Context, model string) error {
	if a.state.model.Parameters.ReasoningEffort == nil || len(a.state.model.Parameters.ReasoningEffort.Options) == 0 {
		a.cache.LastReasoningEffort = ""
		return nil
	}

	selection := make(chan string, 1)

	a.update(func() {
		list := tview.NewList().ShowSecondaryText(false)
		list.SetBorder(true).SetTitle(" reasoning effort? ")
		list.SetBorderPadding(0, 0, 1, 1)

		// add default
		list.AddItem("default", "", 0, func() {
			a.ui.pages.RemovePage("select-model")
			a.ui.app.SetFocus(a.ui.input)
			selection <- "" // empty line assumes default
		})
		list.SetCurrentItem(0)

		for i, effort := range a.state.model.Parameters.ReasoningEffort.Options {
			list.AddItem(effort, "", 0, func() {
				a.ui.pages.RemovePage("select-model")
				a.ui.app.SetFocus(a.ui.input)
				selection <- effort
			})

			if effort == a.cache.LastReasoningEffort {
				list.SetCurrentItem(i + 1)
			}
		}

		// center the list within the output area
		modal := tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(
				tview.NewFlex().
					SetDirection(tview.FlexRow).
					AddItem(nil, 0, 1, false).
					AddItem(list, 0, 1, true).
					AddItem(nil, 0, 1, false),
				50, 2, true).
			AddItem(nil, 0, 1, false)

		a.ui.pages.AddPage("select-model", modal, true, true)
		a.ui.app.SetFocus(list)
	})

	select {
	case <-ctx.Done():
		return ctx.Err()
	case a.cache.LastReasoningEffort = <-selection:
	}

	return nil
}

// TODO: split into separate parse and save functions?
// parseAndSaveFiles saves to the filesystem and returns map[fileExt][]file
func (a *Agent) parseAndSaveFiles(content string) (map[string][]string, error) {
	lines := strings.Split(content, "\n")

	files := make(map[string][]string)
	var file string
	var recording bool

	re := regexp.MustCompile("```(\\w*)")

	var fileExt string
	for _, line := range lines {
		match := re.FindStringSubmatch(line)
		if len(match) > 0 {
			if recording {
				recording = false

				files[fileExt] = append(files[fileExt], file)
				file = ""
			} else {
				recording = true

				fileExt = match[1]
				if fileExt == "" {
					fileExt = "out"
				}
			}

			continue
		}

		if recording {
			file += line + "\n"
		}
	}

	if len(files) == 0 {
		return nil, nil
	}

	// create kagi-downloads if it doesn't exist
	if _, err := os.Stat(kagiDownloads); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(kagiDownloads, 0755); err != nil {
				return nil, fmt.Errorf("os.MkdirAll(): %w", err)
			}
		} else {
			return nil, fmt.Errorf("os.Stat(): %w", err)
		}
	}

	a.logger.Debugf("files to save: %+v", files)

	title := strings.ReplaceAll(a.state.title, " ", "_")
	if title == "" {
		title = "patch"
	}

	i := 1
	for fileExt, extFiles := range files {
		for _, file := range extFiles {
			var path string
			for { // find first non existent filename
				fileName := fmt.Sprintf("%s_%d.%s", title, i, fileExt)
				path = filepath.Join(kagiDownloads, fileName)

				_, err := os.Stat(path)
				if err == nil {
					i++
					continue
				} else {
					if errors.Is(err, os.ErrNotExist) {
						break
					} else {
						return nil, fmt.Errorf("os.Stat(): %w", err)
					}
				}
			}

			err := os.WriteFile(path, []byte(file), 0644)
			if err != nil {
				return nil, fmt.Errorf("os.WriteFile(): %w", err)
			}

			a.output("saved " + path + "\n")
		}
	}

	return files, nil
}

// TODO: remove?
func (a *Agent) applyDiffs(ctx context.Context, diffs []string) error {
	if len(diffs) == 0 {
		return nil
	}
	a.ui.app.EnableMouse(true)
	defer a.ui.app.EnableMouse(false)

	a.output("\nlet's consider applying these diffs to the right (scrollable) ->\n\n")

	for _, diff := range diffs {
		a.update(func() {
			// a.ui.outputAlt.SetText(diff) TODO
		})

		apply, err := a.readInput(ctx, "apply?: ", "")
		if err != nil {
			return fmt.Errorf("a.readInput(): %w", err)
		}

		if apply != "yes" {
			continue
		}

		a.output("\n")

		cmd := exec.Command("patch", "-p1")
		cmd.Stdin = strings.NewReader(diff)

		output, err := cmd.CombinedOutput()
		if err != nil {
			a.output(fmt.Sprintf("patch -p1: %s\noutput:\n%s\n\n", err, output))
		}

		a.output("patch output:\n\n" + string(output) + "\n")

		_, err = a.readInput(ctx, "check patch output and continue?", "")
		if err != nil {
			return fmt.Errorf("a.readInput(): %w", err)
		}
	}

	a.update(func() {
		// a.ui.outputAlt.SetText("")
	})

	return nil
}

func (a *Agent) output(o string) {
	a.update(func() {
		fmt.Fprint(a.ui.output, o)
	})
}

/* TODO: remove
func (a *Agent) outputRight(o string) {
	a.update(func() {
		fmt.Fprint(a.ui.outputAlt, o)
	})
}
*/

func (a *Agent) update(f func()) {
	a.ui.app.QueueUpdateDraw(f)
}

func (a *Agent) listFilePathsInCwd() ([]string, error) {
	var files []string

	skipHidden := true
	toIgnore := make(map[string][]string) // map[dir]filesToIgnore
	ignoreFilenames := []string{".gitignore", ".agentignore"}

	err := filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking path %s: %w", err)
		}

		currentDir := filepath.Dir(path)
		_, ok := toIgnore[currentDir]
		if !ok {
			toIgnore[currentDir] = []string{kagiDownloads}
			a.logger.Debugf("1: toIgnore files initiated:\n%+v", toIgnore)
		}

		filesToIgnore := toIgnore[currentDir]
		filename := info.Name()
		a.logger.Debugf("listFilePathsInCwd(): path: %s, filename: %s, filesToIgnore: %+v", path, filename, filesToIgnore)

		// read files to ignore
		if slices.Contains(ignoreFilenames, filename) {
			lines, err := readFileLinesIfExists(path)
			if err != nil {
				return fmt.Errorf("readFileLinesIfExists(): %w", err)
			}
			filesToIgnore = append(filesToIgnore, lines...)
			toIgnore[currentDir] = filesToIgnore
			a.logger.Debugf("2: read files to ignore: %+v", filesToIgnore)
			return nil
		}

		for _, ignoreFilename := range filesToIgnore {
			if filename == ignoreFilename {
				if info.IsDir() {
					a.logger.Debug("3: skipping dir from filesToIgnore")
					return filepath.SkipDir
				} else {
					a.logger.Debug("4: ignoring file from filesToIgnore")
					return nil
				}
			}

			if strings.Contains(ignoreFilename, "*") {
				matches, err := filepath.Glob(filepath.Join(currentDir, ignoreFilename))
				if err != nil {
					return fmt.Errorf("filepath.Glob(): %w", err)
				}

				for _, match := range matches {
					if path == match {
						a.logger.Debugf("skip glob match: %s", match)
						return nil
					}
				}
			}
		}

		if skipHidden && len(filename) > 1 && string(filename[0]) == "." {
			if info.IsDir() {
				a.logger.Debug("5: skipping hidden directory")
				return filepath.SkipDir
			} else {
				a.logger.Debug("6: ignoring hidden file")
				return nil
			}
		}

		if info.IsDir() {
			a.logger.Debug("7: not including directory")
			return nil
		}

		files = append(files, path)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("filepath.Walk(): %w", err)
	}

	return files, nil
}

func readFileLinesIfExists(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		} else {
			return nil, fmt.Errorf("os.Readfile(): %w", err)
		}
	}
	text := strings.TrimSuffix(string(data), "\n")
	return strings.Split(text, "\n"), nil
}

func (a *Agent) readInput(ctx context.Context, preReadLabel, postReadLabel string) (string, error) {
	a.update(func() {
		a.ui.input.SetLabel(preReadLabel)
		a.ui.input.SetDisabled(false)
		a.ui.app.SetFocus(a.ui.input)
	})

	var line string

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case line = <-a.input:
	}

	a.update(func() {
		a.ui.input.SetLabel(postReadLabel)
		a.ui.input.SetDisabled(true)
	})

	return line, nil
}

// Write implements io.Writer
func (a *Agent) Write(p []byte) (int, error) {
	n, err := a.ui.output.Write(p)
	a.ui.app.Draw()
	return n, err
}

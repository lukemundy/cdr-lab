package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

type releaseMode string

const (
	modeCheck    releaseMode = "check"
	modeRehearse releaseMode = "rehearse"
	modeRelease  releaseMode = "release"
)

const (
	migrationFiles   = ":(glob)services/*/internal/migrations/**"
	eventGoldenFiles = "services/*/internal/app/testdata/events.golden.json"
)

type releasePlan struct {
	Mode            releaseMode `json:"mode"`
	GitHubRepo      string      `json:"githubRepo"`
	Commit          string      `json:"commit"`
	Version         string      `json:"version"`
	PreviousVersion string      `json:"previousVersion"`
	ImageTag        string      `json:"imageTag"`
	// ImagePrefix is the repository prefix of the target images, also the input of cdrlab up in the release mode.
	ImagePrefix     string          `json:"imagePrefix"`
	ChartVersion    string          `json:"chartVersion"`
	ChartRepository string          `json:"chartRepository"`
	Images          []planImage     `json:"images"`
	Charts          []string        `json:"charts"`
	Fresh           freshDeployment `json:"freshDeployment"`
	Changelog       string          `json:"changelog"`
}

// planImage is the image of a service. The release copies Source@Digest to Target:ImageTag.
type planImage struct {
	Name   string `json:"name"`
	Commit string `json:"commit"`
	Digest string `json:"digest"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type freshDeployment struct {
	First bool     `json:"first"`
	Files []string `json:"files"`
}

type planConfig struct {
	repo        gitRepo
	mainRef     string
	mode        releaseMode
	runID       string
	imagePrefix string
	githubRepo  string
	inputs      map[string][]string
	template    string
	ci          func(ctx context.Context, commit string) error
	client      *http.Client
}

// makePlan does each check of a release of rev, and gives the facts that the steps of the release need. It pushes
// nothing.
func makePlan(ctx context.Context, cfg planConfig, rev string) (releasePlan, error) {
	targetPrefix, err := targetPrefix(cfg)
	if err != nil {
		return releasePlan{}, err
	}

	commit, err := checkedOutCommit(ctx, cfg.repo, rev)
	if err != nil {
		return releasePlan{}, err
	}

	if err := checkFirstParent(ctx, cfg.repo, cfg.mainRef, commit); err != nil {
		return releasePlan{}, err
	}

	if err := cfg.ci(ctx, commit); err != nil {
		return releasePlan{}, err
	}

	previous, err := previousVersion(ctx, cfg.repo, commit)
	if err != nil {
		return releasePlan{}, err
	}

	version, err := nextVersion(ctx, cfg.repo, previous, commit)
	if err != nil {
		return releasePlan{}, err
	}

	p := releasePlan{
		Mode:            cfg.mode,
		GitHubRepo:      cfg.githubRepo,
		Commit:          commit,
		Version:         version,
		PreviousVersion: previous,
		ImageTag:        version,
		ImagePrefix:     targetPrefix,
		ChartVersion:    strings.TrimPrefix(version, "v"),
		ChartRepository: "oci://" + targetPrefix + "/charts",
	}

	if cfg.mode == modeRehearse {
		p.ImageTag = "rehearsal-" + cfg.runID
		p.ChartVersion = "0.0.0-rehearsal." + cfg.runID
	}

	if p.Images, err = planImages(ctx, cfg, commit, targetPrefix); err != nil {
		return releasePlan{}, err
	}

	if p.Charts, err = charts(cfg.repo); err != nil {
		return releasePlan{}, err
	}

	if p.Fresh, err = findFreshDeployment(ctx, cfg.repo, previous, commit); err != nil {
		return releasePlan{}, err
	}

	if p.Changelog, err = changelog(ctx, cfg.repo, cfg.template, previous, commit); err != nil {
		return releasePlan{}, err
	}

	return p, nil
}

// targetPrefix gives the repository prefix that the release pushes to. A rehearsal never pushes to the release names.
func targetPrefix(cfg planConfig) (string, error) {
	switch cfg.mode {
	case modeCheck, modeRelease:
		return cfg.imagePrefix, nil
	case modeRehearse:
		if _, err := strconv.ParseUint(cfg.runID, 10, 64); err != nil {
			return "", fmt.Errorf("a rehearsal needs the number of the workflow run as its run ID, not %q", cfg.runID)
		}

		return cfg.imagePrefix + "/rehearsal", nil
	default:
		return "", fmt.Errorf("the mode %q is not %s, %s or %s", cfg.mode, modeCheck, modeRehearse, modeRelease)
	}
}

// checkedOutCommit gives the commit of rev, and refuses when it is not HEAD, because the list of charts comes from the
// working tree.
func checkedOutCommit(ctx context.Context, repo gitRepo, rev string) (string, error) {
	commit, err := repo.commitID(ctx, rev)
	if err != nil {
		return "", err
	}

	head, err := repo.commitID(ctx, "HEAD")
	if err != nil {
		return "", err
	}

	if commit != head {
		return "", fmt.Errorf("check out the release commit %s first: HEAD is %s", commit, head)
	}

	return commit, nil
}

func checkFirstParent(ctx context.Context, repo gitRepo, mainRef, commit string) error {
	history, err := repo.run(ctx, "rev-list", "--first-parent", mainRef, "--")
	if err != nil {
		return err
	}

	if !slices.Contains(history, commit) {
		return fmt.Errorf("the commit %s is not in the first-parent history of %s", commit, mainRef)
	}

	return nil
}

func cog(ctx context.Context, repo gitRepo, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "cog", args...) //nolint:gosec // The arguments are a revision range and a path.
	cmd.Dir = repo.dir

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("cog %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}

	return string(out), nil
}

func planImages(ctx context.Context, cfg planConfig, commit, targetPrefix string) ([]planImage, error) {
	names := make([]string, 0, len(cfg.inputs))
	for name := range cfg.inputs {
		names = append(names, name)
	}

	slices.Sort(names)

	images := make([]planImage, 0, len(names))

	for _, name := range names {
		source := cfg.imagePrefix + "/" + name

		reg, err := newRegistry(cfg.client, source)
		if err != nil {
			return nil, err
		}

		imageCommitID, err := imageCommit(ctx, cfg.repo, func(tag string) (bool, error) { return reg.has(ctx, tag) },
			commit, cfg.inputs[name])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}

		digest, err := reg.digest(ctx, imageCommitID)
		if err != nil {
			return nil, err
		}

		images = append(images, planImage{
			Name: name, Commit: imageCommitID, Digest: digest, Source: source, Target: targetPrefix + "/" + name,
		})
	}

	return images, nil
}

// charts gives the name of each chart directory of the checked-out commit.
func charts(repo gitRepo) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(repo.dir, "charts", "*", "Chart.yaml"))
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(files))
	for _, file := range files {
		names = append(names, filepath.Base(filepath.Dir(file)))
	}

	return names, nil
}

// findFreshDeployment lists the changes from the earlier version that the stored data of a deployment can depend on.
func findFreshDeployment(ctx context.Context, repo gitRepo, previous, commit string) (freshDeployment, error) {
	if previous == "" {
		return freshDeployment{First: true}, nil
	}

	// With no rename detection, a rename is a delete of the old file.
	files, err := repo.run(ctx, "diff", "--name-only", "--no-renames", "--diff-filter=MD", previous, commit, "--",
		migrationFiles)
	if err != nil {
		return freshDeployment{}, err
	}

	tracked, err := repo.run(ctx, "ls-tree", "-r", "--name-only", previous, "--", "services")
	if err != nil {
		return freshDeployment{}, err
	}

	for _, golden := range tracked {
		if ok, _ := path.Match(eventGoldenFiles, golden); !ok {
			continue
		}

		changed, err := changedEntries(ctx, repo, previous, commit, golden)
		if err != nil {
			return freshDeployment{}, err
		}

		files = append(files, changed...)
	}

	if len(files) == 0 {
		files = nil
	}

	return freshDeployment{Files: files}, nil
}

// changedEntries gives the file when the commit deletes it, or else each entry of the earlier version that the commit
// changes or removes.
func changedEntries(ctx context.Context, repo gitRepo, previous, commit, file string) ([]string, error) {
	before, err := readGolden(ctx, repo, previous, file)
	if err != nil {
		return nil, err
	}

	after, err := readGolden(ctx, repo, commit, file)
	if errors.Is(err, errNoFile) {
		return []string{file}, nil
	}

	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(before))
	for key := range before {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	var changed []string

	for _, key := range keys {
		if value, ok := after[key]; !ok || !reflect.DeepEqual(value, before[key]) {
			changed = append(changed, file+": "+key)
		}
	}

	return changed, nil
}

var errNoFile = errors.New("the commit does not have the file")

func readGolden(ctx context.Context, repo gitRepo, commit, file string) (map[string]any, error) {
	data, ok, err := repo.file(ctx, commit, file)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errNoFile
	}

	var entries map[string]any
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("read %s at %s: %w", file, commit, err)
	}

	return entries, nil
}

func changelog(ctx context.Context, repo gitRepo, template, previous, commit string) (string, error) {
	return cog(ctx, repo, "changelog", previous+".."+commit, "--template", template)
}

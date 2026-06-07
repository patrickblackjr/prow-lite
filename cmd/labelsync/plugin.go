package labelsync

// FlagDef documents a config file field for generated docs.
type FlagDef struct {
	Name        string
	Description string
	Default     string
}

// PluginDef is the metadata for the labelsync plugin consumed by cmd/docgen.
var Plugin = struct {
	Name        string
	CLIName     string
	Description string
	Trigger     string
	Flags       []FlagDef
	Behavior    []string
}{
	Name:        "Label sync",
	CLIName:     "labelsync",
	Description: "Synchronises GitHub labels across all repositories accessible to the GitHub App installation. Labels are declared in a YAML file and applied, and optionally pruned, on every run.",
	Trigger:     "Run via `--plugin=labelsync` in CI mode, or on a schedule via the `prow-labelsync` workflow.",
	Flags: []FlagDef{
		{
			Name:        "overwrite",
			Description: "Overwrite existing labels whose color or description differs from the config.",
			Default:     "false",
		},
		{
			Name:        "prune",
			Description: "Delete labels that exist in the repository but are not defined in the config file.",
			Default:     "false",
		},
		{
			Name:        "dry_run",
			Description: "Log what would change without making any GitHub API mutations.",
			Default:     "false",
		},
	},
	Behavior: []string{
		"Lists all repositories accessible to the GitHub App installation; skips archived and forked repos.",
		"For each repository, creates any labels missing from the config.",
		"If `overwrite: true`, updates labels whose color or description has changed.",
		"If `prune: true`, deletes labels not present in the config file.",
		"API calls are parallelised with a concurrency limit of 50 to stay within GitHub secondary rate limits.",
	},
}

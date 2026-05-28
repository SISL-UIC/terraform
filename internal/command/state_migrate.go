// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform/internal/command/arguments"
	"github.com/hashicorp/terraform/internal/command/views"
	"github.com/hashicorp/terraform/internal/tfdiags"
)

// StateMigrateCommand is a Command implementation that migrates
// the state file from one location to another
type StateMigrateCommand struct {
	Meta
}

func (c *StateMigrateCommand) Run(rawArgs []string) int {
	// Parse and apply global view arguments
	common, rawArgs := arguments.ParseView(rawArgs)
	c.Meta.View.Configure(common)

	args, diags := arguments.ParseStateMigrate(rawArgs)

	stateMigrate := views.NewStateMigrate(args.ViewType, c.View)

	if diags.HasErrors() {
		stateMigrate.Diagnostics(diags)
		return 1
	}

	c.Meta.input = args.InputEnabled

	if args.SourceLockFilePath != "" {
		if _, err := os.Stat(args.SourceLockFilePath); err != nil {
			diags = diags.Append(tfdiags.Sourceless(
				tfdiags.Error,
				"Unreadable source provider lock file",
				fmt.Sprintf("%q: %s", args.SourceLockFilePath, err.Error()),
			))
		}
	}

	// It is valid for the destination lockfile to be missing
	// while state exists - e.g. through the use of builtin provider
	// or outputs and use of a builtin backend
	// (as opposed to pluggable state store).
	if args.DestinationLockFilePath != "" {
		if _, err := os.Stat(args.DestinationLockFilePath); err != nil {
			diags = diags.Append(tfdiags.Sourceless(
				tfdiags.Error,
				"Unreadable destination provider lock file",
				fmt.Sprintf("%q: %s", args.DestinationLockFilePath, err.Error()),
			))
		}
	}

	// return validation errors early if there are any
	if diags.HasErrors() {
		stateMigrate.Diagnostics(diags)
		return 1
	}

	c.Meta.includeStateMigrateFiles = true
	root, mDiags := c.Meta.loadSingleModule(".")
	if mDiags.HasErrors() {
		diags = diags.Append(mDiags)
		stateMigrate.Diagnostics(diags)
		return 1
	}

	smi := root.StateMigrationInstructions

	if smi == nil {
		diags = diags.Append(tfdiags.Sourceless(
			tfdiags.Error,
			"No state migration instructions found",
			"No instructions were found in the configuration files. Please ensure that a file with a .tfmigrate.hcl extension is present and contains valid state migration instructions.",
		))
		stateMigrate.Diagnostics(diags)
		return 1
	}

	var source string
	if smi.Backend != nil {
		source = fmt.Sprintf("backend %q", smi.Backend.Type)
	} else if smi.StateStore != nil {
		source = fmt.Sprintf("state store %q (%s)", smi.StateStore.Type,
			smi.StateStore.ProviderAddr)
	}
	var destination string
	if root.Backend != nil {
		destination = fmt.Sprintf("backend %q", root.Backend.Type)
	} else if root.StateStore != nil {
		destination = fmt.Sprintf("state store %q (%s)", root.StateStore.Type,
			root.StateStore.ProviderAddr)
	}

	stateMigrate.Log("Migrating state from %s to %s...", source, destination)

	// TODO: Load the source backend
	// TODO: Load the destination backend
	// TODO: Perform the migration from source to destination

	diags = diags.Append(errors.New("Not implemented yet"))
	stateMigrate.Diagnostics(diags)
	return 1
}

// func (c *StateMigrateCommand) srcBackendFromStateMigrationInstructions(root *configs.Module, locks *depsfile.Locks, viewType arguments.ViewType) (backendrun.OperationsBackend, tfdiags.Diagnostics) {
// 	var diags tfdiags.Diagnostics

// 	sm := root.StateMigrationInstructions

// 	var opts *BackendOpts
// 	switch {
// 	case sm.Backend != nil:
// 		opts = &BackendOpts{
// 			BackendConfig: sm.Backend,
// 			Locks:         locks,
// 			ViewType:      viewType,
// 		}
// 	case sm.StateStore != nil:
// 		// Annotate state_store config representation with info about how the provider
// 		// is supplied to Terraform.
// 		isReattached, err := reattach.IsProviderReattached(sm.StateStore.ProviderAddr, os.Getenv("TF_REATTACH_PROVIDERS"))
// 		if err != nil {
// 			panic(fmt.Sprintf("Unable to determine if provider %s is reattached while initializing the state store. This is a bug in Terraform and should be reported: %v", sm.StateStore.ProviderAddr.ForDisplay(), err))
// 		}
// 		sm.StateStore.ProviderSupplyMode = getproviders.DetermineProviderSupplyMode(c.Meta.isProviderDevOverride(sm.StateStore.ProviderAddr), isReattached, sm.StateStore.ProviderAddr.IsBuiltIn())

// 		// Check the provider for state storage is present, either via the dependency lock file or
// 		// supplied via developer overrides, reattach config, or being built-in.
// 		//
// 		// Remember, the (Meta).backend method is used for non-init commands, so we expect dependency locks
// 		// to be present or for the provider to be otherwise available, e.g. via reattach config.
// 		depsDiags := sm.StateStore.VerifyDependencySelection(locks, root.ProviderRequirements, sm.StateStore.ProviderSupplyMode)
// 		diags = diags.Append(depsDiags)
// 		if depsDiags.HasErrors() {
// 			return nil, diags
// 		}

// 		opts = &BackendOpts{
// 			StateStoreConfig: sm.StateStore,
// 			Locks:            locks,
// 			ViewType:         viewType,
// 		}
// 	default:
// 		diags = diags.Append(tfdiags.Sourceless(
// 			tfdiags.Error,
// 			"Missing state migration configuration",
// 			"One of 'backend' or 'state_store' blocks must be present in the state migration instructions.",
// 		))
// 		return nil, diags
// 	}

// 	// This method should not be used for init commands,
// 	// so we always set this value as false.
// 	opts.Init = false

// 	// Load the backend
// 	be, beDiags := c.Meta.Backend(opts)
// 	diags = diags.Append(beDiags)
// 	if beDiags.HasErrors() {
// 		return nil, diags
// 	}

// 	return be, diags
// }

func (c *StateMigrateCommand) Help() string {
	helpText := `
Usage: terraform [global options] state migrate [options]

  Migrate state from source declared in the migration configuration (*.tfmigrate.hcl)
  to the destination declared in the root module (*.tf).

  An error will be returned if the migration fails, e.g. if the state
  is inaccessible or the migration configuration is invalid.

Options:

  -source-provider-lock-file       Path to a provider lock file for the source provider (requires -input=false).
                                   Defaults to using the working directory's .terraform.lock.hcl file.

  -destination-provider-lock-file  Path to a provider lock file for the destination provider (requires -input=false).
                                   Defaults to using the working directory's .terraform.lock.hcl file.

  -upgrade                         Trigger upgrade of the provider used for state storage.

  -input=true                      Enable input for interactive prompts (defaults to true, set to false in automation).
`
	return strings.TrimSpace(helpText)
}

func (c *StateMigrateCommand) Synopsis() string {
	return "Migrate the state from one location to another"
}

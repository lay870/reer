package state

import (
	"path/filepath"

	"github.com/hashicorp/go-memdb"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl-lang/lang"
	"github.com/hashicorp/hcl/v2"
	tfaddr "github.com/hashicorp/terraform-registry-address"
	tfmod "github.com/hashicorp/terraform-schema/module"

	"github.com/hashicorp/terraform-ls/internal/terraform/ast"
	"github.com/hashicorp/terraform-ls/internal/terraform/datadir"
	op "github.com/hashicorp/terraform-ls/internal/terraform/module/operation"
)

type ModuleMetadata struct {
	CoreRequirements     version.Constraints
	ProviderReferences   map[tfmod.ProviderRef]tfaddr.Provider
	ProviderRequirements map[tfaddr.Provider]version.Constraints
	Variables            map[string]tfmod.Variable
	Outputs              map[string]tfmod.Output
}

func (mm ModuleMetadata) Copy() ModuleMetadata {
	newMm := ModuleMetadata{
		// version.Constraints is practically immutable once parsed
		CoreRequirements: mm.CoreRequirements,
	}

	if mm.ProviderReferences != nil {
		newMm.ProviderReferences = make(map[tfmod.ProviderRef]tfaddr.Provider, len(mm.ProviderReferences))
		for ref, provider := range mm.ProviderReferences {
			newMm.ProviderReferences[ref] = provider
		}
	}

	if mm.ProviderRequirements != nil {
		newMm.ProviderRequirements = make(map[tfaddr.Provider]version.Constraints, len(mm.ProviderRequirements))
		for provider, vc := range mm.ProviderRequirements {
			// version.Constraints is never mutated in this context
			newMm.ProviderRequirements[provider] = vc
		}
	}

	if mm.Variables != nil {
		newMm.Variables = make(map[string]tfmod.Variable, len(mm.Variables))
		for name, variable := range mm.Variables {
			newMm.Variables[name] = variable
		}
	}

	if mm.Outputs != nil {
		newMm.Outputs = make(map[string]tfmod.Output, len(mm.Outputs))
		for name, output := range mm.Outputs {
			newMm.Outputs[name] = output
		}
	}

	return newMm
}

type Module struct {
	Path string

	ModManifest      *datadir.ModuleManifest
	ModManifestErr   error
	ModManifestState op.OpState

	TerraformVersion      *version.Version
	TerraformVersionErr   error
	TerraformVersionState op.OpState

	ProviderSchemaErr   error
	ProviderSchemaState op.OpState

	RefTargets      lang.ReferenceTargets
	RefTargetsErr   error
	RefTargetsState op.OpState

	RefOrigins      lang.ReferenceOrigins
	RefOriginsErr   error
	RefOriginsState op.OpState

	ParsedModuleFiles  ast.ModFiles
	ParsedVarsFiles    ast.VarsFiles
	ModuleParsingErr   error
	VarsParsingErr     error
	ModuleParsingState op.OpState
	VarsParsingState   op.OpState

	Meta      ModuleMetadata
	MetaErr   error
	MetaState op.OpState

	ModuleDiagnostics ast.ModDiags
	VarsDiagnostics   ast.VarsDiags
}

func (m *Module) Copy() *Module {
	if m == nil {
		return nil
	}
	newMod := &Module{
		Path: m.Path,

		ModManifest:      m.ModManifest.Copy(),
		ModManifestErr:   m.ModManifestErr,
		ModManifestState: m.ModManifestState,

		// version.Version is practically immutable once parsed
		TerraformVersion:      m.TerraformVersion,
		TerraformVersionErr:   m.TerraformVersionErr,
		TerraformVersionState: m.TerraformVersionState,

		ProviderSchemaErr:   m.ProviderSchemaErr,
		ProviderSchemaState: m.ProviderSchemaState,

		RefTargets:      m.RefTargets.Copy(),
		RefTargetsErr:   m.RefTargetsErr,
		RefTargetsState: m.RefTargetsState,

		RefOrigins:      m.RefOrigins.Copy(),
		RefOriginsErr:   m.RefOriginsErr,
		RefOriginsState: m.RefOriginsState,

		ModuleParsingErr:   m.ModuleParsingErr,
		VarsParsingErr:     m.VarsParsingErr,
		ModuleParsingState: m.ModuleParsingState,
		VarsParsingState:   m.VarsParsingState,

		Meta:      m.Meta.Copy(),
		MetaErr:   m.MetaErr,
		MetaState: m.MetaState,
	}

	if m.ParsedModuleFiles != nil {
		newMod.ParsedModuleFiles = make(ast.ModFiles, len(m.ParsedModuleFiles))
		for name, f := range m.ParsedModuleFiles {
			// hcl.File is practically immutable once it comes out of parser
			newMod.ParsedModuleFiles[name] = f
		}
	}

	if m.ParsedVarsFiles != nil {
		newMod.ParsedVarsFiles = make(ast.VarsFiles, len(m.ParsedVarsFiles))
		for name, f := range m.ParsedVarsFiles {
			// hcl.File is practically immutable once it comes out of parser
			newMod.ParsedVarsFiles[name] = f
		}
	}

	if m.ModuleDiagnostics != nil {
		newMod.ModuleDiagnostics = make(ast.ModDiags, len(m.ModuleDiagnostics))
		for name, diags := range m.ModuleDiagnostics {
			newMod.ModuleDiagnostics[name] = make(hcl.Diagnostics, len(diags))
			for i, diag := range diags {
				// hcl.Diagnostic is practically immutable once it comes out of parser
				newMod.ModuleDiagnostics[name][i] = diag
			}
		}
	}

	if m.VarsDiagnostics != nil {
		newMod.VarsDiagnostics = make(ast.VarsDiags, len(m.VarsDiagnostics))
		for name, diags := range m.VarsDiagnostics {
			newMod.VarsDiagnostics[name] = make(hcl.Diagnostics, len(diags))
			for i, diag := range diags {
				// hcl.Diagnostic is practically immutable once it comes out of parser
				newMod.VarsDiagnostics[name][i] = diag
			}
		}
	}

	return newMod
}

func newModule(modPath string) *Module {
	return &Module{
		Path:                  modPath,
		ModManifestState:      op.OpStateUnknown,
		TerraformVersionState: op.OpStateUnknown,
		ProviderSchemaState:   op.OpStateUnknown,
		RefTargetsState:       op.OpStateUnknown,
		ModuleParsingState:    op.OpStateUnknown,
		MetaState:             op.OpStateUnknown,
	}
}

func (s *ModuleStore) Add(modPath string) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	// TODO: Introduce Exists method to Txn?
	obj, err := txn.First(s.tableName, "id", modPath)
	if err != nil {
		return err
	}
	if obj != nil {
		return &AlreadyExistsError{
			Idx: modPath,
		}
	}

	err = txn.Insert(s.tableName, newModule(modPath))
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) Remove(modPath string) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	_, err := txn.DeleteAll(s.tableName, "id", modPath)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) CallersOfModule(modPath string) ([]*Module, error) {
	txn := s.db.Txn(false)
	it, err := txn.Get(s.tableName, "id")
	if err != nil {
		return nil, err
	}

	callers := make([]*Module, 0)
	for item := it.Next(); item != nil; item = it.Next() {
		mod := item.(*Module)

		if mod.ModManifest == nil {
			continue
		}
		if mod.ModManifest.ContainsLocalModule(modPath) {
			callers = append(callers, mod)
		}
	}

	return callers, nil
}

func (s *ModuleStore) ModuleByPath(path string) (*Module, error) {
	txn := s.db.Txn(false)

	mod, err := moduleByPath(txn, path)
	if err != nil {
		return nil, err
	}

	return mod, nil
}

func (s *ModuleStore) ModuleCalls(modPath string) ([]tfmod.ModuleCall, error) {
	result := make([]tfmod.ModuleCall, 0)
	modList, err := s.List()
	for _, mod := range modList {
		if mod.ModManifest != nil {
			for _, record := range mod.ModManifest.Records {
				if record.IsRoot() {
					continue
				}
				result = append(result, tfmod.ModuleCall{
					LocalName:  record.Key,
					SourceAddr: record.SourceAddr,
					Path:       filepath.Join(modPath, record.Dir),
				})
			}
		}
	}
	return result, err
}

func (s *ModuleStore) ModuleMeta(modPath string) (*tfmod.Meta, error) {
	mod, err := s.ModuleByPath(modPath)
	if err != nil {
		return nil, err
	}
	return &tfmod.Meta{
		Path:                 mod.Path,
		ProviderReferences:   mod.Meta.ProviderReferences,
		ProviderRequirements: mod.Meta.ProviderRequirements,
		CoreRequirements:     mod.Meta.CoreRequirements,
		Variables:            mod.Meta.Variables,
		Outputs:              mod.Meta.Outputs,
	}, nil
}

func moduleByPath(txn *memdb.Txn, path string) (*Module, error) {
	obj, err := txn.First(moduleTableName, "id", path)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, &ModuleNotFoundError{
			Path: path,
		}
	}
	return obj.(*Module), nil
}

func moduleCopyByPath(txn *memdb.Txn, path string) (*Module, error) {
	mod, err := moduleByPath(txn, path)
	if err != nil {
		return nil, err
	}

	return mod.Copy(), nil
}

func (s *ModuleStore) List() ([]*Module, error) {
	txn := s.db.Txn(false)

	it, err := txn.Get(s.tableName, "id")
	if err != nil {
		return nil, err
	}

	modules := make([]*Module, 0)
	for item := it.Next(); item != nil; item = it.Next() {
		mod := item.(*Module)
		modules = append(modules, mod)
	}

	return modules, nil
}

func (s *ModuleStore) SetModManifestState(path string, state op.OpState) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	mod, err := moduleCopyByPath(txn, path)
	if err != nil {
		return err
	}

	mod.ModManifestState = state

	err = txn.Insert(s.tableName, mod)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) UpdateModManifest(path string, manifest *datadir.ModuleManifest, mErr error) error {
	txn := s.db.Txn(true)
	txn.Defer(func() {
		s.SetModManifestState(path, op.OpStateLoaded)
	})
	defer txn.Abort()

	mod, err := moduleCopyByPath(txn, path)
	if err != nil {
		return err
	}

	mod.ModManifest = manifest
	mod.ModManifestErr = mErr

	err = txn.Insert(s.tableName, mod)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) SetTerraformVersionState(path string, state op.OpState) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	mod, err := moduleCopyByPath(txn, path)
	if err != nil {
		return err
	}

	mod.TerraformVersionState = state
	err = txn.Insert(s.tableName, mod)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) SetProviderSchemaState(path string, state op.OpState) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	mod, err := moduleCopyByPath(txn, path)
	if err != nil {
		return err
	}

	mod.ProviderSchemaState = state
	err = txn.Insert(s.tableName, mod)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) FinishProviderSchemaLoading(path string, psErr error) error {
	txn := s.db.Txn(true)
	txn.Defer(func() {
		s.SetProviderSchemaState(path, op.OpStateLoaded)
	})
	defer txn.Abort()

	mod, err := moduleCopyByPath(txn, path)
	if err != nil {
		return err
	}

	mod.ProviderSchemaErr = psErr

	err = txn.Insert(s.tableName, mod)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) UpdateTerraformVersion(modPath string, tfVer *version.Version, pv map[tfaddr.Provider]*version.Version, vErr error) error {
	txn := s.db.Txn(true)
	txn.Defer(func() {
		s.SetTerraformVersionState(modPath, op.OpStateLoaded)
	})
	defer txn.Abort()

	mod, err := moduleCopyByPath(txn, modPath)
	if err != nil {
		return err
	}

	mod.TerraformVersion = tfVer
	mod.TerraformVersionErr = vErr

	err = txn.Insert(s.tableName, mod)
	if err != nil {
		return err
	}

	err = updateProviderVersions(txn, modPath, pv)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) SetModuleParsingState(path string, state op.OpState) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	mod, err := moduleCopyByPath(txn, path)
	if err != nil {
		return err
	}

	mod.ModuleParsingState = state
	err = txn.Insert(s.tableName, mod)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) SetVarsParsingState(path string, state op.OpState) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	mod, err := moduleCopyByPath(txn, path)
	if err != nil {
		return err
	}

	mod.VarsParsingState = state
	err = txn.Insert(s.tableName, mod)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) UpdateParsedModuleFiles(path string, pFiles ast.ModFiles, pErr error) error {
	txn := s.db.Txn(true)
	txn.Defer(func() {
		s.SetModuleParsingState(path, op.OpStateLoaded)
	})
	defer txn.Abort()

	mod, err := moduleCopyByPath(txn, path)
	if err != nil {
		return err
	}

	mod.ParsedModuleFiles = pFiles

	mod.ModuleParsingErr = pErr

	err = txn.Insert(s.tableName, mod)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) UpdateParsedVarsFiles(path string, vFiles ast.VarsFiles, vErr error) error {
	txn := s.db.Txn(true)
	txn.Defer(func() {
		s.SetVarsParsingState(path, op.OpStateLoaded)
	})
	defer txn.Abort()

	mod, err := moduleCopyByPath(txn, path)
	if err != nil {
		return err
	}

	mod.ParsedVarsFiles = vFiles

	mod.VarsParsingErr = vErr

	err = txn.Insert(s.tableName, mod)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) SetMetaState(path string, state op.OpState) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	mod, err := moduleCopyByPath(txn, path)
	if err != nil {
		return err
	}

	mod.MetaState = state
	err = txn.Insert(s.tableName, mod)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) UpdateMetadata(path string, meta *tfmod.Meta, mErr error) error {
	txn := s.db.Txn(true)
	txn.Defer(func() {
		s.SetMetaState(path, op.OpStateLoaded)
	})
	defer txn.Abort()

	mod, err := moduleCopyByPath(txn, path)
	if err != nil {
		return err
	}

	mod.Meta = ModuleMetadata{
		CoreRequirements:     meta.CoreRequirements,
		ProviderReferences:   meta.ProviderReferences,
		ProviderRequirements: meta.ProviderRequirements,
		Variables:            meta.Variables,
		Outputs:              meta.Outputs,
	}
	mod.MetaErr = mErr

	err = txn.Insert(s.tableName, mod)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) UpdateModuleDiagnostics(path string, diags ast.ModDiags) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	mod, err := moduleCopyByPath(txn, path)
	if err != nil {
		return err
	}

	mod.ModuleDiagnostics = diags

	err = txn.Insert(s.tableName, mod)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) UpdateVarsDiagnostics(path string, diags ast.VarsDiags) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	mod, err := moduleCopyByPath(txn, path)
	if err != nil {
		return err
	}

	mod.VarsDiagnostics = diags

	err = txn.Insert(s.tableName, mod)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) SetReferenceTargetsState(path string, state op.OpState) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	mod, err := moduleByPath(txn, path)
	if err != nil {
		return err
	}

	mod.RefTargetsState = state
	err = txn.Insert(s.tableName, mod)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) UpdateReferenceTargets(path string, refs lang.ReferenceTargets, rErr error) error {
	txn := s.db.Txn(true)
	txn.Defer(func() {
		s.SetReferenceTargetsState(path, op.OpStateLoaded)
	})
	defer txn.Abort()

	mod, err := moduleByPath(txn, path)
	if err != nil {
		return err
	}

	mod.RefTargets = refs
	mod.RefTargetsErr = rErr

	err = txn.Insert(s.tableName, mod)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) SetReferenceOriginsState(path string, state op.OpState) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	mod, err := moduleByPath(txn, path)
	if err != nil {
		return err
	}

	mod.RefOriginsState = state
	err = txn.Insert(s.tableName, mod)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *ModuleStore) UpdateReferenceOrigins(path string, origins lang.ReferenceOrigins, roErr error) error {
	txn := s.db.Txn(true)
	txn.Defer(func() {
		s.SetReferenceOriginsState(path, op.OpStateLoaded)
	})
	defer txn.Abort()

	mod, err := moduleByPath(txn, path)
	if err != nil {
		return err
	}

	mod.RefOrigins = origins
	mod.RefOriginsErr = roErr

	err = txn.Insert(s.tableName, mod)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

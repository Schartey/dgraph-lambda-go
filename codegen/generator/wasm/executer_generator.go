package wasm

/*func generateExecuter(c *config.Config, parsedTree *parser.Tree, pkgs map[string]*types.Package) error {

	f, err := os.Create(c.Exec.Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	var t *template.Template
	if c.Wasm.Lang == config.GOLANG {
		t = templates.GolangExecuterTemplate
	}

	err = t.Execute(f, struct {
		Interfaces  map[string]*parser.Interface
		Enums       map[string]*parser.Enum
		Scalars     map[string]*parser.Scalar
		Models      map[string]*parser.Model
		Packages    map[string]*types.Package
		PackageName string
	}{
		Interfaces:  parsedTree.ModelTree.Interfaces,
		Enums:       parsedTree.ModelTree.Enums,
		Scalars:     parsedTree.ModelTree.Scalars,
		Models:      parsedTree.ModelTree.Models,
		Packages:    pkgs,
		PackageName: c.Model.Package,
	})
	if err != nil {
		return err
	}
	return nil
}*/

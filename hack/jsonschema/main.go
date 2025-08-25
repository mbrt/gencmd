package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"

	"github.com/mbrt/gencmd/config"
)

func do() error {
	// Downgrade the schema version, as not all editors support the newer one:
	// (https://json-schema.org/draft/2020-12/schema)
	jsonschema.Version = "https://json-schema.org/draft-07/schema"
	r := jsonschema.Reflector{
		FieldNameTag: "yaml",
	}
	if err := r.AddGoComments("github.com/mbrt/gencmd", "./config/"); err != nil {
		return err
	}
	s := r.Reflect(config.Config{})
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func main() {
	if err := do(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

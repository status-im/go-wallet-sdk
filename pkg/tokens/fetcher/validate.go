package fetcher

import (
	"errors"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

var (
	ErrTokenListDoesNotMatchSchema = errors.New("token list does not match schema")
)

func validateJsonAgainstSchema(jsonData string, schema string) error {
	var schemaLoader gojsonschema.JSONLoader
	if strings.HasPrefix(schema, "http") {
		schemaLoader = gojsonschema.NewReferenceLoader(schema)
	} else {
		schemaLoader = gojsonschema.NewStringLoader(schema)
	}

	docLoader := gojsonschema.NewStringLoader(jsonData)

	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		return ErrTokenListDoesNotMatchSchema
	}

	return nil
}

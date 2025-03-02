package yema

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/token"
	"github.com/aep/yema"
)

// TypeToCue converts an abstract Type to a CUE value
func ToCue(t *yema.Type) (cue.Value, error) {
	if t == nil {
		return cue.Value{}, fmt.Errorf("nil type provided")
	}

	ctx := cuecontext.New()
	file := &ast.File{}

	if t.Kind != yema.Struct {
		return cue.Value{}, fmt.Errorf("expected root type to be Struct, got %v", t.Kind)
	}

	structExpr, err := typeToAstExpr(t, "")
	if err != nil {
		return cue.Value{}, err
	}

	file.Decls = append(file.Decls, &ast.EmbedDecl{Expr: structExpr})

	value := ctx.BuildFile(file)
	if value.Err() != nil {
		return cue.Value{}, fmt.Errorf("failed to build CUE value: %w", value.Err())
	}

	return value, nil
}

func typeToAstExpr(t *yema.Type, fieldName string) (ast.Expr, error) {
	switch t.Kind {
	case yema.Bool:
		return ast.NewIdent("bool"), nil
	case yema.Int:
		return ast.NewIdent("int"), nil
	case yema.Int8:
		return ast.NewIdent("int8"), nil
	case yema.Int16:
		return ast.NewIdent("int16"), nil
	case yema.Int32:
		return ast.NewIdent("int32"), nil
	case yema.Int64:
		return ast.NewIdent("int64"), nil
	case yema.Uint:
		return ast.NewIdent("uint"), nil
	case yema.Uint8:
		return ast.NewIdent("uint8"), nil
	case yema.Uint16:
		return ast.NewIdent("uint16"), nil
	case yema.Uint32:
		return ast.NewIdent("uint32"), nil
	case yema.Uint64:
		return ast.NewIdent("uint64"), nil
	case yema.Float32:
		return ast.NewIdent("float32"), nil
	case yema.Float64:
		return ast.NewIdent("float64"), nil
	case yema.String:
		return ast.NewIdent("string"), nil
	case yema.Bytes:
		return ast.NewIdent("string"), nil

	case yema.Array:
		if t.Array == nil {
			// Empty array
			return &ast.ListLit{
				Elts: []ast.Expr{
					&ast.Ellipsis{},
				},
			}, nil
		}

		// Get type of array elements
		elemExpr, err := typeToAstExpr(t.Array, fieldName)
		if err != nil {
			return nil, err
		}

		return &ast.ListLit{
			Elts: []ast.Expr{
				&ast.Ellipsis{Type: elemExpr},
			},
		}, nil

	case yema.Struct:
		if t.Struct == nil {
			return nil, fmt.Errorf("struct type with nil Struct field")
		}

		structLit := &ast.StructLit{
			Elts: []ast.Decl{},
		}

		for k, fieldType := range *t.Struct {
			label := ast.NewIdent(k)
			fieldExpr, err := typeToAstExpr(&fieldType, k)
			if err != nil {
				return nil, err
			}

			field := &ast.Field{
				Label:    label,
				Value:    fieldExpr,
				Token:    token.COLON,
				TokenPos: token.Blank.Pos(),
			}

			if fieldType.Optional {
				field.Constraint = token.OPTION
			} else {
				field.Constraint = token.NOT
			}

			structLit.Elts = append(structLit.Elts, field)
		}

		return structLit, nil

	default:
		return nil, fmt.Errorf("unexpected type kind: %v for field %s", t.Kind, fieldName)
	}
}

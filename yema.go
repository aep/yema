package yema

type Kind uint

const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Float32
	Float64
	Array
	Struct
	String
	Bytes
)

type Type struct {
	Kind     Kind
	Optional bool
	Struct   *map[string]Type
	Array    *Type
}

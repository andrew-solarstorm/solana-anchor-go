package anchor_idl

import (
	"encoding/json"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	. "github.com/gagliardetto/utilz"
)

// https://github.com/project-serum/anchor/blob/97e9e03fb041b8b888a9876a7c0676d9bb4736f3/ts/src/idl.ts
type IDL struct {
	Version      string           `json:"version"`
	Docs         []string         `json:"docs"` // @custom
	Instructions []IdlInstruction `json:"instructions"`
	State        *IdlState        `json:"state,omitempty"`
	Accounts     IdlTypeDefSlice  `json:"accounts,omitempty"`
	Types        IdlTypeDefSlice  `json:"types,omitempty"`
	Events       []IdlEvent       `json:"events,omitempty"`
	Errors       []IdlErrorCode   `json:"errors,omitempty"`
	Constants    []IdlConstant    `json:"constants,omitempty"`

	Address  string       `json:"address,omitempty"`
	Metadata *IdlMetadata `json:"metadata,omitempty"` // NOTE: deprecated
}

// TODO: write generator
type IdlConstant struct {
	Name  string   `json:"name"`
	Type  IdlType  `json:"type"`
	Value string   `json:"value"`
	Docs  []string `json:"docs"` // @custom
}

type IdlMetadata struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type IdlTypeDefSlice []IdlTypeDef

func (named IdlTypeDefSlice) GetByName(name string) *IdlTypeDef {
	for i := range named {
		v := named[i]
		if v.Name == name {
			return &v
		}
	}
	return nil
}

// Validate validates and IDL
func (idl *IDL) Validate() error {
	// TODO
	return nil
}

type IdlEvent struct {
	Name          string   `json:"name"`
	Discriminator *[8]byte `json:"discriminator,omitempty"`
}

type IdlInstruction struct {
	Name          string              `json:"name"`
	Discriminator *[8]byte            `json:"discriminator,omitempty"`
	Docs          []string            `json:"docs"` // @custom
	Accounts      IdlAccountItemSlice `json:"accounts"`
	Args          []IdlField          `json:"args"`
}

type IdlAccountItemSlice []IdlAccountItem

func (slice IdlAccountItemSlice) NumAccounts() (count int) {

	for _, item := range slice {
		if item.IdlAccount != nil {
			count++
		}

		if item.IdlAccounts != nil {
			count += item.IdlAccounts.Accounts.NumAccounts()
		}
	}

	return count
}

func (slice IdlAccountItemSlice) Walk(
	parentGroupPath string,
	previousIndex *int,
	parentGroup *IdlAccounts,
	callback func(string, int, *IdlAccounts, *IdlAccount) bool,
) {
	defaultVal := -1
	if previousIndex == nil {
		previousIndex = &defaultVal
	}
	for _, item := range slice {
		item.Walk(parentGroupPath, previousIndex, parentGroup, callback)
	}
}

type IdlState struct {
	Struct  IdlTypeDef       `json:"struct"`
	Methods []IdlStateMethod `json:"methods"`
}

type IdlStateMethod = IdlInstruction

// IdlAccountItem is of type IdlAccountItem = IdlAccount | IdlAccounts;
type IdlAccountItem struct {
	IdlAccount  *IdlAccount
	IdlAccounts *IdlAccounts
}

func (item *IdlAccountItem) Walk(
	parentGroupPath string,
	previousIndex *int,
	parentGroup *IdlAccounts,
	callback func(string, int, *IdlAccounts, *IdlAccount) bool,
) {
	defaultVal := -1
	if previousIndex == nil {
		previousIndex = &defaultVal
	}
	if item.IdlAccount != nil {
		*previousIndex++
		doContinue := callback(parentGroupPath, *previousIndex, parentGroup, item.IdlAccount)
		if !doContinue {
			return
		}
	}

	if item.IdlAccounts != nil {
		var thisGroupName string
		if parentGroupPath == "" {
			thisGroupName = item.IdlAccounts.Name
		} else {
			thisGroupName = parentGroupPath + "/" + item.IdlAccounts.Name
		}
		item.IdlAccounts.Accounts.Walk(thisGroupName, previousIndex, item.IdlAccounts, callback)
	}
}

// TODO: verify with examples
func (item *IdlAccountItem) UnmarshalJSON(data []byte) error {

	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if temp == nil {
		return fmt.Errorf("envelope is nil: %v", item)
	}

	switch v := temp.(type) {
	case map[string]interface{}:
		{
			// Ln(ShakespeareBG("::IdlAccountItem"))
			// spew.Dump(v)

			if len(v) == 0 {
				return nil
			}

			// Multiple accounts:
			if _, ok := v["accounts"]; ok {
				if err := TranscodeJSON(temp, &item.IdlAccounts); err != nil {
					return err
				}
			}
			// Single account:
			// TODO: check both writable and signer
			_, signer := v["signer"]
			_, writable := v["writable"]
			if signer || writable || v["address"] != "" {
				if err := TranscodeJSON(temp, &item.IdlAccount); err != nil {
					return err
				}
			} else {
				panic(Sf("what is this?:\n%s", spew.Sdump(temp)))
			}
		}
	default:
		return fmt.Errorf("unknown kind: %s", spew.Sdump(temp))
	}

	return nil
}

type IdlAccount struct {
	Docs     []string       `json:"docs"` // @custom
	Name     string         `json:"name"`
	Signer   bool           `json:"signer"`
	Writable bool           `json:"writable"`
	Optional bool           `json:"optional"`          // @custom
	Address  string         `json:"address,omitempty"` // constant address
	PDA      *idlAccountPDA `json:"pda,omitempty"`
}

type idlAccountPDA struct {
	Seeds   []idlAccountPDASeed `json:"seeds"`
	Program *idlAccountPDASeed  `json:"program,omitempty"`
}

type idlAccountPDASeed struct {
	Kind  string `json:"kind"`  // const or account
	Value []byte `json:"value"` // const
	Path  string `json:"path,omitempty"`
}

// IdlAccounts is a nested/recursive version of IdlAccount.
type IdlAccounts struct {
	Name     string              `json:"name"`
	Docs     []string            `json:"docs"` // @custom
	Accounts IdlAccountItemSlice `json:"accounts"`
}

type IdlField struct {
	Name string   `json:"name"`
	Docs []string `json:"docs"` // @custom
	Type IdlType  `json:"type"`
}

type IdlTypeAsString string

const (
	IdlTypeBool   IdlTypeAsString = "bool"
	IdlTypeU8     IdlTypeAsString = "u8"
	IdlTypeI8     IdlTypeAsString = "i8"
	IdlTypeU16    IdlTypeAsString = "u16"
	IdlTypeI16    IdlTypeAsString = "i16"
	IdlTypeU32    IdlTypeAsString = "u32"
	IdlTypeI32    IdlTypeAsString = "i32"
	IdlTypeU64    IdlTypeAsString = "u64"
	IdlTypeI64    IdlTypeAsString = "i64"
	IdlTypeU128   IdlTypeAsString = "u128"
	IdlTypeI128   IdlTypeAsString = "i128"
	IdlTypeBytes  IdlTypeAsString = "bytes"
	IdlTypeString IdlTypeAsString = "string"
	IdlTypePubkey IdlTypeAsString = "pubkey"
	IdlTypeF32    IdlTypeAsString = "f32"
	IdlTypeF64    IdlTypeAsString = "f64"

	// Custom additions:
	IdlTypeUnixTimestamp IdlTypeAsString = "unixTimestamp"
	IdlTypeHash          IdlTypeAsString = "hash"
	IdlTypeDuration      IdlTypeAsString = "duration"

	// | IdlTypeVec
	// | IdlTypeOption
	// | IdlTypeDefined;
)

type IdlTypeVec struct {
	Vec IdlType `json:"vec"`
}

type IdlTypeOption struct {
	Option IdlType `json:"option"`
}

type IdLTypeDefinedName struct {
	Name string `json:"name"`
}

// User defined type.
type IdlTypeDefined struct {
	Defined IdLTypeDefinedName `json:"defined"`
}

// IdlTypeArray is a Wrapper type:
type IdlTypeArray struct {
	Elem IdlType
	Num  int
}

func (env *IdlType) UnmarshalJSON(data []byte) error {

	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if temp == nil {
		return fmt.Errorf("envelope is nil: %v", env)
	}

	switch v := temp.(type) {
	case string:
		{
			env.AsString = IdlTypeAsString(v)
		}
	case map[string]interface{}:
		{
			// Ln(PurpleBG("::IdlType"))
			// spew.Dump(v)

			if len(v) == 0 {
				return nil
			}

			if _, ok := v["vec"]; ok {
				var target IdlTypeVec
				if err := TranscodeJSON(temp, &target); err != nil {
					return err
				}
				env.AsIdlTypeVec = &target
			}
			if _, ok := v["option"]; ok {
				var target IdlTypeOption
				if err := TranscodeJSON(temp, &target); err != nil {
					return err
				}
				env.AsIdlTypeOption = &target
			}
			if _, ok := v["defined"]; ok {
				var target IdlTypeDefined
				if err := TranscodeJSON(temp, &target); err != nil {
					return err
				}
				env.AsIdlTypeDefined = &target
			}
			if got, ok := v["array"]; ok {

				if _, ok := got.([]interface{}); !ok {
					panic(Sf("array is not in expected format:\n%s", spew.Sdump(got)))
				}
				arrVal := got.([]interface{})
				if len(arrVal) != 2 {
					panic(Sf("array is not of expected length:\n%s", spew.Sdump(got)))
				}
				var target IdlTypeArray
				if err := TranscodeJSON(arrVal[0], &target.Elem); err != nil {
					return err
				}

				target.Num = int(arrVal[1].(float64))

				env.AsIdlTypeArray = &target
			}
			// panic(Sf("what is this?:\n%s", spew.Sdump(temp)))
		}
	default:
		return fmt.Errorf("unknown kind: %s", spew.Sdump(temp))
	}

	return nil
}

// IdlType is a Wrapper type:
type IdlType struct {
	AsString         IdlTypeAsString
	AsIdlTypeVec     *IdlTypeVec
	AsIdlTypeOption  *IdlTypeOption
	AsIdlTypeDefined *IdlTypeDefined
	AsIdlTypeArray   *IdlTypeArray
}

func (env *IdlType) IsString() bool {
	return env.AsString != ""
}
func (env *IdlType) IsIdlTypeVec() bool {
	return env.AsIdlTypeVec != nil
}
func (env *IdlType) IsIdlTypeOption() bool {
	return env.AsIdlTypeOption != nil
}
func (env *IdlType) IsIdlTypeDefined() bool {
	return env.AsIdlTypeDefined != nil
}
func (env *IdlType) IsArray() bool {
	return env.AsIdlTypeArray != nil
}

// Getters:
func (env *IdlType) GetString() IdlTypeAsString {
	return env.AsString
}
func (env *IdlType) GetIdlTypeVec() *IdlTypeVec {
	return env.AsIdlTypeVec
}
func (env *IdlType) GetIdlTypeOption() *IdlTypeOption {
	return env.AsIdlTypeOption
}
func (env *IdlType) GetIdlTypeDefined() *IdlTypeDefined {
	return env.AsIdlTypeDefined
}
func (env *IdlType) GetArray() *IdlTypeArray {
	return env.AsIdlTypeArray
}
func (env *IdlType) GetDefinedFieldName() *string {
	if env.IsIdlTypeDefined() {
		return &env.AsIdlTypeDefined.Defined.Name
	}
	if env.IsIdlTypeVec() {
		return env.AsIdlTypeVec.Vec.GetDefinedFieldName()
	}
	if env.IsIdlTypeOption() {
		return env.AsIdlTypeOption.Option.GetDefinedFieldName()
	}
	if env.IsArray() {
		return env.AsIdlTypeArray.Elem.GetDefinedFieldName()
	}
	return nil
}

type IdlTypeDef struct {
	Name          string       `json:"name"`
	Type          IdlTypeDefTy `json:"type"`
	Docs          []string     `json:"docs,omitempty"`
	Discriminator *[8]byte     `json:"discriminator,omitempty"`
}

type IdlTypeDefTy struct {
	Kind     IdlTypeDefTyKind     `json:"kind"`
	Fields   *IdlStructFieldSlice `json:"fields,omitempty"`
	Variants *IdlEnumVariantSlice `json:"variants,omitempty"`
}

type IdlTypeDefTyKind string

const (
	IdlTypeDefTyKindStruct IdlTypeDefTyKind = "struct"
	IdlTypeDefTyKindEnum   IdlTypeDefTyKind = "enum"
)

type IdlStructFieldSlice []IdlField

type IdlEnumVariantSlice []IdlEnumVariant

func (slice IdlEnumVariantSlice) IsAllUint8() bool {
	for _, elem := range slice {
		if !elem.IsUint8() {
			return false
		}
	}
	return true
}

func (slice IdlEnumVariantSlice) IsSimpleEnum() bool {
	return slice.IsAllUint8()
}

func (slice IdlEnumVariantSlice) GetEnumVariantTypeName() []string {
	var result []string
	for _, variant := range slice {
		result = append(result, variant.Name)

	}
	return result
}

type IdlTypeDefStruct = []IdlField

type IdlEnumVariant struct {
	Name   string         `json:"name"`
	Docs   []string       `json:"docs"` // @custom
	Fields *IdlEnumFields `json:"fields,omitempty"`
}

func (variant *IdlEnumVariant) IsUint8() bool {
	// it's a simple uint8 if there is no fields data
	return variant.Fields == nil
}

// TODO
// type IdlEnumFields = IdlEnumFieldsNamed | IdlEnumFieldsTuple;
type IdlEnumFields struct {
	IdlEnumFieldsNamed *IdlEnumFieldsNamed
	IdlEnumFieldsTuple *IdlEnumFieldsTuple
}

type IdlEnumFieldsNamed []IdlField

type IdlEnumFieldsTuple []IdlType

// TODO: verify with examples
func (env *IdlEnumFields) UnmarshalJSON(data []byte) error {
	var tmp interface{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if tmp == nil {
		return fmt.Errorf("envelope is nil: %v", env)
	}

	fields, ok := tmp.([]interface{})
	if !ok {
		return fmt.Errorf("fields must be a slice")
	}

	if len(fields) == 0 {
		return nil
	}
	if m, ok := fields[0].(map[string]interface{}); ok && m["name"] != nil {
		// If has `name` field, then it's most likely a IdlEnumFieldsNamed.
		if err := TranscodeJSON(tmp, &env.IdlEnumFieldsNamed); err != nil {
			return err
		}
	} else {
		if err := TranscodeJSON(tmp, &env.IdlEnumFieldsTuple); err != nil {
			return err
		}
	}

	// panic(Sf("what is this?:\n%s", spew.Sdump(temp)))

	return nil
}

type IdlErrorCode struct {
	Code int    `json:"code"`
	Name string `json:"name"`
	Msg  string `json:"msg,omitempty"`
}

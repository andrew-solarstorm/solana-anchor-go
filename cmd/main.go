package main

import (
	"flag"
	"fmt"
	"time"

	anchor_idl "github.com/fragmetric-labs/solana-anchor-go"
	. "github.com/gagliardetto/utilz"
)

const generatedDir = "generated"

// TODO:
// - tests where type has field that is a complex enum (represented as an interface): assign a random concrete value from the possible enum variants.
// - when printing tree, check for len before accessing array indexes.

func main() {
	anchor_idl.Conf.Encoding = anchor_idl.EncodingBorsh
	anchor_idl.Conf.TypeID = anchor_idl.TypeIDAnchor

	filenames := FlagStringArray{}
	flag.Var(&filenames, "src", "Path to source; can use multiple times.")
	flag.StringVar(&anchor_idl.Conf.DstDir, "dst", generatedDir, "Destination folder")
	flag.StringVar(&anchor_idl.Conf.Package, "pkg", "", "Set package name to generate, default value is metadata.name of the source IDL.")
	flag.BoolVar(&anchor_idl.Conf.Debug, "debug", false, "debug mode")
	flag.BoolVar(&anchor_idl.Conf.RemoveAccountSuffix, "remove-account-suffix", false, "Remove \"Account\" suffix from accessors (if leads to duplication, e.g. \"SetFooAccountAccount\")")

	flag.StringVar((*string)(&anchor_idl.Conf.Encoding), "codec", string(anchor_idl.EncodingBorsh), "Choose codec")
	flag.StringVar((*string)(&anchor_idl.Conf.TypeID), "type-id", string(anchor_idl.TypeIDAnchor), "Choose typeID kind")
	flag.StringVar(&anchor_idl.Conf.ModPath, "mod", "", "Generate a go.mod file with the necessary dependencies, and this module")
	flag.Parse()

	if err := anchor_idl.Conf.Validate(); err != nil {
		panic(fmt.Errorf("error while validating config: %w", err))
	}

	var ts time.Time
	if anchor_idl.GetConfig().Debug {
		ts = time.Unix(0, 0)
	} else {
		ts = time.Now()
	}

	anchor_idl.GenerateFromIDLs(filenames, ts)
}

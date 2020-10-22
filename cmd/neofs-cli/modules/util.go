package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/nspcc-dev/neofs-api-go/pkg"
	"github.com/nspcc-dev/neofs-api-go/pkg/token"
	v2ACL "github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/spf13/cobra"
)

var (
	utilCmd = &cobra.Command{
		Use:   "util",
		Short: "Utility operations",
	}

	signCmd = &cobra.Command{
		Use:   "sign",
		Short: "sign NeoFS structure",
	}

	signBearerCmd = &cobra.Command{
		Use:   "bearer-token",
		Short: "sign bearer token to use it in requests",
		RunE:  signBearerToken,
	}

	convertCmd = &cobra.Command{
		Use:   "convert",
		Short: "convert representation of NeoFS structures",
	}

	convertEACLCmd = &cobra.Command{
		Use:   "eacl",
		Short: "convert representation of extended ACL table",
		RunE:  convertEACLTable,
	}
)

func init() {
	rootCmd.AddCommand(utilCmd)

	utilCmd.AddCommand(signCmd)
	utilCmd.AddCommand(convertCmd)

	signCmd.AddCommand(signBearerCmd)
	signBearerCmd.Flags().String("from", "", "File with JSON or binary encoded bearer token to sign")
	_ = signBearerCmd.MarkFlagFilename("from")
	_ = signBearerCmd.MarkFlagRequired("from")
	signBearerCmd.Flags().String("to", "", "File to dump signed bearer token (default: binary encoded)")
	signBearerCmd.Flags().Bool("json", false, "Dump bearer token in JSON encoding")

	convertCmd.AddCommand(convertEACLCmd)
	convertEACLCmd.Flags().String("from", "", "File with JSON or binary encoded extended ACL table")
	_ = convertEACLCmd.MarkFlagFilename("from")
	_ = convertEACLCmd.MarkFlagRequired("from")
	convertEACLCmd.Flags().String("to", "", "File to dump extended ACL table (default: binary encoded)")
	convertEACLCmd.Flags().Bool("json", false, "Dump extended ACL table in JSON encoding")
}

func signBearerToken(cmd *cobra.Command, _ []string) error {
	btok, err := getBearerToken(cmd, "from")
	if err != nil {
		return err
	}

	key, err := getKey()
	if err != nil {
		return err
	}

	err = completeBearerToken(btok)
	if err != nil {
		return err
	}

	err = btok.SignToken(key)
	if err != nil {
		return err
	}

	to := cmd.Flag("to").Value.String()
	jsonFlag, _ := cmd.Flags().GetBool("json")

	var data []byte
	if jsonFlag || len(to) == 0 {
		data = v2ACL.BearerTokenToJSON(btok.ToV2())
		if len(data) == 0 {
			return errors.New("can't JSON encode bearer token")
		}
	} else {
		data, err = btok.ToV2().StableMarshal(nil)
		if err != nil {
			return errors.New("can't binary encode bearer token")
		}
	}

	if len(to) == 0 {
		prettyPrintJSON(cmd, data)

		return nil
	}

	err = ioutil.WriteFile(to, data, 0644)
	if err != nil {
		return fmt.Errorf("can't write signed bearer token to file: %w", err)
	}

	cmd.Printf("signed bearer token was successfully dumped to %s\n", to)

	return nil
}

func convertEACLTable(cmd *cobra.Command, _ []string) error {
	pathFrom := cmd.Flag("from").Value.String()
	to := cmd.Flag("to").Value.String()
	jsonFlag, _ := cmd.Flags().GetBool("json")

	table, err := parseEACL(pathFrom)
	if err != nil {
		return err
	}

	var data []byte
	if jsonFlag || len(to) == 0 {
		data = v2ACL.TableToJSON(table.ToV2())
		if len(data) == 0 {
			return errors.New("can't JSON encode extended ACL table")
		}
	} else {
		data, err = table.ToV2().StableMarshal(nil)
		if err != nil {
			return errors.New("can't binary encode extended ACL table")
		}
	}

	if len(to) == 0 {
		prettyPrintJSON(cmd, data)

		return nil
	}

	err = ioutil.WriteFile(to, data, 0644)
	if err != nil {
		return fmt.Errorf("can't write exteded ACL table to file: %w", err)
	}

	cmd.Printf("extended ACL table was successfully dumped to %s\n", to)

	return nil
}

func completeBearerToken(btok *token.BearerToken) error {
	if v2 := btok.ToV2(); v2 != nil {
		// set eACL table version, because it usually omitted
		table := v2.GetBody().GetEACL()
		table.SetVersion(pkg.SDKVersion().ToV2())

		// back to SDK token
		btok = token.NewBearerTokenFromV2(v2)
	} else {
		return errors.New("unsupported bearer token version")
	}

	return nil
}

func prettyPrintJSON(cmd *cobra.Command, data []byte) {
	buf := new(bytes.Buffer)
	if err := json.Indent(buf, data, "", "  "); err != nil {
		printVerbose("Can't pretty print json: %w", err)
	}

	cmd.Println(buf)
}

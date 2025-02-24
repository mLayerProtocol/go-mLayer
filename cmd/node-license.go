package cmd

import (
	"fmt"

	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/internal/chain"
	"github.com/mlayerprotocol/go-mlayer/internal/chain/api"
	"github.com/spf13/cobra"
)




var nodeLicenseCmd = &cobra.Command{
	Use:   "license",
	Short: "View licenses & generate registration data for your node",
	Long: `Use this command to register and view assigned licenses:

	mLayer (message layer) is an open, decentralized 
	communication network that enables the creation, 
	transmission and termination of data of all sizes, 
	leveraging modern protocols. mLayer is a comprehensive 
	suite of communication protocols designed to evolve with 
	the ever-advancing realm of cryptography. 
	Visit the mLayer [documentation](https://mlayer.gitbook.io/introduction/what-is-mlayer) to learn more
	.`,
	// Run: accountFunc,
}

func init() {
	licenseListCmd.Flags().StringP(string(KEYSTORE_DIR), "K", "", "The keystore directory. This is the directory the keys are stored")
	licenseListCmd.Flags().StringP(string(PRIVATE_KEY), "k", "", "Node operator's private key")
	licenseListCmd.Flags().StringP(string(KEYSTORE_PASSWORD), "P", "", "Key store password")


	licenseRegisterCmd.Flags().StringP(string(KEYSTORE_DIR), "K", "", "The keystore directory. This is the directory the keys are stored")
	licenseRegisterCmd.Flags().StringP(string(KEYSTORE_PASSWORD), "P", "", "Key store password")
	licenseRegisterCmd.Flags().StringP(string(PRIVATE_KEY), "k", "", "Node operator's private key")
}


var licenseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all assigned licenses",
	Long: `List all licenses this nodes has been assigned to:

	mLayer (message layer) is an open, decentralized 
	communication network that enables the creation, 
	transmission and termination of data of all sizes, 
	leveraging modern protocols. mLayer is a comprehensive 
	suite of communication protocols designed to evolve with 
	the ever-advancing realm of cryptography. 
	Visit the mLayer [documentation](https://mlayer.gitbook.io/introduction/what-is-mlayer) to learn more
	.`,
	Run: licenseListFunc,
}



var licenseRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Generate registration data",
	Long: `Import private key or mnemonic to keystore:

	mLayer (message layer) is an open, decentralized 
	communication network that enables the creation, 
	transmission and termination of data of all sizes, 
	leveraging modern protocols. mLayer is a comprehensive 
	suite of communication protocols designed to evolve with 
	the ever-advancing realm of cryptography. 
	Visit the mLayer [documentation](https://mlayer.gitbook.io/introduction/what-is-mlayer) to learn more
	.`,
	Run: licenseRegisterFunc,
}


func licenseListFunc(_cmd *cobra.Command, _args []string) {
	cfg := configs.Config
	dir, _ := _cmd.Flags().GetString(string(KEYSTORE_DIR))	
	storeFilePath := getKeyStoreFilePath("account", dir)
	cfg = injectPrivateKey(&cfg, _cmd, storeFilePath)
	chain.RegisterProvider(
		"31337", api.NewGenericAPI(),
	)
	ethAPI, err := api.NewEthAPI(cfg.ChainId, cfg.EvmRpcConfig[string(cfg.ChainId)], &cfg.PrivateKeySECP)
	if err != nil {
		logger.Fatal(err)
	}
	chain.RegisterProvider(
		"84532", ethAPI,
	)
	chainIfo, err := ethAPI.GetChainInfo()
	if err != nil {
		logger.Fatal(err)
	}
	sentryLicense, err := ethAPI.GetSentryLicenses(cfg.PublicKeySECP, chainIfo.CurrentCycle)
	if err != nil {
		logger.Fatal(err)
	}
	
	fmt.Printf("\nS/N   |  SENTRY LICENSE ID [%d]\n", len(sentryLicense))
	fmt.Println("---------------------------")
	if len(sentryLicense) == 0 {
		println("0 assigned")
	}
	for i, license := range sentryLicense {
		fmt.Println(fmt.Sprintf("%d       %d", i+1, license))
	}

	fmt.Println()
	fmt.Println()

	valLicense, err := ethAPI.GetValidatorLicenses(cfg.PublicKeySECP, chainIfo.CurrentCycle)
	if err != nil {
		logger.Fatal(err)
	}
	
	fmt.Printf("\nS/N   |  VALIDATOR LICENSE ID [%d]\n", len(valLicense))
	fmt.Println("-----------------------------")
	if len(valLicense) == 0 {
		println("0 assigned")
	}
	for i, license := range valLicense {
		fmt.Println(fmt.Sprintf("%d       %d", i+1, license))
	}
	fmt.Println()
	fmt.Println()
	



}




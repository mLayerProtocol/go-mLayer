package cmd

// import (
// 	"crypto/ecdsa"
// 	"crypto/ed25519"
// 	"encoding/hex"
// 	"fmt"
// 	"log"
// 	"os"
// 	"strings"

// 	"github.com/btcsuite/btcd/btcec/v2"
// 	"github.com/ethereum/go-ethereum/accounts/keystore"
// 	"github.com/ethereum/go-ethereum/crypto"
// 	ethcrypto "github.com/ethereum/go-ethereum/crypto"
// 	"github.com/mlayerprotocol/go-mlayer/configs"
// 	mlCrypto "github.com/mlayerprotocol/go-mlayer/internal/crypto"
// 	"github.com/spf13/cobra"

// 	"github.com/tyler-smith/go-bip32"
// 	bip39 "github.com/tyler-smith/go-bip39"
// 	"golang.org/x/term"
// )

// // var teaProgram *tea.Program

// func printError(message string) {
// 	fmt.Print(formatError(message))
// }
// func formatError(message string) string {
// 	return fmt.Sprintf("\n\033[31m%s\033[0m", message)
// }

// var accountCmd = &cobra.Command{
// 	Use:   "account",
// 	Short: "Manage account for goml",
// 	Long: `Use this command to manage goml key pairs:

// 	mLayer (message layer) is an open, decentralized
// 	communication network that enables the creation,
// 	transmission and termination of data of all sizes,
// 	leveraging modern protocols. mLayer is a comprehensive
// 	suite of communication protocols designed to evolve with
// 	the ever-advancing realm of cryptography.
// 	Visit the mLayer [documentation](https://mlayer.gitbook.io/introduction/what-is-mlayer) to learn more
// 	.`,
// 	// Run: accountFunc,
// }

// var initCmd = &cobra.Command{
// 	Use:   "init",
// 	Short: "Initialize keystore",
// 	Long: `Initialize creates a new keystore:

// 	mLayer (message layer) is an open, decentralized
// 	communication network that enables the creation,
// 	transmission and termination of data of all sizes,
// 	leveraging modern protocols. mLayer is a comprehensive
// 	suite of communication protocols designed to evolve with
// 	the ever-advancing realm of cryptography.
// 	Visit the mLayer [documentation](https://mlayer.gitbook.io/introduction/what-is-mlayer) to learn more
// 	.`,
// 	Run: initFunc,
// }

// var importCmd = &cobra.Command{
// 	Use:   "import",
// 	Short: "Import private key",
// 	Long: `Import private key or mnemonic to keystore:

// 	mLayer (message layer) is an open, decentralized
// 	communication network that enables the creation,
// 	transmission and termination of data of all sizes,
// 	leveraging modern protocols. mLayer is a comprehensive
// 	suite of communication protocols designed to evolve with
// 	the ever-advancing realm of cryptography.
// 	Visit the mLayer [documentation](https://mlayer.gitbook.io/introduction/what-is-mlayer) to learn more
// 	.`,
// 	Run: importFunc,
// }

// func init() {
// 	accountCmd.AddCommand(initCmd)
// 	rootCmd.AddCommand(accountCmd)
// 	accountCmd.Flags().StringP(string(PRIVATE_KEY), "e", "", "The network private key. This is the key used to sign handshakes and messages")
// }

// func initFunc(_cmd *cobra.Command, _args []string) {
// 	cfg := configs.Config

// 	ksPath := cfg.KeyStoreDir
// 	if len(ksPath) == 0 {
// 		if len(cfg.DataDir) == 0 {
// 			cfg.DataDir = "./data"
// 		}
// 	if strings.HasSuffix(cfg.DataDir, "/") {
// 		ksPath = fmt.Sprintf("%skeystores/.goml", cfg.DataDir)
// 	} else {
// 		ksPath = fmt.Sprintf("%s/keystores/.goml", cfg.DataDir)
// 	}
// 	err := os.MkdirAll(ksPath, os.ModePerm)
// 		if err != nil {
// 			logger.Errorf("Error creating keystore directory at %s", ksPath)
// 			panic(err)
// 		}
// 	}
// 	// fmt.Printf("\nInitializing keystore at %s...", ksPath)
// 	//fmt.Println("")
// 	// interfaceRegistry := types.NewInterfaceRegistry()
// 	// cdc := codec.NewProtoCodec(interfaceRegistry)
// 	// kr, err := keyring.New("goml", keyring.BackendFile, ksPath, os.Stdin, cdc )
// 	// if err != nil {
//     //     fmt.Println("Error initializing keyring:", err)
//     //     return
//     // }

// 	// keyName := "primary" // replace with your key's name
// 	// _, err = kr.Key(keyName)
// 	// if err == nil {
// 	// 	fmt.Println("Key store already exists. Delete existing store? (y or n): ")
// 	// 	var del string
// 	// 	fmt.Scanf("%s", &del)
// 	// 	if !strings.EqualFold(del, "y") && !strings.EqualFold(del, "yes") {
// 	// 		return
// 	// 	}

// 	// }
// 	// curPass, err := term.ReadPassword(int(os.Stdin.Fd()))
// 	// 	if err != nil {
// 	// 		log.Fatal("Error reading keystore password:", err)
// 	// 	}
// 	// _, mnemonic, err := kr.NewMnemonic("my_key", keyring.English, sdk.FullFundraiserPath, string(curPass), hd.Secp256k1)
//     // if err != nil {
//     //     fmt.Println("Error creating new mnemonic:", err)
//     //     return
//     // }
// 	// logger.Debugf("Mnemonic: %s", mnemonic)

// 	// var one = 1;
// 	// if one == 1 {
// 	// 	return
// 	// }

// 	// strConfigPath := fmt.Sprintf("%s/active_keys.json", ksPath)
// 	// config := make(map[string]interface{})
// 	// config["0"] = ""
// 	// config["1"] = ""
// 	// curValue, err := utils.ReadJSONFromFile(strConfigPath)
// 	// if err == nil {
// 	// 	// panic(err)
// 	// 	config = curValue
// 	// }
// 	ks := keystore.NewKeyStore(ksPath, keystore.StandardScryptN, keystore.StandardScryptP)
// 	accounts := ks.Accounts()
// 	fmt.Printf("Initializing keystore at %s...!\n", ksPath)
// 	if len(accounts) > 0 {
// 		fmt.Printf("\033[31m%s\033[0m\n", "Error: Key store already exists. Please empty keystore directory to continue.")
// 		fmt.Println()
// 		return
// 		// var del string
// 		// // ask if user want to clear current keystore

// 		// fmt.Println("Key store already exists. Delete existing store? (y or n): ")
// 		// fmt.Scanf("%s %d", &del)
// 		// if !strings.EqualFold(del, "y") && !strings.EqualFold(del, "yes") {
// 		// 	return
// 		// }

// 		// fmt.Println("Enter existing key store password: ")
// 		// curPass, err := term.ReadPassword(int(os.Stdin.Fd()))
// 		// if err != nil {
// 		// 	log.Fatal("Error reading keystore password:", err)
// 		// }
// 		// var deleteError error
// 		// for _, account := range ks.Accounts() {
// 		// 	if account.Address == config["1"] {
// 		// 		err = ks.Delete(account, string(curPass))
// 		// 		if err != nil {
// 		// 			deleteError = err
// 		// 			break
// 		// 		}
// 		// 		break
// 		// 	}
// 		// }
// 		// if deleteError != nil {
// 		// 	fmt.Printf("Error deleting keystore: %s", err.Error())
// 		// 	fmt.Println("")
// 		// 	return
// 		// }
// 		// fmt.Println("Existing keystore deleted!")
// 	}
// 	 // Generate a new key
// 	fmt.Println("Enter new key store password: ")
// 	newPass, err := term.ReadPassword(int(os.Stdin.Fd()))
// 		if err != nil {
// 			log.Fatal("Error reading new keystore password:", err)
// 		}
// 		fmt.Println("Confirm new key store password: ")
// 		newPass2, err := term.ReadPassword(int(os.Stdin.Fd()))
// 		if err != nil {
// 			log.Fatal("Error reading new keystore password:", err)
// 		}
// 		if !strings.EqualFold(string(newPass), string(newPass2)) {
// 			fmt.Println("Passwords dont match!")
// 			return
// 		}
// 		fmt.Println("Initializing keystore...")
// 	account0, err := ks.NewAccount(string(newPass))
// 	 if err != nil {
// 		 log.Fatalf("Failed to create new account: %v", err)
// 	 }

// 	//  account1, err := ks.NewAccount(string(newPass))
// 	//  if err != nil {
// 	// 	log.Fatalf("Failed to create new account: %v", err)
// 	// }

// 	//  config["0"] = account0.Address.String()
// 	//  config["1"] = account1.Address.String()
// 	//  if err := utils.WriteJSONToFile(strConfigPath, config); err != nil {
// 	// 	panic(err)
// 	//  }

//     // Extract the private key
//     privateKeyBytes, err := mlCrypto.GetPrivateKeyFromKeyStore(ksPath, account0, string(newPass))
// 	if err != nil {
// 		log.Fatalf("Failed to create private key: %v", err)
// 		return
// 	}
// 	fmt.Println("Key store initialized successfully! Please backup your password!")
// 	_, pubKeySECP := btcec.PrivKeyFromBytes(privateKeyBytes)
// 	fmt.Printf("\n - Private Key: 0x%s", hex.EncodeToString(privateKeyBytes))
// 	fmt.Printf("\n - Payer Address: %s", account0.Address.String())
// 	fmt.Printf("\n - Licence Operator Public Key (SECP): %s", hex.EncodeToString(pubKeySECP.SerializeCompressed()))
// 	privateKeyBytesEDD := ed25519.NewKeyFromSeed(privateKeyBytes)
// 	publicKeyBytes := privateKeyBytesEDD[32:]
//    	fmt.Printf("\n - Network Public Key (EDD): %s", hex.EncodeToString(publicKeyBytes))

// 	copy(newPass,[]byte{})
// 	fmt.Println("")
// }

// // Import your private key or mnemonic
// func importFunc(_cmd *cobra.Command, _args []string) {
// 	cfg := configs.Config

// 	ksPath := cfg.KeyStoreDir
// 	if len(ksPath) == 0 {
// 		if len(cfg.DataDir) == 0 {
// 			cfg.DataDir = "./data"
// 		}
// 	if strings.HasSuffix(cfg.DataDir, "/") {
// 		ksPath = fmt.Sprintf("%skeystores/.goml", cfg.DataDir)
// 	} else {
// 		ksPath = fmt.Sprintf("%s/keystores/.goml", cfg.DataDir)
// 	}
// 	err := os.MkdirAll(ksPath, os.ModePerm)
// 		if err != nil {
// 			logger.Errorf("Error creating keystore directory at %s", ksPath)
// 			panic(err)
// 		}
// 	}

// 	ks := keystore.NewKeyStore(ksPath, keystore.StandardScryptN, keystore.StandardScryptP)
// 	accounts := ks.Accounts()
// 	fmt.Printf("Initializing keystore at %s...!\n", ksPath)
// 	if len(accounts) > 0 {
// 		fmt.Printf("\033[31m%s\033[0m\n", "Error: Key store already exists. Please empty keystore directory to continue.")
// 		fmt.Println()
// 		return
// 	}
// 	 // Generate a new key
// 	fmt.Println("Enter new key store password: ")
// 	newPass, err := term.ReadPassword(int(os.Stdin.Fd()))
// 		if err != nil {
// 			log.Fatal("Error reading new keystore password:", err)
// 		}
// 		fmt.Println("Confirm new key store password: ")
// 		newPass2, err := term.ReadPassword(int(os.Stdin.Fd()))
// 		if err != nil {
// 			log.Fatal("Error reading new keystore password:", err)
// 		}
// 		if !strings.EqualFold(string(newPass), string(newPass2)) {
// 			fmt.Println("Passwords dont match!")
// 			return
// 		}
// 		//fmt.Println("Initializing keystore...")

// 	fmt.Println("Paste private key or mnemonic: ")
// 	pKeyOrMnemonic, err := term.ReadPassword(int(os.Stdin.Fd()))
// 	var pKey *ecdsa.PrivateKey
// 	if len(pKeyOrMnemonic) != 32 && !strings.HasPrefix(strings.ToLower(string(pKeyOrMnemonic)), "0x") {
// 		if !bip39.IsMnemonicValid(string(pKeyOrMnemonic))  {
// 			fmt.Println(formatError("Invalid private key or mnemonic"))
// 			return
// 		}
// 		seed := bip39.NewSeed(string(pKeyOrMnemonic), "")
// 		pKey, err = derivePrivateKeyFromSeed(seed)
// 		if err != nil {
// 			fmt.Println(formatError("Invalid private key or mnemonic"))
// 			return
// 		}
// 	} else {
// 		pkHex := strings.Replace(string(pKeyOrMnemonic), "0x", "", 1)
// 		d, err := hex.DecodeString(pkHex)
// 		if err != nil {
// 			fmt.Println(formatError("Invalid private key or mnemonic"))
// 			return
// 		}
// 		pKey, err = crypto.ToECDSA(d)
// 		if err != nil {
// 			fmt.Println(formatError("Invalid private key or mnemonic"))
// 			return
// 		}
// 	}
//     // Derive the private key using a specific derivation path (BIP-44 standard)

// 	account, err := ks.ImportECDSA(pKey, string(newPass))

// 	 if err != nil {
// 		 log.Fatalf("Failed to create new account: %v", err)
// 	 }

// 	//  account1, err := ks.NewAccount(string(newPass))
// 	//  if err != nil {
// 	// 	log.Fatalf("Failed to create new account: %v", err)
// 	// }

// 	//  config["0"] = account0.Address.String()
// 	//  config["1"] = account1.Address.String()
// 	//  if err := utils.WriteJSONToFile(strConfigPath, config); err != nil {
// 	// 	panic(err)
// 	//  }

//     // Extract the private key
//     privateKeyBytes, err := mlCrypto.GetPrivateKeyFromKeyStore(ksPath, account0, string(newPass))
// 	if err != nil {
// 		log.Fatalf("Failed to create private key: %v", err)
// 		return
// 	}
// 	fmt.Println("Key store initialized successfully! Please backup your password!")
// 	_, pubKeySECP := btcec.PrivKeyFromBytes(privateKeyBytes)
// 	fmt.Printf("\n - Private Key: 0x%s", hex.EncodeToString(privateKeyBytes))
// 	fmt.Printf("\n - Payer Address: %s", account0.Address.String())
// 	fmt.Printf("\n - Licence Operator Public Key (SECP): %s", hex.EncodeToString(pubKeySECP.SerializeCompressed()))
// 	privateKeyBytesEDD := ed25519.NewKeyFromSeed(privateKeyBytes)
// 	publicKeyBytes := privateKeyBytesEDD[32:]
//    	fmt.Printf("\n - Network Public Key (EDD): %s", hex.EncodeToString(publicKeyBytes))

// 	copy(newPass,[]byte{})
// 	fmt.Println("")
// }

// // derivePrivateKeyFromSeed derives a private key from a seed and a specific derivation path
// func derivePrivateKeyFromSeed(seed []byte) (*ecdsa.PrivateKey, error) {
//     // Create a master key from the seed

//     masterKey, err := bip32.NewMasterKey(seed)
//     if err != nil {
//         log.Fatalf("Failed to generate master key: %v", err)
//     }
// 	path := []uint32{
//         bip32.FirstHardenedChild + 44, // Purpose: BIP-44
//         bip32.FirstHardenedChild + 60, // Coin type: Ethereum (60)
//         bip32.FirstHardenedChild + 0,  // Account: 0
//         0,                             // Change: External chain
//         0,                             // Address index: 0
//     }

//     chilaKey := masterKey
//     for _, n := range path {
//         chilaKey, err = chilaKey.NewChilaKey(n)
//         if err != nil {
//             log.Fatalf("Failed to derive child key: %v", err)
//         }
//     }

//     // Convert the BIP-32 private key to an ECDSA private key
//    return ethcrypto.ToECDSA(chilaKey.Key)

// }
// // func accountFunc(_cmd *cobra.Command, _args []string) {
// // 	//
// // 	// cfg := configs.Config
// // 	// ctx := context.Background()

// // 	operatorPrivateKey, err := _cmd.Flags().GetString(string(PRIVATE_KEY))
// // 	if err != nil || len(operatorPrivateKey) == 0 {
// // 		panic("operators private_key is required. Use --private-key flag or environment var ML_PRIVATE_KEY")
// // 	}

// // 	p := tea.NewProgram(localtea.InitialOptionModel([]string{"Existing private key", "Generate private key"}))
// // 	if _, err := p.Run(); err != nil {
// // 		fmt.Printf("Alas, there's been an error: %v", err)
// // 		os.Exit(1)
// // 	}
// // 	if localtea.GlobalWizardModel.Cursor == 0 {
// // 		p := tea.NewProgram(localtea.InitialOptionModel([]string{"From Memory", "New Key"}))
// // 		if _, err := p.Run(); err != nil {
// // 			fmt.Printf("Alas, there's been an error: %v", err)
// // 			os.Exit(1)
// // 		}
// // 		if localtea.GlobalWizardModel.Cursor == 0 {
// // 			key, err := localio.ReadFromFile(constants.DefaultPKeyPath)
// // 			if err != nil {
// // 				fmt.Println("Error running script:", err)
// // 				os.Exit(1)
// // 			}
// // 			scriptPath := "./scripts/dev.sh" // Replace with your script path
// // 			err = runScript(scriptPath, key)
// // 			if err != nil {
// // 				fmt.Println("Error running script:", err)
// // 				os.Exit(1)
// // 			} else {
// // 				fmt.Println("Script ran successfully!")
// // 			}
// // 		}
// // 		if localtea.GlobalWizardModel.Cursor == 1 {
// // 			//
// // 			p := tea.NewProgram(localtea.InitialInputModel())
// // 			if _, err := p.Run(); err != nil {
// // 				fmt.Printf("Alas, there's been an error: %v", err)
// // 				os.Exit(1)
// // 			}
// // 			key := localtea.GlobalInputWizardModel.TextInput.Value()
// // 			scriptPath := "./scripts/dev.sh" // Replace with your script path
// // 			err := runScript(scriptPath, key)
// // 			if err != nil {
// // 				fmt.Println("Error running script:", err)
// // 				os.Exit(1)
// // 			} else {
// // 				fmt.Println("Script ran successfully!")
// // 			}
// // 		}

// // 	} else if localtea.GlobalWizardModel.Cursor == 1 {
// // 		fmt.Printf("globalWizardModel.cursor %v:", localtea.GlobalWizardModel.Cursor)
// // 		publicKey, privateKey, err := localcrypto.GenerateKeyPair()
// // 		if err != nil {
// // 			printError(fmt.Sprintf("Keypair error: %s", err.Error()))
// // 			return
// // 		}
// // 		decodedPublicKey := localcrypto.KeyToString(publicKey)

// // 		decodedPrivateKey := localcrypto.KeyToString(privateKey)

// // 		teaProgram := tea.NewProgram(localtea.InitialOptionModel([]string{fmt.Sprintf("Public Key: %v", decodedPublicKey), fmt.Sprintf("Private Key: %v", decodedPrivateKey), "Continue"}))
// // 		if _, err := teaProgram.Run(); err != nil {
// // 			printError(fmt.Sprintf("Alas, theres been an error: %v", err))
// // 			os.Exit(1)
// // 		}

// // 		if localtea.GlobalWizardModel.Cursor == 2 {
// // 			scriptPath := "./scripts/dev.sh" // Replace with your script path
// // 			err := runScript(scriptPath, decodedPrivateKey)
// // 			if err != nil {
// // 				printError(fmt.Sprintf("Error: running script: %s", err.Error()))
// // 				os.Exit(1)
// // 			} else {
// // 				fmt.Println("Script ran successfully!")
// // 			}

// // 		}
// // 	}

// // }

// // func runScript(scriptPath string, key string) error {
// // 	err := localio.CreateOrUpdateFile(constants.DefaultPKeyPath, key)
// // 	if err != nil {
// // 		fmt.Println("Error running script:", err)
// // 		os.Exit(1)
// // 	}
// // 	cmd := exec.Command("/bin/sh", scriptPath, key)

// // 	cmd.Stdout = os.Stdout
// // 	cmd.Stderr = os.Stderr
// // 	return cmd.Run()
// // }

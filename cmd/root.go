// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/db"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	"github.com/spf13/cobra"
	"log"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "matrix-twitch-bridge",
	//TODO Descriptions
	Short: "",
	Long:  ``,
	PreRun: func(cmd *cobra.Command, args []string) {
		db.Init()
		log.Println("DB Set Up")
	},
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(util.CfgFile); os.IsNotExist(err) {
			asLogic.Init()
		} else {
			asLogic.Run()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&util.CfgFile, "config", "c", "./config.yaml", "config file (default is ./config.yaml . It will get generated if no value is given)")
	rootCmd.PersistentFlags().StringVar(&util.DbFile, "database", "./twitch.db", "db file where data gets saved/cached to (default is ./twitch.db .  It will get generated if no value is given)")
	rootCmd.PersistentFlags().StringVar(&util.ClientID, "client_id", "", "client_id of the registered Twitch App")
	rootCmd.PersistentFlags().StringVar(&util.ClientSecret, "client_secret", "", "client_secret of the registered Twitch App")
	rootCmd.PersistentFlags().StringVar(&util.BotAToken, "bot_accessToken", "", "accessToken of the Twitch Bot User. You can acquire this by opening https://twitchapps.com/tmi and removing \"oauth:\" at the front")
	rootCmd.PersistentFlags().StringVar(&util.BotUName, "bot_username", "", "username of the Twitch Bot User.")
	rootCmd.PersistentFlags().StringVar(&util.Publicaddress, "public_address", "", "Address of the Public Listening HTTP Server (used for the Twitch Callback)")
	rootCmd.PersistentFlags().StringVar(&util.TLSCert, "tls_cert", "", "Path to TLS Cert File.")
	rootCmd.PersistentFlags().StringVar(&util.TLSKey, "tls_key", "", "Path to TLS Key File.")
}

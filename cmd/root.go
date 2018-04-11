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
	"github.com/spf13/cobra"
)

var cfgFile string
var dbFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "matrix-twitch-bridge",
	//TODO Descriptions
	Short: "",
	Long:  ``,
	PreRun: func(cmd *cobra.Command, args []string) {
		db.Init(dbFile)
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if cfgFile == "" {
			asLogic.Init()
		} else {
			asLogic.Run(cfgFile)
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config.yaml", "config file (default is ./config.yaml . It will get generated if no value is given)")
	rootCmd.PersistentFlags().StringVar(&dbFile, "db", "./twitch.db", "db file where data gets saved/cached to (default is ./twitch.db .  It will get generated if no value is given)")
}

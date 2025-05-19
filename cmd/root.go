// Copyright 2025 Ryan White
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
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var outputFormat string
var cpuProfile string
var memProfile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mot",
	Short: "A qBittorrent CLI",
	Long: `
 ▄▀▀▄ ▄▀▄  ▄▀▀▀▀▄   ▄▀▀▀█▀▀▄ 
█  █ ▀  █ █      █ █    █  ▐ 
▐  █    █ █      █ ▐   █     
  █    █  ▀▄    ▄▀    █      
▄▀   ▄▀     ▀▀▀▀    ▄▀       
█    █             █         
▐    ▐             ▐         

mot: A CLI for interacting with qBittorrent.
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug, _ := cmd.Flags().GetBool("debug"); !debug {
			return
		}
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			<-c
			f.Close()
			os.Exit(0)
		}()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if debug, _ := cmd.Flags().GetBool("debug"); !debug {
			return
		}

		// Stop CPU profiling
		pprof.StopCPUProfile()

		f, err := os.Create(memProfile)
		if err != nil {
			log.Fatal("could not create memory profile:", err)
		}

		defer f.Close()
		runtime.GC()

		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile:", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringP("username", "u", "", "qBittorrent username")
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))

	rootCmd.PersistentFlags().StringP("password", "p", "", "qBittorrent password")
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))

	rootCmd.PersistentFlags().String("url", "", "qBittorrent url")
	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))

	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "output format")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mot.yaml)")

	rootCmd.PersistentFlags().StringVar(&cpuProfile, "profile-cpu", "cpu.prof", "write cpu profiles to file")
	rootCmd.PersistentFlags().MarkHidden("profile-cpu")
	rootCmd.PersistentFlags().StringVar(&memProfile, "profile-mem", "mem.prof", "write memory profiles to file")
	rootCmd.PersistentFlags().MarkHidden("profile-mem")
	rootCmd.PersistentFlags().Bool("debug", false, "enable debugging for mot")

	rootCmd.AddGroup(
		&cobra.Group{ID: "basic", Title: "Basic Commands:"},
		&cobra.Group{ID: "torrent", Title: "Torrent Management Commands:"},
		&cobra.Group{ID: "other", Title: "Other Commands:"},
	)
	rootCmd.SetCompletionCommandGroupID("other")
	rootCmd.SetHelpCommandGroupID("other")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else if f, ok := os.LookupEnv("MOT_CONFIG"); ok {
		viper.SetConfigFile(f)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".mot" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".mot")
	}

	viper.SetEnvPrefix("mot")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// TODO: only print when debugging/increased verbosity option is set
		// fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		_ = err
	}
}

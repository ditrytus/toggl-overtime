/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/now"
	"github.com/spf13/cobra"

	linq "github.com/ahmetb/go-linq/v3"
	toggl "github.com/jason0x43/go-toggl"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile         string
	token           string
	workspace       string
	projects        []string
	workDayDuration time.Duration
	nonWorkingDays  []string
	startDate       string
	endDate         string
)

var rootCmd = &cobra.Command{
	Use:     "toggl-overtime",
	Short:   "Calculates overtime based on time logged with Toggl app.",
	Long:    `Calculates overtime based on time logged with Toggl app.`,
	Version: "0.1",
	RunE: func(cmd *cobra.Command, args []string) error {
		// fmt.Printf("config: %s\n", cfgFile)
		// fmt.Printf("token: %s\n", token)
		// fmt.Printf("workspace: %s\n", workspace)
		// fmt.Printf("projects: %v\n", projects)
		// fmt.Printf("work-day-duration: %v\n", workDayDuration)
		// fmt.Printf("non-working-days: %v\n", nonWorkingDays)
		// fmt.Printf("start-date: %s\n", startDate)
		// fmt.Printf("end-date: %s\n", endDate)

		toggl.DisableLog()
		session := toggl.OpenSession(token)

		account, err := session.GetAccount()
		if err != nil {
			return fmt.Errorf("could not get toggl account: %w", err)
		}

		var (
			wid    int
			wfound bool
		)
		for _, w := range account.Data.Workspaces {
			if w.Name == workspace {
				wid = w.ID
				wfound = true
			}
		}
		if !wfound {
			return fmt.Errorf("workspace with name '%s' does not exist", workspace)
		}

		report, err := session.GetSummaryReport(wid, startDate, endDate)
		if err != nil {
			return fmt.Errorf("could not get summary report: %w", err)
		}

		var totalWorkedTime time.Duration
		for _, data := range report.Data {
			if linq.From(projects).Contains(data.Title.Project) {
				totalWorkedTime += time.Millisecond * time.Duration(data.Time)
			}
		}

		from, err := now.Parse(startDate)
		if err != nil {
			return fmt.Errorf("invalid start date: %w", err)
		}

		until, err := now.Parse(endDate)
		if err != nil {
			return fmt.Errorf("invalid end date: %w", err)
		}

		fY, fM, fD := from.Date()
		beginningOfStartDate := time.Date(fY, fM, fD, 0, 0, 0, 0, from.Location())
		uY, uM, uD := until.Date()
		endOfEndDate := time.Date(uY, uM, uD, 23, 59, 50, 0, until.Location())

		var expectedWorkTime time.Duration
		for day := beginningOfStartDate; day.Before(endOfEndDate); day = day.Add(time.Hour * 24) {
			if !linq.From(nonWorkingDays).Contains(day.Weekday().String()) {
				expectedWorkTime += workDayDuration
			}
		}

		overtime := totalWorkedTime - expectedWorkTime
		fmt.Println(overtime)

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(
		&cfgFile, "config", "",
		"config file (default is $HOME/.toggl-overtime.yaml)")
	rootCmd.PersistentFlags().StringVarP(
		&token, "token", "t", "",
		"Toggl API authorization token")
	rootCmd.PersistentFlags().StringVarP(
		&workspace, "workspace", "w", "Personal",
		"Toggl workspace")
	rootCmd.PersistentFlags().StringSliceVarP(
		&projects, "projects", "p", []string{"Work"},
		"Toggl projects")
	rootCmd.PersistentFlags().DurationVarP(
		&workDayDuration, "work-day-duration", "d", time.Hour*8,
		"base duration of a working day (default it 8 hours)")
	rootCmd.PersistentFlags().StringSliceVarP(
		&nonWorkingDays, "non-working-days", "n", []string{"Saturday", "Sunday"},
		"list of days of the week that shouldn't be count as working days (default is Saturday and Sunday)")
	rootCmd.PersistentFlags().StringVarP(
		&startDate, "start-date", "s", now.BeginningOfMonth().String(),
		"start date of period for which to calculate overtime (default is beginning of the current month)")
	rootCmd.PersistentFlags().StringVarP(
		&endDate, "end-date", "e", now.EndOfDay().String(),
		"end date of period for which to calculate overtime (defult is end of today)")

}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".toggl-overtime")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

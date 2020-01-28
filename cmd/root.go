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
	cfgFile           string
	token             string
	workspace         string
	projects          []string
	workDayDuration   time.Duration
	excludeDaysOfWeek []string
	excludeDates      []string
	startDate         string
	endDate           string
)

var rootCmd = &cobra.Command{
	Use:     "toggl-overtime",
	Short:   "Calculates overtime based on time logged with Toggl app.",
	Long:    `Calculates overtime based on time logged with Toggl app.`,
	Version: "0.1",
	RunE: func(cmd *cobra.Command, args []string) error {
		toggl.DisableLog()

		if flagShouldDefaultToConfig(cmd, "token") {
			token = viper.GetString("token")
		}

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

		if flagShouldDefaultToConfig(cmd, "exclude-dates") {
			excludeDates = viper.GetStringSlice("exclude-dates")
		}

		fY, fM, fD := from.Date()
		beginningOfStartDate := time.Date(fY, fM, fD, 0, 0, 0, 0, from.Location())

		uY, uM, uD := until.Date()
		endOfEndDate := time.Date(uY, uM, uD, 23, 59, 59, 0, until.Location())

		var expectedWorkTime time.Duration
		for day := beginningOfStartDate; day.Before(endOfEndDate); day = day.Add(time.Hour * 24) {
			if linq.From(excludeDaysOfWeek).Contains(day.Weekday().String()) {
				continue
			}
			if linq.From(excludeDates).Contains(day.Format("2006-01-02")) {
				continue
			}
			expectedWorkTime += workDayDuration
		}

		overtime := totalWorkedTime - expectedWorkTime

		fmt.Println(overtime)

		return nil
	},
}

func flagShouldDefaultToConfig(cmd *cobra.Command, name string) bool {
	return !cmd.PersistentFlags().Changed(name) && viper.IsSet(name)
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
		"base duration of a working day")
	rootCmd.PersistentFlags().StringSliceVarP(
		&excludeDaysOfWeek, "non-working-days", "n", []string{"Saturday", "Sunday"},
		"list of days of the week that shouldn't be count as working days")
	rootCmd.PersistentFlags().StringSliceVarP(
		&excludeDates, "exclude-dates", "x", []string{""},
		"list of dates that shouldn't be count as working days")
	rootCmd.PersistentFlags().StringVarP(
		&startDate, "start-date", "s", now.BeginningOfMonth().String(),
		"start date of period for which to calculate overtime")
	rootCmd.PersistentFlags().StringVarP(
		&endDate, "end-date", "e", now.EndOfDay().String(),
		"end date of period for which to calculate overtime")

}

// func defaultFromConfig() {
// 	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
// 		if !rootCmd.Flags().Changed(f.Name) && viper.IsSet(f.Name) {
// 			fmt.Println(fmt.Sprintf("%v", viper.Get(f.Name)))
// 			err := f.Value.Set(fmt.Sprintf("%v", viper.Get(f.Name)))
// 			if err != nil {
// 				log.Panicf("could not assign default value %v", err)
// 			}
// 		}
// 	})
// }

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

	viper.ReadInConfig()
}

package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/iamhectorsosa/anchor/internal/database"
	"github.com/iamhectorsosa/anchor/internal/logger"
	"github.com/iamhectorsosa/anchor/internal/store"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "anchor [key] [...$1] | [key='value']",
	Short: "Anchor is a CLI tool for managing your anchors.",
	Long: `Anchor is a CLI tool for managing your anchors.

To get an anchor, use: anchor [key] [...$1]
To add anchors, use: anchor [key='value']`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.New()
		if len(args) >= 1 {
			input := args[0]
			if strings.Contains(input, "=") {
				inputSlice := strings.SplitN(input, "=", 2)
				if len(inputSlice) != 2 {
					return log.Error("invalid format. Use: key='value'")
				}

				key := inputSlice[0]
				value := strings.TrimSpace(strings.Trim(inputSlice[1], "'"))

				db, cleanup, err := database.New()
				if err != nil {
					return log.Error("database.New, err=%v", err)
				}
				defer cleanup()

				if err = db.Create(key, value); err != nil {
					return log.Error("db.Create, err=%v", err)
				}

				log.Info("Anchor successfully created, key=%q value=%q.", key, value)
				return nil
			}

			db, cleanup, err := database.New()
			if err != nil {
				return log.Error("database.New, err=%v", err)
			}
			defer cleanup()

			key := input
			anchor, err := db.Read(key)
			if err != nil {
				return log.Error("db.Read, err=%v", err)
			}

			value := anchor.Value
			placeholderCount := strings.Count(value, "$")

			if placeholderCount != len(args)-1 {
				return log.Error("missing arguments for %s=%q", key, value)
			}

			for i, arg := range args[1:] {
				placeholder := fmt.Sprintf("$%d", i+1)
				value = strings.ReplaceAll(value, placeholder, arg)
			}

			cmd := exec.Command("pbcopy")
			cmd.Stdin = bytes.NewReader([]byte(value))
			if err := cmd.Run(); err != nil {
				return log.Error("pbcopy in cmd.Run, err=%v", err)
			}

			log.Info("Copied to clipboard, value=%q", value)

			cmd = exec.Command("open", value)
			if err := cmd.Run(); err != nil {
				return log.Error("open in cmd.Run, err=%v", err)
			}

			return nil
		} else {
			return cmd.Help()
		}
	},
}

var ls = &cobra.Command{
	Use:   "ls",
	Short: "List all anchors",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.New()
		db, cleanup, err := database.New()
		if err != nil {
			return log.Error("database.New, err=%v", err)
		}
		defer cleanup()

		anchors, err := db.ReadAll()
		if err != nil {
			return log.Error("db.ReadAll, err=%v", err)
		}

		log.Info("Found %d anchors...", len(anchors))

		if len(anchors) == 0 {
			return nil
		}

		maxKeyLen, maxValueLen := 0, 0
		for _, s := range anchors {
			if len(s.Key) > maxKeyLen {
				maxKeyLen = len(s.Key) + 6
			}
			if len(s.Value) > maxValueLen {
				maxValueLen = len(s.Value)
			}
		}

		anchors = append([]store.Anchor{store.Anchor{
			Id:    0,
			Key:   "KEY",
			Value: "VALUE",
		}}, anchors...)

		evenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
		for i, s := range anchors {
			key := fmt.Sprintf("%-*s", maxKeyLen, s.Key)
			value := fmt.Sprintf("%-*s", maxValueLen, s.Value)
			if i%2 == 0 {
				key = evenStyle.Render(key)
				value = evenStyle.Render(value)
			}
			fmt.Println(key, value)
		}

		return nil
	},
}

var update = &cobra.Command{
	Use:   "update [key='new_value']",
	Short: "Update an anchor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.New()
		input := args[0]
		inputSlice := strings.SplitN(input, "=", 2)
		if len(inputSlice) != 2 {
			return log.Error("invalid format. Use: key='new_value'")
		}

		key := inputSlice[0]
		newValue := strings.TrimSpace(strings.Trim(inputSlice[1], "'"))

		db, cleanup, err := database.New()
		if err != nil {
			return log.Error("database.New, err=%v", err)
		}
		defer cleanup()

		anchor, err := db.Read(key)
		if err != nil {
			return log.Error("db.Read, err=%v", err)
		}

		if err = db.Update(store.Anchor{
			Id:    anchor.Id,
			Key:   key,
			Value: newValue,
		}); err != nil {
			return log.Error("db.Update, err=%v", err)
		}

		log.Info("Anchor successfully updated, key=%q value=%q.", key, newValue)
		return nil
	},
}

var delete = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a anchor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.New()
		key := args[0]
		db, cleanup, err := database.New()
		if err != nil {
			return log.Error("database.New, err=%v", err)
		}
		defer cleanup()

		if err = db.Delete(key); err != nil {
			return log.Error("db.Delete, err=%v", err)
		}

		log.Info("Anchor successfully deleted, key=%q.", key)
		return nil
	},
}

var reset = &cobra.Command{
	Use:   "reset",
	Short: "Reset all anchors",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.New()
		db, cleanup, err := database.New()
		if err != nil {
			return log.Error("database.New, err=%v", err)
		}
		defer cleanup()

		if err = db.Reset(); err != nil {
			return log.Error("db.Reset, err=%v", err)
		}

		log.Info("Anchors have been successfully reset")
		return nil
	},
}

var (
	exportPath     string
	importFilePath string
	importUrlPath  string
)

var export = &cobra.Command{
	Use:   "export",
	Short: "Export all anchors",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.New()
		db, cleanup, err := database.New()
		if err != nil {
			return log.Error("database.New, err=%v", err)
		}
		defer cleanup()

		anchors, err := db.ReadAll()
		if err != nil {
			return log.Error("db.ReadAll, err=%v", err)
		}

		log.Info("Generating report with %d anchors...", len(anchors))

		filename := filepath.Join(exportPath, fmt.Sprintf("anchor-%s.csv", time.Now().Format("2006-01-02")))
		filename = filepath.Clean(filename)
		file, err := os.Create(filename)
		if err != nil {
			return log.Error("os.Create, err=%v", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		if err := writer.Write([]string{"Key", "Value"}); err != nil {
			return fmt.Errorf("writer.Write, err=%v", err)
		}

		for _, anchor := range anchors {
			if err := writer.Write([]string{anchor.Key, anchor.Value}); err != nil {
				return fmt.Errorf("writer.Write, err=%v", err)
			}
		}

		log.Info("CSV file successfully created at path=%q", filename)
		return nil
	},
}

var importc = &cobra.Command{
	Use:   "import",
	Short: "Import anchors",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.New()

		if importFilePath == "" && importUrlPath == "" {
			return log.Error("a valid path or url is required, path=%q, url=%q", importFilePath, importUrlPath)
		}

		var reader io.Reader
		if importFilePath != "" {
			importFilePath = filepath.Clean(importFilePath)
			file, err := os.Open(importFilePath)
			if err != nil {
				return log.Error("os.Open, err=%v", err)
			}
			defer file.Close()
			reader = file
		} else {
			req, err := http.NewRequest(http.MethodGet, importUrlPath, nil)
			if err != nil {
				return log.Error("http.NewRequest, err=%v", err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return log.Error("http.Get, err=%v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return log.Error("resp.StatusCode, status=%v", resp.Status)
			}
			reader = resp.Body
		}

		csvReader := csv.NewReader(reader)
		records, err := csvReader.ReadAll()
		if err != nil {
			return log.Error("csvReader.ReadAll, err=%v", err)
		}

		var anchors []store.Anchor
		for _, record := range records[1:] {
			if len(record) < 2 {
				continue
			}
			anchor := store.Anchor{
				Key:   record[0],
				Value: record[1],
			}
			anchors = append(anchors, anchor)
		}

		if len(anchors) == 0 {
			return log.Error("no valid anchors where found")
		}

		db, cleanup, err := database.New()
		if err != nil {
			return log.Error("database.New, err=%v", err)
		}
		defer cleanup()

		if err := db.Import(anchors); err != nil {
			return log.Error("db.Import, err=%v", err)
		}

		source := importFilePath
		if source == "" {
			source = importUrlPath
		}
		log.Info("CSV file successfully imported from %q", source)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(ls)
	rootCmd.AddCommand(update)
	rootCmd.AddCommand(delete)
	rootCmd.AddCommand(reset)
	rootCmd.AddCommand(export)
	rootCmd.AddCommand(importc)
	export.Flags().StringVarP(&exportPath, "path", "p", ".", "Path to directory for CSV output")
	importc.Flags().StringVarP(&importFilePath, "path", "p", "", "Path to directory of your CSV file")
	importc.Flags().StringVarP(&importUrlPath, "url", "u", "", "URL of your remote CSV file")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
}
